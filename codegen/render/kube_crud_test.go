package render_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/autopilot/cli/pkg/utils"
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/render/api/things.test.io/v1/clientset/versioned"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/test"
	"io/ioutil"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"path/filepath"
)

func applyFile(file string) error {
	path := filepath.Join(util.MustGetThisDir(), "deploy", file)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return utils.KubectlApply(b)
}

func deleteFile(file string) error {
	path := filepath.Join(util.MustGetThisDir(), "deploy", file)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return utils.KubectlDelete(b)
}

var _ = Describe("Generated Clients", func() {
	var (
		group = model.Group{
			GroupVersion: schema.GroupVersion{
				Group:   "things.test.io",
				Version: "v1",
			},
			Resources: []model.Resource{
				{
					Kind:   "Paint",
					Spec:   model.Field{Type: "PaintSpec"},
					Status: &model.Field{Type: "TubeStatus"},
				},
			},
		}
	)
	BeforeEach(func() {
		group.Init()
		err := applyFile("things.test.io-v1-crds.yaml")
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		group.Init()
		err := deleteFile("things.test.io-v1-crds.yaml")
		Expect(err).NotTo(HaveOccurred())
	})

	It("uses the generated clientsets to crud", func() {
		clientset, err := versioned.NewForConfig(test.MustConfig())
		Expect(err).NotTo(HaveOccurred())

		things, err := clientset.ThingsV1().Paints("default").List(v1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(things.Items).To(HaveLen(1))
	})
})
