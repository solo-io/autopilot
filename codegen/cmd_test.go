package codegen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/autopilot/codegen/model"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/solo-io/autopilot/codegen"
)

var _ = Describe("Cmd", func() {
	It("generates controller code and manifests for a proto file", func() {

		cmd := &Command{
			Groups: []Group{
				{
					GroupVersion: schema.GroupVersion{
						Group:   "core",
						Version: "v1",
					},
					Module: "k8s.io/api",
					Resources: []Resource{
						{
							Kind: "Secret",
						},
					},
					RenderController:      true,
					CustomTypesImportPath: "k8s.io/api/core/v1",
					ApiRoot:               "codegen/render/api",
				},

				{
					GroupVersion: schema.GroupVersion{
						Group:   "things.test.io",
						Version: "v1",
					},
					Module: "github.com/solo-io/autopilot",
					Resources: []Resource{
						{
							Kind:   "Paint",
							Spec:   Field{Type: "PaintSpec"},
							Status: &Field{Type: "PaintStatus"},
						},
					},
					RenderProtos:     true,
					RenderManifests:  true,
					RenderTypes:      true,
					RenderClients:    true,
					RenderController: true,
					ApiRoot:          "codegen/render/api",
					ProtoDir:         "codegen/render/api",
				},
			},
		}

		err := cmd.Execute()
		Expect(err).NotTo(HaveOccurred())
	})
})
