package test

import (
	"context"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func MustConfig() *rest.Config {
	cfg, err := config.GetConfig()
	Expect(err).NotTo(HaveOccurred())
	return cfg
}

func MustManager() (manager.Manager, func()) {
	ctx, cancel := context.WithCancel(context.TODO())

	cfg := MustConfig()
	mgr, err := manager.New(cfg, manager.Options{})
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer ginkgo.GinkgoRecover()
		err = mgr.Start(ctx.Done())
		Expect(err).NotTo(HaveOccurred())
	}()

	mgr.GetCache().WaitForCacheSync(ctx.Done())

	return mgr, cancel
}
