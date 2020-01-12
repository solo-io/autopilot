package main

import (
	"flag"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	generatorargs "k8s.io/code-generator/cmd/deepcopy-gen/args"
	"k8s.io/gengo/examples/deepcopy-gen/generators"
)

//go:generate go run generate.go

func main() {
	log.Printf("generating deecopy code for internal types")
	if err := gen(); err != nil {
		log.Fatal(err)
	}
}

func gen() error {
	_ = flag.Set("logtostderr", "true")
	apiPath := "./apis/..."
	args, cargs := generatorargs.NewDefaults()
	args.InputDirs = []string{apiPath}
	args.OutputPackagePath = apiPath
	args.OutputFileBaseName = "zz_generated.deepcopy"
	cargs.BoundingDirs = []string{apiPath}

	if err := generatorargs.Validate(args); err != nil {
		return errors.Wrap(err, "deepcopy-gen argument validation error")
	}

	log.Printf("Generating Deepcopy code for API: %#v", args)
	err := args.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	)
	if err != nil {
		return errors.Wrap(err, "deepcopy-gen generator error")
	}
	return nil
}
