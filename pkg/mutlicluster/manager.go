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

func NewContextualManager(parentCtx context.Context, cfg *rest.Config, namespace string) (*ContextualManager, error) {
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
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
