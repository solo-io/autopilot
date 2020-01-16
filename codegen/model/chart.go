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
	Image     Image                    `json:"image,omitempty"`
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type Image struct {
	Tag        string        `json:"tag,omitempty"  desc:"tag for the container"`
	Repository string        `json:"repository,omitempty"  desc:"image name (repository) for the container."`
	Registry   string        `json:"registry,omitempty" desc:"image prefix/registry e.g. (quay.io/solo-io)"`
	PullPolicy v1.PullPolicy `json:"pullPolicy,omitempty"  desc:"image pull policy for the container"`
	PullSecret string        `json:"pullSecret,omitempty" desc:"image pull policy for the container "`

	// options for building the image
	Build *BuildOptions `json:"-"`
}

type BuildOptions struct {
	// path to the main.go file
	MainFile string

	// push image after  build
	Push bool
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
