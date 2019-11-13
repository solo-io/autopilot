// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	generatorargs "k8s.io/code-generator/cmd/deepcopy-gen/args"
	"k8s.io/gengo/examples/deepcopy-gen/generators"
)

func DeepcopyGen(api string) error {
	_ = flag.Set("logtostderr", "true")
	api = filepath.FromSlash(api)
	// Use relative API path so the generator writes to the correct path.
	apiPath := "." + string(filepath.Separator) + api[strings.Index(api, "pkg/apis"):]
	args, cargs := generatorargs.NewDefaults()
	args.InputDirs = []string{apiPath}
	args.OutputPackagePath = apiPath
	args.OutputFileBaseName = "zz_generated.deepcopy"
	cargs.BoundingDirs = []string{apiPath}

	if err := generatorargs.Validate(args); err != nil {
		return errors.Wrap(err, "deepcopy-gen argument validation error")
	}

	inGopathSrc, err := WdInGoPathSrc()
	if err != nil {
		return err
	}
	if !inGopathSrc {
		args.OutputBase = ""
	}

	log.Printf("Generating Deepcopy code for API: %#v", args)
	err = args.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	)
	if err != nil {
		return errors.Wrap(err, "deepcopy-gen generator error")
	}
	return nil
}
