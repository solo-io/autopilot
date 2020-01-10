package codegen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/autopilot/codegen/model"
	. "github.com/solo-io/autopilot/codegen/render"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/solo-io/autopilot/codegen"
)

var _ = Describe("Cmd", func() {
	It("generates controller code and manifests for a proto file", func() {

		cmd := &Command{
			AppName:  "painting-app",
			ProtoDir: "codegen/render/api",
			Groups: []model.Group{
				{
					GroupVersion: schema.GroupVersion{
						Group:   "things.test.io",
						Version: "v1",
					},
					Module: "github.com/solo-io/wasme",
					Resources: []model.Resource{
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
				},
			},
			ApiRoot:      "codegen/render/api",
			ManifestRoot: "codegen/render",
		}

		err := cmd.Execute()
		Expect(err).NotTo(HaveOccurred())
	})
})
