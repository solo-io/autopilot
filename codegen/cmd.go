package codegen

import (
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/proto"
	"github.com/solo-io/autopilot/codegen/render"
	"github.com/solo-io/autopilot/codegen/writer"
)

// runs the codegen compilation for the current Go module
type Command struct {
	// Should we compile protos?
	CompileProtos bool

	// Should we generate kubernetes manifests?
	GenerateKubeManifests bool

	// Should we generate kubernetes Go structs?
	GenerateKubeStructs bool

	// Should we generate kubernetes Go clients?
	GenerateKubeClients bool

	// Should we generate kubernetes Go controllers?
	GenerateKubeController bool

	// search protos recursively starting from this directory
	// TODO: use anyvendor to support external imports
	ProtoDir string

	// the k8s api groups for which to compile
	Groups []model.Group

	// root for all the api types, clientsets, etc
	ApiWriter writer.Writer

	// root for deployment manifests (kube YAML)
	ManifestWriter writer.Writer
}

// function to execute Autopilot code gen from another repository
func (c Command) Execute() error {
	if c.CompileProtos {
		if err := proto.CompileProtos(c.ProtoDir); err != nil {
			return err
		}
	}
	if err := c.WriteGeneratedFiles(); err != nil {
		return err
	}
	return nil
}

func (c Command) WriteGeneratedFiles() error {
	for _, group := range c.Groups {
		// init connects children to their parents
		group.Init()

		if c.GenerateKubeStructs {
			apiTypes, err := render.RenderApiTypes(group)
			if err != nil {
				return err
			}

			if err := c.ApiWriter.WriteFiles(apiTypes); err != nil {
				return err
			}
		}

		if c.GenerateKubeManifests {
			manifests, err := render.RenderApiTypes(group)
			if err != nil {
				return err
			}

			if err := c.ManifestWriter.WriteFiles(manifests); err != nil {
				return err
			}
		}

		if c.GenerateKubeClients {
			// only writes if the ApiWriter is real
			// we need to generate kube clientset in the same dir as the types
			// this constraint can be relaxed by modifying the render.KubeCodegen function
			realWriter, ok := c.ApiWriter.(*writer.DefaultWriter)
			if ok {
				if err := render.KubeCodegen(realWriter.Root, group); err != nil{
					return err
				}
			}
		}
	}

	return nil
}
