package testpkg

import (
	"context"
	"fmt"
	v1 "github.com/solo-io/autopilot/examples/test/pkg/apis/tests/v1"
	"github.com/solo-io/autopilot/pkg/aliases"
	"istio.io/api/networking/v1alpha3"
)

func Initialize(ctx context.Context, test *v1.Test) (aliases.VirtualServices, v1.TestPhase, error) {
	return aliases.VirtualServices{{
		ObjectMeta: test.ObjectMeta,
		Spec: v1alpha3.VirtualService{
			Hosts: []string{
				fmt.Sprintf("%v.%v", test.Spec.Target.Name, test.Spec.Target.Namespace),
			},
			Http: []*v1alpha3.HTTPRoute{test.Spec.Vs.Spec.Http[0]},
		},
	}}, v1.TestPhaseProcessing, nil
}
