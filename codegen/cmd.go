package codegen

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/anyvendor/pkg/manager"
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/proto"
	"github.com/solo-io/autopilot/codegen/render"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/codegen/writer"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
)

// runs the codegen compilation for the current Go module
type Command struct {
	// the name of the app or component
	// used to label k8s manifests
	AppName string

	// config to vendor protos and other non-go files
	// Optional: If nil will not be used
	AnyVendorConfig *sk_anyvendor.Imports

	// the k8s api groups for which to compile
	Groups []render.Group

	// optinal helm chart to render
	Chart *model.Chart

	// optional Go/Docker images to build
	Builds []model.Build

	// the root directory for generated Kube manfiests
	ManifestRoot string

	// the root directory for Build files (Dockerfile, entrypoint script, etc.)
	BuildRoot string

	// the path to the root dir of the module on disk
	// files will be written relative to this dir,
	// except kube clientsets which
	// will generate to the module of the group
	moduleRoot string

	// the name of the go module (as a go package)
	moduleName string

	// context of the command
	ctx context.Context
}

// function to execute Autopilot code gen from another repository
func (c Command) Execute() error {
	c.ctx = context.Background()
	c.moduleRoot = util.GetModuleRoot()
	c.moduleName = util.GetGoModule()

	if err := c.generateChart(); err != nil {
		return err
	}

	for _, group := range c.Groups {
		// init connects children to their parents
		group.Init()

		if err := c.generateGroup(group); err != nil {
			return err
		}
	}

	for _, build := range c.Builds {
		if err := c.generateBuild(build); err != nil {
			return err
		}
		if err := c.buildPushImage(build); err != nil {
			return err
		}
	}
	return nil
}

func (c Command) generateChart() error {
	if c.Chart != nil {
		files, err := render.RenderChart(*c.Chart)
		if err != nil {
			return err
		}

		writer := &writer.DefaultFileWriter{Root: filepath.Join(c.moduleRoot, c.ManifestRoot, c.AppName)}

		if err := writer.WriteFiles(files); err != nil {
			return err
		}
	}

	return nil
}

func (c Command) generateGroup(grp model.Group) error {
	if err := c.compileProtos(&grp); err != nil {
		return err
	}

	writer := &writer.DefaultFileWriter{Root: c.moduleRoot}

	apiTypes, err := render.RenderApiTypes(grp)
	if err != nil {
		return err
	}

	if err := writer.WriteFiles(apiTypes); err != nil {
		return err
	}

	manifests, err := render.RenderManifests(c.AppName, filepath.Join(c.ManifestRoot, c.AppName), grp)
	if err != nil {
		return err
	}

	if err := writer.WriteFiles(manifests); err != nil {
		return err
	}

	if err := render.KubeCodegen(grp); err != nil {
		return err
	}

	return nil
}

func (c Command) compileProtos(grp *render.Group) error {
	if !grp.RenderProtos {
		return nil
	}

	if grp.ProtoDir == "" {
		grp.ProtoDir = anyvendor.DefaultDepDir
	}

	if c.AnyVendorConfig != nil {
		mgr, err := manager.NewManager(c.ctx, c.moduleRoot)
		if err != nil {
			return err
		}

		if err := mgr.Ensure(c.ctx, c.AnyVendorConfig.ToAnyvendorConfig()); err != nil {
			return err
		}
	}
	descriptors, err := proto.CompileProtos(
		grp.Module,
		grp.ApiRoot,
		grp.ProtoDir,
	)
	if err != nil {
		return err
	}

	// set the descriptors on the group for compilation
	grp.Descriptors = descriptors

	return nil
}

func (c Command) generateBuild(build model.Build) error {
	buildFiles, err := render.RenderBuild(build)
	if err != nil {
		return err
	}

	writer := &writer.DefaultFileWriter{Root: c.BuildRoot}

	if err := writer.WriteFiles(buildFiles); err != nil {
		return err
	}

	return nil
}

func (c Command) buildPushImage(build model.Build) error {
	ldFlags := fmt.Sprintf("-X %v/pkg/version.Version=%v", c.moduleRoot, build.Tag)

	// get the main package from the main directory
	// assumes package == module name + main dir path
	mainkPkg := filepath.Join(c.moduleName, filepath.Dir(build.MainFile))

	buildDir := filepath.Join(c.BuildRoot, build.Repository)

	binName := filepath.Join(buildDir, build.Repository + "-linux-amd64")

	log.Printf("Building main package at %v ...", mainkPkg)

	err := util.GoBuild(util.GoCmdOptions{
		BinName: binName,
		Args: []string{
			"-ldflags=" + ldFlags,
			`-gcflags='all="-N -l"''`,
		},
		PackagePath: mainkPkg,
		Env: []string{
			"GO111MODULE=on",
			"CGO_ENABLED=0",
			"GOARCH=amd64",
			"GOOS=linux",
		},
	})
	if err != nil {
		return err
	}

	defer os.Remove(binName)

	fullImageName := fmt.Sprintf("%v/%v:%v", build.Registry, build.Repository, build.Tag)

	log.Printf("Building docker image %v ...", fullImageName)
	if err := dockerCommand("build", "-t", fullImageName, buildDir); err != nil {
		return err
	}

	if !build.Push {
		return nil
	}

	log.Printf("Pushing docker image %v ...", fullImageName)

	return dockerCommand("push", fullImageName)
}

func dockerCommand(args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = os.Environ()
	return cmd.Run()
}
