package test

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func MustManager(namespace string) (manager.Manager, func()) {
	ctx, cancel := context.WithCancel(context.TODO())

	cfg := config.GetConfigOrDie()
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer ginkgo.GinkgoRecover()
		err = mgr.Start(ctx.Done())
		Expect(err).NotTo(HaveOccurred())
	}()

	mgr.GetCache().WaitForCacheSync(ctx.Done())

	return mgr, cancel
}
