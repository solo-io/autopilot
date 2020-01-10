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

func MustManager(ns string) (manager.Manager, func()) {
	cfg := MustConfig()
	return ManagerWithOpts(cfg, manager.Options{Namespace: ns})
}

func ManagerWithOpts(cfg *rest.Config, opts manager.Options) (manager.Manager, func()) {
	ctx, cancel := context.WithCancel(context.TODO())

	mgr, err := manager.New(cfg, opts)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer ginkgo.GinkgoRecover()
		err = mgr.Start(ctx.Done())
		Expect(err).NotTo(HaveOccurred())
	}()

	mgr.GetCache().WaitForCacheSync(ctx.Done())

	return mgr, cancel
}
