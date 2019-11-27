package test_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	test2 "github.com/solo-io/autopilot/codegen/test"
	"github.com/solo-io/autopilot/test"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var _ = Describe("Bitconneeeeect", func() {
	It("work", func() {
		mgr, cancel := test.MustManager()
		defer cancel()

		cmReconciler := newBasicReconciler("configmap")
		secReconciler := newBasicReconciler("secret")

		g := &errgroup.Group{}

		g.Go(func() error {
			return test2.RegisterDynamicReconciler(cmReconciler.ctx.Done(), cmReconciler.name, mgr, cmReconciler, &v1.ConfigMap{})
		})

		g.Go(func() error {
			return test2.RegisterDynamicReconciler(secReconciler.ctx.Done(), secReconciler.name, mgr, secReconciler, &v1.Secret{})
		})

		go func() {
			defer GinkgoRecover()
			err := g.Wait()
			Expect(err).NotTo(HaveOccurred())
		}()
	})
})

type basicReconciler struct {
	called int
	name   string
	ctx    context.Context
	cancel context.CancelFunc
}

func newBasicReconciler(name string) *basicReconciler {
	ctx, cancel := context.WithCancel(context.Background())
	return &basicReconciler{name: name, ctx: ctx, cancel: cancel}
}

func (r *basicReconciler) Reconcile(reconcile.Request) (reconcile.Result, error) {
	r.called++
	log.Printf("called %v times")
	return reconcile.Result{RequeueAfter: time.Second / 2}, nil
}
