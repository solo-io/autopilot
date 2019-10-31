package deploy

import (
	"github.com/solo-io/autopilot/codegen/model"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func RoleBinding(data *model.TemplateData) runtime.Object {
	return roleBinding(data)
}

func roleBinding(data *model.TemplateData) *v1.RoleBinding {
	return &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.OperatorName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "RoleBinding",
		},
		Subjects: []v1.Subject{{
			Kind: "ServiceAccount",
			Name: data.OperatorName,
		}},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     data.OperatorName,
		},
	}
}
