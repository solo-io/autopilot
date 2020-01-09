package render_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/autopilot/codegen/proto"
	. "github.com/solo-io/autopilot/codegen/render"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/codegen/writer"
	//"github.com/solo-io/autopilot/test"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Render", func() {
	var (
		group = Group{
			GroupVersion: schema.GroupVersion{
				Group:   "things.test.io",
				Version: "v1",
			},
			RenderTypes:      true,
			RenderManifests:  true,
			RenderClients:    true,
			RenderController: true,
			Resources: []Resource{
				{
					Kind:   "Paint",
					Spec:   Field{Type: "PaintSpec"},
					Status: &Field{Type: "PaintStatus"},
				},
			},
		}

		goModule = util.GetGoModule()
		apiDir   = "codegen/render/api"
	)
	group.Init()

	It("compiles protos", func() {
		err := proto.CompileProtos(
			goModule,
			apiDir,
			apiDir,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("renders the files for the group", func() {
		files, err := RenderApiTypes(goModule, apiDir, group)
		Expect(err).NotTo(HaveOccurred())

		w := &writer.DefaultWriter{
			Root: util.GetModuleRoot(),
		}

		err = w.WriteFiles(files)
		Expect(err).NotTo(HaveOccurred())
	})
	It("generates kube clientset", func() {
		err := KubeCodegen(
			apiDir,
			group)
		Expect(err).NotTo(HaveOccurred())
	})
	It("generates the CRD manifest", func() {

		files, err := RenderManifests("painting", "./manifests", group)
		Expect(err).NotTo(HaveOccurred())

		w := &writer.DefaultWriter{
			Root: "./codegen/render",
		}

		err = w.WriteFiles(files)
		Expect(err).NotTo(HaveOccurred())

	})
})
