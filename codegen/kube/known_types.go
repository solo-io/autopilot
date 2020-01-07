package kube

import (
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/codegen/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"log"
	"sort"
)

var schemes = runtime.SchemeBuilder{
	k8sscheme.AddToScheme,
}

var KnownK8sTypes = func() []schema.GroupVersionKind {
	gvks, err := getK8sTypes()
	if err != nil {
		log.Fatalf("failed to get k8s builtin types: %v", err)
	}
	return gvks
}()

// gets all built-in GVKs for K8s
func getK8sTypes() ([]schema.GroupVersionKind, error) {
	scheme := runtime.NewScheme()
	if err := schemes.AddToScheme(scheme); err != nil {
		return nil, err
	}
	var allKnownTypes []schema.GroupVersionKind
	for gvk := range scheme.AllKnownTypes() {
		allKnownTypes = append(allKnownTypes, gvk)
	}
	sort.SliceStable(allKnownTypes, func(i, j int) bool {
		return allKnownTypes[i].String() < allKnownTypes[j].String()
	})
	return allKnownTypes, nil
}

func IsKnownType(project *v1.AutopilotProject, parameter model.Parameter) bool {
	for _, gvk := range KnownK8sTypes {
		if gvk.Version == parameter.Version() &&
			gvk.Group == parameter.Group &&
			gvk.Kind == parameter.Kind {
			return true
		}
	}
	for _, res := range project.Resources {
		if res.Version == parameter.Version() &&
			res.Group == parameter.Group &&
			res.Kind == parameter.Kind {
			return true
		}
	}
	return false
}
