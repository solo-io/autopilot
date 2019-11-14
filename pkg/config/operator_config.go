package config

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/pkg/defaults"
)

// the default config represents a boilerplate config wired to be run with istio (installed to istio-system)
// these values will be overridden at boot by the user's `config.yaml` if it exists
var DefaultConfig = v1.AutopilotOperator{
	Version: "0.0.1",

	MeshProvider: v1.MeshProvider_Istio,

	ControlPlaneNs: defaults.IstioNamespace,

	WorkInterval: ptypes.DurationProto(time.Second * 5),

	MetricsAddr: ":9091",

	EnableLeaderElection: true,

	WatchNamespace: os.Getenv(defaults.WatchNamespaceEnvVar),

	LogLevel: &wrappers.UInt32Value{Value: 1},
}

// GetConfig attempts to read the autopilot-operator.yaml config file
// If no file is specified, it looks in the default location (project root/autopilot-operator.yaml)
func ConfigFromFile(file string) (*v1.AutopilotOperator, error) {
	if file == "" {
		file = defaults.AutopilotFile
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var cfg v1.AutopilotOperator
	if err := util.UnmarshalYaml(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

var ContextKey = &v1.AutopilotOperator{}

func ConfigFromContext(ctx context.Context) *v1.AutopilotOperator {
	op := ctx.Value(ContextKey)
	if op == nil {
		return &DefaultConfig
	}
	operator, ok := op.(*v1.AutopilotOperator)
	if !ok {
		return &DefaultConfig
	}
	return operator
}

func ContextWithConfig(ctx context.Context, operator *v1.AutopilotOperator) context.Context {
	return context.WithValue(ctx, ContextKey, operator)
}
