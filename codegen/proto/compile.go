package proto

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"

	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
)

// make sure the pkg matches the go_package option in the proto
// TODO: validate this
func CompileProtos(goModule, apiRoot, protoDir string) ([]*model.DescriptorWithPath, error) {
	// need to be in module root so protoc runs on the expected apiRoot
	if err := os.Chdir(util.GetModuleRoot()); err != nil {
		return nil, err
	}

	protoDir, err := filepath.Abs(protoDir)
	if err != nil {
		return nil, err
	}
	protoOutDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer os.Remove(protoOutDir)

	descriptors, err := collector.NewCollector(
		nil,
		[]string{protoDir}, // import the inputs dir
		nil,
		[]string{
			"jsonshim",
		},
		protoOutDir,
		func(file string) bool {
			return true
		}).CollectDescriptorsFromRoot(protoDir, nil)
	if err != nil {
		return nil, err
	}

	// copy the files generated for our package into our repo from the
	// tmp dir
	return descriptors, copyFiles(filepath.Join(protoOutDir, goModule, apiRoot), apiRoot)
}

func copyFiles(srcDir, destDir string) error {
	if err := filepath.Walk(srcDir, func(srcFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		destFile := filepath.Join(destDir, strings.TrimPrefix(srcFile, srcDir))
		destFile = strings.TrimPrefix(destFile, "/")

		// copy
		srcReader, err := os.Open(srcFile)
		if err != nil {
			return err
		}
		defer srcReader.Close()

		if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(destFile)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		log.Printf("copying %v -> %v", srcFile, destFile)
		_, err = io.Copy(dstFile, srcReader)
		return err

	}); err != nil {
		return err
	}

	return nil
}
