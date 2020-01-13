package codegen

import (
	"context"

	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/anyvendor/pkg/manager"
	"github.com/solo-io/autopilot/codegen/ap_anyvendor"
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

	// config to vendor protos and other non-go files
	// Optional: If nil will not be used
	AnyVendorConfig *ap_anyvendor.Imports

	// the k8s api groups for which to compile
	Groups []render.Group

	// the root directory for generated Kube manfiests
	ManifestRoot string

	// the path to the root dir of the module on disk
	// files will be written relative to this dir,
	// except kube clientsets which
	// will generate to the module of the group
	moduleRoot string

	// context of the command
	ctx context.Context
}

// function to execute Autopilot code gen from another repository
func (c Command) Execute() error {
	c.ctx = context.Background()
	c.moduleRoot = util.GetModuleRoot()
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

	writer := &writer.DefaultFileWriter{Root: c.moduleRoot}

	apiTypes, err := render.RenderApiTypes(grp)
	if err != nil {
		return err
	}

	if err := writer.WriteFiles(apiTypes); err != nil {
		return err
	}

	manifests, err := render.RenderManifests(c.AppName, c.ManifestRoot, grp)
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

func (c Command) compileProtos(grp render.Group) error {
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

	if err := proto.CompileProtos(
		grp.Module,
		grp.ApiRoot,
		grp.ProtoDir,
	); err != nil {
		return err
	}

	return nil
}
