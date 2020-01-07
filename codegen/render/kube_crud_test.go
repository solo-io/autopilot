package render_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/autopilot/cli/pkg/utils"
	"github.com/solo-io/autopilot/codegen/model"
	. "github.com/solo-io/autopilot/codegen/render/api/things.test.io/v1"
	"github.com/solo-io/autopilot/codegen/render/api/things.test.io/v1/clientset/versioned"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/test"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/randutils"
	kubehelp "github.com/solo-io/go-utils/testutils/kube"
	"io/ioutil"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
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
					Status: &model.Field{Type: "PaintStatus"},
				},
			},
		}

		ns   string
		kube kubernetes.Interface
	)
	BeforeEach(func() {
		group.Init()
		err := applyFile("things.test.io-v1-crds.yaml")
		Expect(err).NotTo(HaveOccurred())
		ns = randutils.RandString(4)
		kube = kubehelp.MustKubeClient()
		err = kubeutils.CreateNamespacesInParallel(kube, ns)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		group.Init()
		err := deleteFile("things.test.io-v1-crds.yaml")
		Expect(err).NotTo(HaveOccurred())
		err = kubeutils.DeleteNamespacesInParallelBlocking(kube, ns)
		Expect(err).NotTo(HaveOccurred())
	})

	It("uses the generated clientsets to crud", func() {
		clientset, err := versioned.NewForConfig(test.MustConfig())
		Expect(err).NotTo(HaveOccurred())

		paint, err := clientset.ThingsV1().Paints(ns).Create(&Paint{
			ObjectMeta: v1.ObjectMeta{
				Name: "paint-1",
			},
			Spec: PaintSpec{
				Color: &PaintColor{
					Hue:   "prussian blue",
					Value: 0.5,
				},
				PaintType: &PaintSpec_Acrylic{
					Acrylic: &AcrylicType{
						Body: AcrylicType_Heavy,
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		written, err := clientset.ThingsV1().Paints(ns).Get(paint.Name, v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		Expect(written.Spec).To(Equal(paint.Spec))

		status := PaintStatus{
			ObservedGeneration: written.Generation,
			PercentRemaining:   22,
		}

		written.Status = status

		_, err = clientset.ThingsV1().Paints(ns).UpdateStatus(written)
		Expect(err).NotTo(HaveOccurred())

		written, err = clientset.ThingsV1().Paints(ns).Get(paint.Name, v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		Expect(written.Status).To(Equal(status))
	})
})
