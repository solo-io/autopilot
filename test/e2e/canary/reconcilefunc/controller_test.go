package reconcilefunc_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/autopilot/pkg/controller"
	"github.com/solo-io/autopilot/pkg/ezkube"
	aphandler "github.com/solo-io/autopilot/pkg/handler"
	"github.com/solo-io/autopilot/test"
	"github.com/solo-io/go-utils/testutils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var _ = Describe("Canary Reconcile", func() {
	var (
		client    ezkube.RestClient
		mgr       manager.Manager
		cancel    context.CancelFunc
		namespace string
		ctx       = context.TODO()
	)
	BeforeEach(func() {

		log.SetLogger(zap.New())

		mgr, cancel = test.MustManager(namespace)
		client = ezkube.NewRestClient(mgr)
		namespace = "controller-test-" + testutils.RandString(6)

		err := client.Create(ctx, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		_ = client.Delete(ctx, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		})
		cancel()
	})
	It("works", func() {
		logger := aphandler.LoggingPredicate{log.Log}

		recFunc := func(topLevelResource runtime.Object) (result reconcile.Result, e error) {
			result.RequeueAfter = time.Second
			return
		}

		labelMatch := map[string]string{"service": "mesh"}

		predicates := []predicate.Predicate{
			aphandler.LabelMatchingPredicate{
				Selector: labels.SelectorFromSet(labelMatch),
			},
			logger,
		}

		opts := Controller{
			Name:               "test",
			Ctx:                context.Background(),
			Reconcile: func(topLevelResource runtime.Object) (result reconcile.Result, e error) {
				return recFunc(topLevelResource)
			},
			TopLevelResource:   &v1.ConfigMap{},
			TopLevelPredicates: predicates,
			InputResources: map[runtime.Object][]predicate.Predicate{
				&v1.Secret{}: predicates,
			},
		}

		err := opts.AddToManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		cm := v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
			Name:      "top-level-resource",
			Namespace: namespace,
			Labels:    labelMatch,
		}}
		recFunc = func(topLevelResource runtime.Object) (result reconcile.Result, e error) {
			e = client.Get(context.TODO(), &cm)
			return
		}
		data := map[string]string{"isti": "o"}
		err = client.Create(context.TODO(), &v1.ConfigMap{
			ObjectMeta: cm.ObjectMeta,
			Data:       data,
		})
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() map[string]string {
			return cm.Data
		}, time.Second * 10).Should(Equal(data))

		sec := v1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      "input-resource",
			Namespace: namespace,
			Labels:    labelMatch,
		}}
		recFunc = func(topLevelResource runtime.Object) (result reconcile.Result, e error) {
			e = client.Get(context.TODO(), &sec)
			return
		}
		secData := map[string][]byte{"linker": []byte("d")}
		err = client.Create(context.TODO(), &v1.Secret{
			ObjectMeta: sec.ObjectMeta,
			Data:       secData,
		})
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() map[string][]byte {
			return sec.Data
		}, time.Second * 10).Should(Equal(secData))
	})
})
