package run

import (
	"context"
	"flag"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	zaputil "sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/pkg/config"
	"github.com/solo-io/autopilot/pkg/defaults"
	"github.com/solo-io/autopilot/pkg/scheduler"
	"github.com/solo-io/autopilot/pkg/utils"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	// load auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	// the global scheme used by the operator
	schemeBuilder = runtime.SchemeBuilder{
		clientgoscheme.AddToScheme,
	}

	// update the DefaultRunOptions at init time to manually override default run options
	DefaultRunOptions = Options{
		Ctx:          context.Background(),
		OperatorFile: defaults.OperatorFile,
	}

	logLevel = zap.NewAtomicLevel()
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

// the main entrypoint for the Autopilot Operator
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

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	// set zap as the global logger
	logf.SetLogger(zaputil.New(func(options *zaputil.Options) {
		options.Level = &logLevel
	}))

	// cancel the root context on Signal
	ctx := contextWithStop(cfg.Ctx, ctrl.SetupSignalHandler())

	cfgs, err := watchOperatorConfigs(ctx, logger, cfg.OperatorFile)
	if err != nil {
		logger.Error(err, "failed starting config watcher, using default config", "config", config.DefaultConfig)
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
func singleConfig(ctx context.Context, operator *v1.AutopilotOperator) <-chan *v1.AutopilotOperator {
	configs := make(chan *v1.AutopilotOperator)
	go func() {
		select {
		case <-ctx.Done():
			return
		case configs <- operator:
		}
	}()
	return configs
}

func watchOperatorConfigs(ctx context.Context, logger logr.Logger, operatorFile string) (<-chan *v1.AutopilotOperator, error) {
	configs := make(chan *v1.AutopilotOperator)

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

		lastConfig := operator

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logger.Info("new Operator config detected!", "file", event.Name)

				// set up filewatcher for the operator config
				operator, err := config.ConfigFromFile(operatorFile)
				if err != nil {
					logger.Error(err, "failed to read operator config file")
					continue
				}

				if proto.Equal(lastConfig, operator) {
					continue
				}

				select {
				case <-ctx.Done():
					return
				case configs <- operator:
					lastConfig = operator
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
	configs <-chan *v1.AutopilotOperator,
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
			operatorCtx, cancel = operatorContext(ctx, operator, logger)

			instance := operatorInstance{
				ctx:          operatorCtx,
				config:       operator,
				scheme:       scheme,
				addTomanager: addTomanager,
				logger:       logger,
			}

			go func() {
				logger.Info("Warning: Flushing Operator Metrics!")

				// metrics must be flushed as the new Controller re-registers metrics with the same name
				metrics.Registry = prometheus.NewRegistry()

				if err := instance.Start(); err != nil {
					logger.Error(err, "failed to start operator instance")
					os.Exit(1)
				}
			}()
		}
	}
}

// the operatorInstance launches instances of the operator on config changes
type operatorInstance struct {
	ctx          context.Context
	config       *v1.AutopilotOperator
	scheme       *runtime.Scheme
	addTomanager AddToManager
	logger       logr.Logger
}

func (instance *operatorInstance) Start() error {

	enableLeaderElection := instance.config.EnableLeaderElection
	leaderElectionNamespace := instance.config.LeaderElectionNamespace

	if enableLeaderElection {
		leaderNs, err := utils.GetInClusterNamesapce()
		if err != nil {
			// override if running out-of-cluster
			instance.logger.Info("Skipping leader-election when running out of cluster")
			enableLeaderElection = false
		} else if leaderElectionNamespace == "" {
			// use currently deployed namespace as default leader ns
			leaderElectionNamespace = leaderNs
		}
	}

	level := 1

	if instance.config.LogLevel != nil {
		level = int(instance.config.LogLevel.Value)
	}

	// zap levels start with -1 (for debug)
	// ours starts with 0 for debug
	logLevel.SetLevel(zapcore.Level(level - 1))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  instance.scheme,
		MetricsBindAddress:      instance.config.MetricsAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: leaderElectionNamespace,
		// TODO: webhook support
	})
	if err != nil {
		return err
	}
	params := scheduler.Params{
		Ctx:       instance.ctx,
		Manager:   mgr,
		Namespace: instance.config.WatchNamespace,
		Logger:    instance.logger,
	}

	if err := instance.addTomanager(params); err != nil {
		return err
	}

	return mgr.Start(instance.ctx.Done())
}

func operatorContext(ctx context.Context, operator *v1.AutopilotOperator, logger logr.Logger) (context.Context, context.CancelFunc) {
	ctx = config.ContextWithConfig(ctx, operator)
	ctx = utils.ContextWithLogger(ctx, logger)
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
