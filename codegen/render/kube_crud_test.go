package render_test

import (
	"io/ioutil"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/autopilot/cli/pkg/utils"
	. "github.com/solo-io/autopilot/codegen/render/api/things.test.io/v1"
	"github.com/solo-io/autopilot/codegen/render/api/things.test.io/v1/clientset/versioned"
	"github.com/solo-io/autopilot/codegen/render/api/things.test.io/v1/controller"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/test"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/randutils"
	kubehelp "github.com/solo-io/go-utils/testutils/kube"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/log"
	zaputil "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func applyFile(file string) error {
	path := filepath.Join(util.MustGetThisDir(), "manifests", file)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return utils.KubectlApply(b)
}

func deleteFile(file string) error {
	path := filepath.Join(util.MustGetThisDir(), "manifests", file)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return utils.KubectlDelete(b)
}

var _ = Describe("Generated Code", func() {
	var (
		ns        string
		kube      kubernetes.Interface
		clientset versioned.Interface
		logLevel  = zap.NewAtomicLevel()
	)
	BeforeEach(func() {
		logLevel.SetLevel(zap.DebugLevel)
		log.SetLogger(zaputil.New(
			zaputil.Level(&logLevel),
		))
		log.Log.Info("test")
		err := applyFile("things.test.io_v1_crds.yaml")
		Expect(err).NotTo(HaveOccurred())
		ns = randutils.RandString(4)
		kube = kubehelp.MustKubeClient()
		err = kubeutils.CreateNamespacesInParallel(kube, ns)
		Expect(err).NotTo(HaveOccurred())
		clientset, err = versioned.NewForConfig(test.MustConfig())
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := deleteFile("things.test.io_v1_crds.yaml")
		Expect(err).NotTo(HaveOccurred())
		err = kubeutils.DeleteNamespacesInParallelBlocking(kube, ns)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("kube clientsests", func() {
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

	Context("kube controllers", func() {
		var (
			mgr    manager.Manager
			cancel = func() {}
		)
		BeforeEach(func() {
			mgr, cancel = test.MustManager(ns)
		})
		AfterEach(cancel)

		It("uses the generated controller to reconcile", func() {

			ctl, err := controller.NewPaintController("blick", mgr)
			Expect(err).NotTo(HaveOccurred())

			var created, updated, deleted *Paint
			handler := &controller.PaintEventHandlerFuncs{
				OnCreate: func(obj *Paint) error {
					created = obj
					return nil
				},
				OnUpdate: func(_, new *Paint) error {
					updated = new
					return nil
				},
				OnDelete: func(obj *Paint) error {
					deleted = obj
					return nil
				},
			}

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

			paint.GetObjectKind().GroupVersionKind()

			err = ctl.AddEventHandler(handler)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() *Paint {
				return created
			}, time.Second).ShouldNot(BeNil())

			// update
			paint.Spec.Color = &PaintColor{Value: 0.7}

			paint, err = clientset.ThingsV1().Paints(ns).Update(paint)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() *Paint {
				return updated
			}, time.Second).ShouldNot(BeNil())

			// delete
			err = clientset.ThingsV1().Paints(ns).Delete(paint.Name, nil)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() *Paint {
				return deleted
			}, time.Second).ShouldNot(BeNil())
		})
	})
})
