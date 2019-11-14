package deploy

import (
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/pkg/defaults"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

func SingleNamespaceOperator(data *model.ProjectData) runtime.Object {
	return deployment(data, false)
}

func AllNamespacesOperator(data *model.ProjectData) runtime.Object {
	return deployment(data, true)
}

func deployment(data *model.ProjectData, clusterScoped bool) *appsv1.Deployment {
	labels := map[string]string{"name": data.OperatorName}

	watchNamespaceEnv := v1.EnvVar{
		Name: defaults.WatchNamespaceEnvVar,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		},
	}
	if clusterScoped {
		watchNamespaceEnv.ValueFrom = nil
		watchNamespaceEnv.Value = metav1.NamespaceAll // watch all namespaces
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.OperatorName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   data.OperatorName,
					Labels: labels,
				},
				Spec: v1.PodSpec{
					ServiceAccountName: data.OperatorName,
					Containers: []v1.Container{{
						Name:            data.OperatorName,
						Image:           "REPLACE_IMAGE",
						Command:         []string{data.OperatorName},
						ImagePullPolicy: v1.PullAlways,
						Env: []v1.EnvVar{
							watchNamespaceEnv,
							{
								Name: defaults.PodNameEnvVar,
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "metadata.name",
									},
								},
							},
							{
								Name:  defaults.OperatorNameEnvVar,
								Value: data.OperatorName,
							},
						},
						WorkingDir: "/config",
						VolumeMounts: []v1.VolumeMount{{
							Name:      data.OperatorName,
							ReadOnly:  true,
							MountPath: "/config",
						}},
					}},
					Volumes: []v1.Volume{{
						Name: data.OperatorName,
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{Name: data.OperatorName},
							},
						},
					}},
				},
			},
		},
	}
}
