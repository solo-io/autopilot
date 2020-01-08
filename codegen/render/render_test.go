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
			Resources: []Resource{
				{
					Kind:   "Paint",
					Spec:   Field{Type: "PaintSpec"},
					Status: &Field{Type: "PaintStatus"},
				},
			},
		}
	)
	group.Init()

	It("compiles protos", func() {
		err := proto.CompileProtos(".")
		Expect(err).NotTo(HaveOccurred())
	})

	It("renders the files for the group", func() {
		files, err := RenderApiTypes(group)
		Expect(err).NotTo(HaveOccurred())

		w := &writer.DefaultWriter{
			Root:           "api",
		}

		err = w.WriteFiles(files)
		Expect(err).NotTo(HaveOccurred())

		err = util.KubeCodegen(
			"things.test.io",
			"v1",
			"./codegen/render/api")
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
