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
	"github.com/solo-io/autopilot/pkg/scheduler"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
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

// Function to wire the scheduler into the Manager
type AddToManager func(params scheduler.Params) error

// the main entrypoint for the AutoPilot Operator
func Run(addToManager AddToManager) error {
	logger := logf.Log

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

	cfgs, err := watchOperatorConfigs(ctx, logger, cfg.OperatorFile)
	if err != nil {
		logger.Error(err, "failed starting config watcher, using default config: %#v", config.DefaultConfig)
		cfgs = singleConfig(ctx, &config.DefaultConfig)
	}

	go runOperatorOnConfigChange(
		ctx,
		cfgs,
		logger,
		scheme,
		addToManager)

	<-ctx.Done()
	logger.Info("Gracefully shut down...")
	return nil
}

// a channel that only ever sends a single config
func singleConfig(ctx context.Context, operator *v1.AutoPilotOperator) <-chan *v1.AutoPilotOperator {
	configs := make(chan *v1.AutoPilotOperator)
	go func() {
		select {
		case <-ctx.Done():
			return
		case configs <- operator:
		}
	}()
	return configs
}

func watchOperatorConfigs(ctx context.Context, logger logr.Logger, operatorFile string) (<-chan *v1.AutoPilotOperator, error) {
	configs := make(chan *v1.AutoPilotOperator)

	// set up the config watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to start file watcher")
	}

	if err := watcher.Add(operatorFile); err != nil {
		return nil, errors.Wrapf(err, "starting file watch for %v", operatorFile)
	}

	// get initial read on cfg to send
	operator, err := config.ConfigFromFile(operatorFile)
	if err != nil {
		return nil, err
	}

	go func() {
		defer watcher.Close()

		// send initial config
		select {
		case <-ctx.Done():
			return
		case configs <- operator:
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logger.Info("new Operator config detected!", "event", event)

				// set up filewatcher for the operator config
				operator, err := config.ConfigFromFile(operatorFile)
				if err != nil {
					logger.Error(err, "failed to read operator config file")
					continue
				}

				select {
				case <-ctx.Done():
					return
				case configs <- operator:
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error(err, "file watcher encountered error")
			}
		}
	}()

	return configs, nil
}

func runOperatorOnConfigChange(
	ctx context.Context,
	configs <-chan *v1.AutoPilotOperator,
	logger logr.Logger,
	scheme *runtime.Scheme,
	addTomanager AddToManager) {

	var operatorCtx context.Context
	var cancel context.CancelFunc = func() {}
	for {
		select {
		case <-ctx.Done():
			return
		case operator, ok := <-configs:
			if !ok {
				return
			}
			logger.Info("Starting Operator with config", "config", operator)

			cancel()

			// initialize a new context for the operator
			operatorCtx, cancel = operatorContext(ctx, operator)

			instance := operatorInstance{
				ctx:          operatorCtx,
				config:       operator,
				scheme:       scheme,
				addTomanager: addTomanager,
				logger:       logger,
			}

			if err := instance.Start(); err != nil {
				logger.Error(err, "failed to read operator config file")
				continue
			}
		}
	}
}

// the operatorInstance launches instances of the operator on config changes
type operatorInstance struct {
	ctx          context.Context
	config       *v1.AutoPilotOperator
	scheme       *runtime.Scheme
	addTomanager AddToManager
	logger       logr.Logger
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
	params := scheduler.Params{
		Ctx:       r.ctx,
		Manager:   mgr,
		Namespace: r.config.WatchNamespace,
		Logger:    r.logger,
	}

	if err := r.addTomanager(params); err != nil {
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
