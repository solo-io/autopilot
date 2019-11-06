// run this file in order to generate a kubernetes-YAML file for your project's top-level CRD
package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"istio.io/api/networking/v1alpha3"
	v12 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/solo-io/autopilot/codegen/util"

	v1 "github.com/solo-io/autopilot/examples/test/pkg/apis/tests/v1"
)

//go:generate go run create_cr_yaml.go

// TODO: modify this object and re-run the script in order to produce the output YAML file
var ExampleTest = &v1.Test{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: "default",
		Name:      "example",
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "Test",
		APIVersion: "autopiot.example.io/v1",
	},
	Spec: v1.TestSpec{
		Target: v12.ObjectReference{
			Name:      "podinfo-primary",
			Namespace: "test",
		},
		Faults: v1.HTTPFaultInjection{
			Abort: &v1alpha3.HTTPFaultInjection_Abort{
				ErrorType: &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{
					HttpStatus: 555,
				},
				Percentage: &v1alpha3.Percent{
					Value: 90,
				},
			},
		},
		Threshold: 90,
		Timeout:   metav1.Duration{time.Minute},
	},
}

// modify this string to change the output file path
var OutputFile = filepath.Join(util.MustGetThisDir(), "..", "deploy", "example_test.yaml")

func main() {
	yam, err := yaml.Marshal(ExampleTest)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(OutputFile, yam, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
