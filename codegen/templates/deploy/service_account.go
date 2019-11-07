package deploy

import (
	"github.com/solo-io/autopilot/codegen/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func ServiceAccount(data *model.ProjectData) runtime.Object {
	return serviceAccount(data)
}

func serviceAccount(data *model.ProjectData) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.OperatorName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
	}
}
