package controller_test
//
//import (
//	"context"
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//	. "github.com/solo-io/autopilot/pkg/controller"
//	"github.com/solo-io/autopilot/pkg/ezkube"
//	aphandler "github.com/solo-io/autopilot/pkg/handler"
//	predicate2 "github.com/solo-io/autopilot/pkg/predicate"
//	"github.com/solo-io/autopilot/test"
//	"github.com/solo-io/go-utils/kubeutils"
//	"github.com/solo-io/go-utils/testutils"
//	v1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/labels"
//	"k8s.io/apimachinery/pkg/runtime"
//	"os"
//	"sigs.k8s.io/controller-runtime/pkg/log"
//	"sigs.k8s.io/controller-runtime/pkg/log/zap"
//	"sigs.k8s.io/controller-runtime/pkg/manager"
//	"sigs.k8s.io/controller-runtime/pkg/predicate"
//	"sigs.k8s.io/controller-runtime/pkg/reconcile"
//	"time"
//)
//
//func initManangerClientNamespace(ctx context.Context, kcEnv, namespace string) (manager.Manager, ezkube.RestClient, context.CancelFunc) {
//	kubeconfig := os.Getenv(kcEnv)
//	if kubeconfig == "" {
//		Fail("expected " + kcEnv + " to be a path to a kubeconfig file")
//	}
//
//	cfg, err := kubeutils.GetConfig("", kubeconfig)
//	Expect(err).NotTo(HaveOccurred())
//
//	mgr, cancel := test.ManagerWithOpts(cfg, manager.Options{
//		Namespace:namespace,
//		MetricsBindAddress: "0",
//	})
//	client := ezkube.NewRestClient(mgr)
//
//	err = client.Create(ctx, &v1.Namespace{
//		ObjectMeta: metav1.ObjectMeta{Name: namespace},
//	})
//	Expect(err).NotTo(HaveOccurred())
//
//	return mgr, client, cancel
//}
//
//var _ = FDescribe("MultiClusterController", func() {
//	var (
//		mgr1, mgr2       manager.Manager
//		client1, client2 ezkube.RestClient
//		cancel1, cancel2 context.CancelFunc
//		namespace        string
//		ctx              = context.TODO()
//	)
//	BeforeEach(func() {
//		log.SetLogger(zap.New())
//		namespace = "controller-test-" + testutils.RandString(6)
//
//		mgr1, client1, cancel1 = initManangerClientNamespace(ctx, "KUBECONFIG_1", namespace)
//		mgr2, client2, cancel2 = initManangerClientNamespace(ctx, "KUBECONFIG_2", namespace)
//	})
//	AfterEach(func() {
//		if client1 != nil {
//			_ = client1.Delete(ctx, &v1.Namespace{
//				ObjectMeta: metav1.ObjectMeta{Name: namespace},
//			})
//		}
//		if client2 != nil {
//			_ = client2.Delete(ctx, &v1.Namespace{
//				ObjectMeta: metav1.ObjectMeta{Name: namespace},
//			})
//		}
//		if cancel1 != nil {
//			cancel1()
//		}
//		if cancel2 != nil {
//			cancel2()
//		}
//	})
//	It("watches the top level resource and its dependencies across clusters", func() {
//		logger := predicate2.EventLogger{log.Log}
//
//		recFunc := func(primaryResource runtime.Object) (result reconcile.Result, e error) {
//			result.RequeueAfter = time.Second
//			return
//		}
//
//		labelMatch := map[string]string{"service": "mesh"}
//
//		predicates := []predicate.Predicate{
//			predicate2.LabelMatcher{
//				Selector: labels.SelectorFromSet(labelMatch),
//			},
//			logger,
//		}
//
//		opts1 := Controller{
//			Name: "test-cluster1",
//			Ctx:  context.Background(),
//			Reconcile: func(primaryResource runtime.Object) (result reconcile.Result, e error) {
//				return recFunc(primaryResource)
//			},
//			PrimaryResource:   &v1.ConfigMap{},
//			PrimaryPredicates: predicates,
//			InputResources: map[runtime.Object][]predicate.Predicate{
//				&v1.Secret{}: predicates,
//			},
//		}
//		opts2 := Controller{
//			Name: "test-cluster2",
//			Ctx:  context.Background(),
//			Reconcile: func(primaryResource runtime.Object) (result reconcile.Result, e error) {
//				return recFunc(primaryResource)
//			},
//			PrimaryResource:   &v1.ConfigMap{},
//			PrimaryPredicates: predicates,
//			InputResources: map[runtime.Object][]predicate.Predicate{
//				&v1.Secret{}: predicates,
//			},
//		}
//
//		err := opts1.AddPrimaryWatchToManager(mgr1)
//		Expect(err).NotTo(HaveOccurred())
//		err = opts2.AddPrimaryWatchToManager(mgr2)
//		Expect(err).NotTo(HaveOccurred())
//
//		cm := v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
//			Name:      "top-level-resource",
//			Namespace: namespace,
//			Labels:    labelMatch,
//		}}
//		recFunc = func(primaryResource runtime.Object) (result reconcile.Result, e error) {
//			cm = *primaryResource.(*v1.ConfigMap)
//			return
//		}
//		data := map[string]string{"isti": "o"}
//		err = client1.Create(context.TODO(), &v1.ConfigMap{
//			ObjectMeta: cm.ObjectMeta,
//			Data:       data,
//		})
//		Expect(err).NotTo(HaveOccurred())
//		err = client2.Create(context.TODO(), &v1.ConfigMap{
//			ObjectMeta: cm.ObjectMeta,
//			Data:       data,
//		})
//		Expect(err).NotTo(HaveOccurred())
//		Eventually(func() map[string]string {
//			return cm.Data
//		}, time.Second*10).Should(Equal(data))
//
//		sec := v1.Secret{ObjectMeta: metav1.ObjectMeta{
//			Name:      "input-resource",
//			Namespace: namespace,
//			Labels:    labelMatch,
//		}}
//		recFunc = func(primaryResource runtime.Object) (result reconcile.Result, e error) {
//				e = client1.Get(context.TODO(), &sec)
//			return
//		}
//		secData := map[string][]byte{"linker": []byte("d")}
//		err = client1.Create(context.TODO(), &v1.Secret{
//			ObjectMeta: sec.ObjectMeta,
//			Data:       secData,
//		})
//		Expect(err).NotTo(HaveOccurred())
//		Eventually(func() map[string][]byte {
//			return sec.Data
//		}, time.Second*10).Should(Equal(secData))
//	})
//})
