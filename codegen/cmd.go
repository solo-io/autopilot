package codegen

import (
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/proto"
	"github.com/solo-io/autopilot/codegen/render"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/codegen/writer"
)

// runs the codegen compilation for the current Go module
type Command struct {
	// the name of the app or component
	// used to label k8s manifests
	AppName string

	// search protos recursively starting from this directory
	// TODO: use anyvendor to support external imports
	ProtoDir string

	// the k8s api groups for which to compile
	Groups []render.Group

	// the root directory for generated API code
	ApiRoot string

	// the root directory for generated Kube manfiests
	ManifestRoot string

	// the go module of the project
	// set by Execute()
	goModule string
}

// function to execute Autopilot code gen from another repository
func (c Command) Execute() error {
	c.goModule = util.GetGoModule()
	for _, group := range c.Groups {
		// init connects children to their parents
		group.Init()

		if err := c.writeGeneratedFiles(group); err != nil {
			return err
		}
	}
	return nil
}

func (c Command) writeGeneratedFiles(grp model.Group) error {
	if err := c.compileProtos(grp); err != nil {
		return err
	}

	apiWriter := &writer.DefaultWriter{Root: c.ApiRoot}

	apiTypes, err := render.RenderApiTypes(c.goModule, c.ApiRoot, grp)
	if err != nil {
		return err
	}

	if err := apiWriter.WriteFiles(apiTypes); err != nil {
		return err
	}

	manifestWriter := &writer.DefaultWriter{Root: c.ManifestRoot}

	manifests, err := render.RenderManifests(c.AppName, c.ManifestRoot, grp)
	if err != nil {
		return err
	}

	if err := manifestWriter.WriteFiles(manifests); err != nil {
		return err
	}

	if err := render.KubeCodegen(c.ApiRoot, grp); err != nil {
		return err
	}

	return nil
}

func (c Command) compileProtos(grp render.Group) error {
	if !grp.RenderProtos {
		return nil
	}

	if err := proto.CompileProtos(
		c.goModule,
		c.ApiRoot,
		c.ProtoDir,
	); err != nil {
		return err
	}

	return nil
}
