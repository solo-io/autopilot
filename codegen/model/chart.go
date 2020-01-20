package model

import (
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type Chart struct {
	Operators []Operator

	Values interface{}

	// goes into the chart.yaml
	Data Data
}

type Operator struct {
	Name string

	// deployment config
	Deployment Deployment
	// these populate the generated ClusterRole for the operator
	Rbac []rbacv1.PolicyRule
	// these populate the generated ClusterRole for the operator
	Volumes []v1.Volume
	// mount these volumes to the operator container
	VolumeMounts []v1.VolumeMount
	// add a manifest for each configmap
	ConfigMaps []v1.ConfigMap

	// args for the container
	Args []string
}

// values for Deployment template
type Deployment struct {
	// use a DaemonSet instead of a Deployment
	UseDaemonSet bool                     `json:"-"`
	Image        Image                    `json:"image,omitempty"`
	Resources    *v1.ResourceRequirements `json:"resources,omitempty"`
}

type Image struct {
	Tag        string        `json:"tag,omitempty"  desc:"tag for the container"`
	Repository string        `json:"repository,omitempty"  desc:"image name (repository) for the container."`
	Registry   string        `json:"registry,omitempty" desc:"image prefix/registry e.g. (quay.io/solo-io)"`
	PullPolicy v1.PullPolicy `json:"pullPolicy,omitempty"  desc:"image pull policy for the container"`
	PullSecret string        `json:"pullSecret,omitempty" desc:"image pull policy for the container "`
}

type Data struct {
	ApiVersion  string   `json:"apiVersion,omitempty"`
	Description string   `json:"description,omitempty"`
	Name        string   `json:"name,omitempty"`
	Version     string   `json:"version,omitempty"`
	Home        string   `json:"home,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Sources     []string `json:"sources,omitempty"`
}
