package main

import (
	"github.com/golang/protobuf/ptypes/wrappers"
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/util"
	"log"
	"os"
)

func AutopilotGen() error {
	if err := os.Chdir(util.MustGetThisDir()); err != nil {
		return err
	}
	project := v1.AutopilotProject{
		OperatorName: "hello-operator",
		Resources: []*v1.Resource{{
			Kind:    "Hello",
			Group:   "examples",
			Version: "v1",
			Phases: []*v1.Phase{
				{
					Name:        "Initial",
					Description: "Initial phase of Hello",
					Initial:     true,
					Final:       false,
					Inputs: []*v1.Input{{
						InputType: &v1.Input_Resource{
							Resource: &v1.ResourceParameter{
								Kind:    "Deployment",
								Group:   "apps",
								Version: "v1",
								List:    false,
							},
						},
					}},
					Outputs: []*v1.Output{{
						OutputType: &v1.Output_Resource{
							Resource: &v1.ResourceParameter{
								Kind:    "Service",
								Group:   "",
								Version: "v1",
								List:    false,
							},
						},
					}},
				},
			},
			EnableController: &wrappers.BoolValue{
				Value: true,
			},
			EnableFinalizer: true,
		}},
	}

	operator := v1.AutopilotOperator{
		Version:                 "",
		MeshProvider:            0,
		ControlPlaneNs:          "",
		WorkInterval:            nil,
		MetricsAddr:             "",
		EnableLeaderElection:    false,
		WatchNamespace:          "",
		LeaderElectionNamespace: "",
		LogLevel:                nil,
		XXX_NoUnkeyedLiteral:    struct{}{},
		XXX_unrecognized:        nil,
		XXX_sizecache:           0,
	}

	return codegen.GenerateProject(project, operator, false, false)
}

func main() {
	if err := AutopilotGen(); err != nil {
		log.Fatal(err)
	}
}
