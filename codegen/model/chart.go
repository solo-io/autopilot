package model

import v1 "k8s.io/api/core/v1"

type Chart struct {
	Operators []Operator

	Values interface{}

	// goes into the chart.yaml
	Data Data
}

type Operator struct {
	Name       string
	Deployment Deployment
	Args       []string
}

// values for Deployment template
type Deployment struct {
	Image          Image                    `json:"image,omitempty"`
	FloatingUserId bool                     `json:"floatingUserId" desc:"set to true to allow the cluster to dynamically assign a user ID"`
	RunAsUser      float64                  `json:"runAsUser" desc:"Explicitly set the user ID for the container to run as. Default is 10101"`
	Resources      *v1.ResourceRequirements `json:"resources,omitempty"`
}

type Image struct {
	Tag        string `json:"tag,omitempty"  desc:"tag for the container"`
	Repository string `json:"repository,omitempty"  desc:"image name (repository) for the container."`
	Registry   string `json:"registry,omitempty" desc:"image prefix/registry e.g. (quay.io/solo-io)"`
	PullPolicy string `json:"pullPolicy,omitempty"  desc:"image pull policy for the container"`
	PullSecret string `json:"pullSecret,omitempty" desc:"image pull policy for the container "`
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
