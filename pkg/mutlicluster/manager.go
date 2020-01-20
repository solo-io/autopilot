package multicluster

import (
	"context"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ContextualManager struct {
	mgr    manager.Manager
	ctx    context.Context
	cancel context.CancelFunc
}

type ContextualManagerOptions struct {
	Namespace              string
	MetricsBindAddress     string
	HealthProbeBindAddress string
}

func (c ContextualManagerOptions) Convert() manager.Options {
	managerOpts := manager.Options{}
	if c.HealthProbeBindAddress == "" {
		managerOpts.HealthProbeBindAddress = "0"
	}
	if c.MetricsBindAddress == "" {
		managerOpts.MetricsBindAddress = "0"
	}
	return managerOpts
}

func NewContextualManager(parentCtx context.Context, mgr manager.Manager) *ContextualManager {
	ctx, cancel := context.WithCancel(parentCtx)
	return &ContextualManager{
		mgr:    mgr,
		ctx:    ctx,
		cancel: cancel,
	}
}

func NewContextualManagerFromCfg(parentCtx context.Context, cfg *rest.Config,
	opts ContextualManagerOptions) (*ContextualManager, error) {

	mgr, err := manager.New(cfg, opts.Convert())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parentCtx)
	return &ContextualManager{
		mgr:    mgr,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (c *ContextualManager) Manager() manager.Manager {
	return c.mgr
}

func (c *ContextualManager) Context() context.Context {
	return c.ctx
}

func (c *ContextualManager) Stop() {
	c.cancel()
}
