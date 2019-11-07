package deploy

import (
	"github.com/solo-io/autopilot/codegen/model"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func RoleBinding(data *model.ProjectData) runtime.Object {
	return roleBinding(data)
}

func roleBinding(data *model.ProjectData) *v1.RoleBinding {
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

func ClusterRoleBinding(data *model.ProjectData) runtime.Object {
	return clusterRoleBinding(data)
}

func clusterRoleBinding(data *model.ProjectData) *v1.ClusterRoleBinding {
	return &v1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.OperatorName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "ClusterRoleBinding",
		},
		Subjects: []v1.Subject{{
			Kind:      "ServiceAccount",
			Name:      data.OperatorName,
			Namespace: "REPLACE_NAMESPACE",
		}},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     data.OperatorName,
		},
	}
}
