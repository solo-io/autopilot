package run

import (
	"context"
	"flag"
	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/pkg/errors"
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/pkg/config"
	"github.com/solo-io/autopilot/pkg/defaults"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	// the global scheme used by the operator
	schemeBuilder = runtime.SchemeBuilder{}

	// update the DefaultRunOptions at init time to manually override default run options
	DefaultRunOptions = Options{
		Ctx:          context.Background(),
		OperatorFile: defaults.OperatorFile,
	}
)

// init functions should register their types with this scheme
func RegisterAddToScheme(s func(scheme *runtime.Scheme) error) {
	schemeBuilder = append(schemeBuilder, s)
}

// Bootstrap config for the Run function
type Options struct {
	// root context for the operator. cancel this to shutdown gracefully
	Ctx context.Context

	// path to the operator config file
	OperatorFile string
}

// Function to wire controllers into the Manager
type AddToManager func(ctx context.Context, mgr ctrl.Manager, namespace string) error

// the main entrypoint for the AutoPilot Operator
func Run(addToManager AddToManager) error {

	// initialize scheme
	scheme := runtime.NewScheme()

	if err := schemeBuilder.AddToScheme(scheme); err != nil {
		return err
	}

	cfg := DefaultRunOptions

	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	// cancel the root context on Signal
	ctx := contextWithStop(cfg.Ctx, ctrl.SetupSignalHandler())

	// set up the config watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrapf(err, "failed to start file watcher")
	}

	if err := watcher.Add(cfg.OperatorFile); err != nil {
		return errors.Wrapf(err, "starting file watch for %v", cfg.OperatorFile)
	}

	defer watcher.Close()

	logger := logf.Log

	go runOperatorOnConfigChange(
		ctx,
		watcher,
		logger,
		scheme,
		addToManager,
		cfg.OperatorFile)

	<-ctx.Done()
	logger.Info("Gracefully shut down...")
	return nil
}

func runOperatorOnConfigChange(
	ctx context.Context,
	watcher *fsnotify.Watcher,
	logger logr.Logger,
	scheme *runtime.Scheme,
	addTomanager AddToManager,
	operatorFile string) {

	var operatorCtx context.Context
	var cancel context.CancelFunc = func() {}
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			logger.Info("Operator File change detected! restarting...", "event", event)

			cancel()
			// set up filewatcher for the operator config
			operator, err := config.ConfigFromFile(operatorFile)
			if err != nil {
				logger.Error(err, "failed to read operator config file")
				continue
			}

			// initialize a new context for the operator
			operatorCtx, cancel = operatorContext(ctx, operator)

			instance := operatorInstance{
				ctx:          operatorCtx,
				config:       operator,
				scheme:       scheme,
				addTomanager: addTomanager,
			}

			if err := instance.Start(); err != nil {
				logger.Error(err, "failed to read operator config file")
				continue
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

// the operatorInstance launches instances of the operator on config changes
type operatorInstance struct {
	ctx          context.Context
	config       *v1.AutoPilotOperator
	scheme       *runtime.Scheme
	addTomanager AddToManager
}

func (r *operatorInstance) Start() error {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             r.scheme,
		MetricsBindAddress: r.config.MetricsAddr,
		LeaderElection:     r.config.EnableLeaderElection,
		// TODO: webhook support
	})
	if err != nil {
		return err
	}
	if err := r.addTomanager(r.ctx, mgr, r.config.WatchNamespace); err != nil {
		return err
	}

	return mgr.Start(r.ctx.Done())
}

func operatorContext(ctx context.Context, operator *v1.AutoPilotOperator) (context.Context, context.CancelFunc) {
	ctx = config.ContextWithConfig(ctx, operator)
	return context.WithCancel(ctx)
}

func contextWithStop(ctx context.Context, stop <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-stop
		cancel()
	}()
	return ctx
}
