package proto

import (
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CompileProtos(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	protoOutDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.Remove(protoOutDir)

	_, err = collector.NewCollector(
		nil,
		[]string{dir},
		nil,
		[]string{
			"jsonshim",
		},
		protoOutDir,
		func(file string) bool {
			return true
		}).CollectDescriptorsFromRoot(dir, nil)
if err != nil {
	return err
}

	// make sure this matches the go_package option in the proto
	pkg := util.GetGoPkg()

	// copy the files generated for our package into our repo from the
	// tmp dir
	err = copyFiles(filepath.Join(protoOutDir, pkg), dir)

	return err
}

func copyFiles(srcDir, destDir string) error {
	if err := filepath.Walk(srcDir, func(srcFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		destFile := strings.TrimPrefix(srcFile, srcDir)
		destFile = strings.TrimPrefix(destFile, "/")

		dir := filepath.Dir(filepath.Join(destDir, destFile))
		if err := os.MkdirAll(dir, 0755); err != nil{
			return err
		}

		// copy
		srcReader, err := os.Open(srcFile)
		if err != nil {
			return err
		}
		defer srcReader.Close()

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