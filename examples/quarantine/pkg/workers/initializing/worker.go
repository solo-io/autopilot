package initializing

import (
	"context"
	"github.com/solo-io/autopilot/pkg/aliases"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/autopilot/pkg/utils"

	v1 "github.com/solo-io/autopilot/examples/quarantine/pkg/apis/quarantines/v1"
)

type Worker struct {
	Kube utils.EzKube
}

func (w *Worker) Sync(ctx context.Context, quarantine *v1.Quarantine) (Outputs, v1.QuarantinePhase, error) {
	return Outputs{
		Services: []*aliases.Service{{
			ObjectMeta: metav1.ObjectMeta{Namespace: quarantine.Namespace, Name: "i-made-ths"},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Name:     "http",
					Port:     1987,
					Protocol: corev1.ProtocolTCP,
				}},
				Selector: map[string]string{"app": "quarantine"},
			},
		}},
	}, v1.QuarantinePhaseProcessing, nil
}
