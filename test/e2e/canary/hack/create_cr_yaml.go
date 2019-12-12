// run this file in order to generate a kubernetes-YAML file for your project's top-level CRD
package main

import (
	"io/ioutil"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/solo-io/autopilot/codegen/util"

	v1 "github.com/solo-io/autopilot/test/e2e/canary/pkg/apis/canarydeployments/v1"
)

//go:generate go run create_cr_yaml.go

// TODO: modify this object and re-run the script in order to produce the output YAML file
var ExampleCanaryDeployment = &v1.CanaryDeployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "example",
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "CanaryDeployment",
		APIVersion: "autopilot.examples.io/v1",
	},
	Spec: v1.CanaryDeploymentSpec{
		// fill me in!
	},
}

// modify this string to change the output file path
var OutputFile = filepath.Join(util.MustGetThisDir(), "..", "deploy", "canarydeployment_example.yaml")

func main() {
	yam, err := yaml.Marshal(ExampleCanaryDeployment)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(OutputFile, yam, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
