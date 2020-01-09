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
		apiRoot = "codegen/render/api"
	)
	group.Init()

	FIt("compiles protos", func() {
		err := proto.CompileProtos(
			goModule,
			apiRoot,
			apiRoot,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	FIt("renders the files for the group", func() {
		files, err := RenderApiTypes(goModule, apiRoot, group)
		Expect(err).NotTo(HaveOccurred())

		w := &writer.DefaultWriter{
			Root: "api",
		}

		err = w.WriteFiles(files)
		Expect(err).NotTo(HaveOccurred())
	})
	It("generates kube clientset", func() {
		err := KubeCodegen(
			"codegen/render/api",
			group)
		Expect(err).NotTo(HaveOccurred())
	})
	It("generates the CRD manifest", func() {

		files, err := RenderManifests(group)
		Expect(err).NotTo(HaveOccurred())

		w := &writer.DefaultWriter{
			Root: "deploy",
		}

		err = w.WriteFiles(files)
		Expect(err).NotTo(HaveOccurred())

	})
})
