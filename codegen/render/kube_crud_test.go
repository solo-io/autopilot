package render_test

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/solo-io/autopilot/pkg/reconcile"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/autopilot/cli/pkg/utils"
	. "github.com/solo-io/autopilot/codegen/test/api/things.test.io/v1"
	"github.com/solo-io/autopilot/codegen/test/api/things.test.io/v1/controller"
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
	path := filepath.Join(util.GetModuleRoot(), "codegen/test/chart/crds", file)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return utils.KubectlApply(b)
}

var _ = Describe("Generated Code", func() {
	var (
		ns        string
		kube      kubernetes.Interface
		clientSet Clientset
		logLevel  = zap.NewAtomicLevel()
		ctx       = context.TODO()
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
		clientSet, err = NewClientsetFromConfig(test.MustConfig())
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kube, ns)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("kube clientsests", func() {
		It("uses the generated clientsets to crud", func() {

			paint := &Paint{
				ObjectMeta: v1.ObjectMeta{
					Name:      "paint-1",
					Namespace: ns,
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
			}

			err := clientSet.Paints().CreatePaint(ctx, paint)
			Expect(err).NotTo(HaveOccurred())

			written, err := clientSet.Paints().GetPaint(ctx, client.ObjectKey{
				Namespace: paint.Namespace,
				Name:      paint.Name,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(written.Spec).To(Equal(paint.Spec))

			status := PaintStatus{
				ObservedGeneration: written.Generation,
				PercentRemaining:   22,
			}

			written.Status = status

			err = clientSet.Paints().UpdatePaintStatus(ctx, written)
			Expect(err).NotTo(HaveOccurred())

			written, err = clientSet.Paints().GetPaint(ctx, client.ObjectKey{
				Namespace: paint.Namespace,
				Name:      paint.Name,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(written.Status).To(Equal(status))
		})
	})

	Context("kube reconciler", func() {
		var (
			mgr    manager.Manager
			cancel = func() {}
		)
		BeforeEach(func() {
			mgr, cancel = test.MustManager(ns)
		})
		AfterEach(cancel)

		It("uses the generated controller to reconcile", func() {

			paint := &Paint{
				ObjectMeta: v1.ObjectMeta{
					Name:      "paint-2",
					Namespace: ns,
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
			}

			err := clientSet.Paints().CreatePaint(ctx, paint)
			Expect(err).NotTo(HaveOccurred())

			paint.GetObjectKind().GroupVersionKind()

			loop := controller.NewPaintReconcileLoop("blick", mgr)

			var reconciled *Paint
			var deleted reconcile.Request
			reconciler := &controller.PaintReconcilerFuncs{
				OnReconcilePaint: func(obj *Paint) (result reconcile.Result, err error) {
					reconciled = obj
					return
				},
				OnReconcilePaintDeletion: func(req reconcile.Request) {
					deleted = req
					return
				},
			}
			err = loop.RunPaintReconciler(ctx, reconciler)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() *Paint {
				return reconciled
			}, time.Second).ShouldNot(BeNil())

			// update
			paint.Spec.Color = &PaintColor{Value: 0.7}

			err = clientSet.Paints().UpdatePaint(ctx, paint)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() PaintSpec {
				return reconciled.Spec
			}, time.Second).Should(Equal(paint.Spec))

			// delete
			err = clientSet.Paints().DeletePaint(ctx, client.ObjectKey{
				Name:      paint.Name,
				Namespace: paint.Namespace,
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() reconcile.Request {
				return deleted
			}, time.Second).Should(Equal(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      paint.Name,
				Namespace: paint.Namespace,
			}}))
		})
	})
})
