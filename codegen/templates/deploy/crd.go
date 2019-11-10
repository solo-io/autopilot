package deploy

import (
	"fmt"

	"github.com/solo-io/autopilot/codegen/model"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CustomResourceDefinition(data *model.ProjectData) runtime.Object {
	return customResourceDefinition(data)
}

func customResourceDefinition(data *model.ProjectData) *apiextv1beta1.CustomResourceDefinition {
	crd := &apiextv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextv1beta1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", data.KindLowerPlural, data.Group),
		},
		Spec: apiextv1beta1.CustomResourceDefinitionSpec{
			Group: data.Group,
			Scope: apiextv1beta1.NamespaceScoped,
			Versions: []apiextv1beta1.CustomResourceDefinitionVersion{
				{Name: data.Version, Served: true, Storage: true},
			},
			Subresources: &apiextv1beta1.CustomResourceSubresources{
				Status: &apiextv1beta1.CustomResourceSubresourceStatus{},
			},
			Names: apiextv1beta1.CustomResourceDefinitionNames{
				Plural:   data.KindLowerPlural,
				Singular: data.KindLower,
				Kind:     data.Kind,
				ListKind: data.Kind + "List",
			},
		},
	}
	return crd
}
