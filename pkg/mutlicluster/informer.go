package multicluster

import (
	"context"
	"sync"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	MultiClusterLabel      = "solo.io/kubeconfig"
	MultiClusterController = "multi-cluster-controller"
)

//go:generate mockgen -destination ./mocks/mc_manager.go github.com/solo-io/autopilot/pkg/multicluster  ClusterInformer

// This interface is intended to be used as callbacks for a resource watcher, where the
// resources represent KubeConfigs
type ClusterInformer interface {
	ClusterAdded(cfg *rest.Config, name string) error
	ClusterRemoved(name string) error
}

// This interface is intended to be used as an extension to the ClusterInformer.
// Only one manager needs to be created per cluster, so these callbacks will be
// called when a manager has been created for a given cluster
type ManagerInformer interface {
	ClusterAdded(mgr *ContextualManager, name string) error
	ClusterRemoved(name string) error
}

type MultiClusterManager struct {
	ctx      context.Context
	informer ManagerInformer
	lock     sync.RWMutex
	managers map[string]*ContextualManager
}

func NewMultiClusterManager(ctx context.Context, mgr manager.Manager,
	informer ManagerInformer) (*MultiClusterManager, error) {

	ctxMgr := NewContextualManager(ctx, mgr)
	mcMgr := &MultiClusterManager{
		ctx: ctx,
		managers: map[string]*ContextualManager{
			"": ctxMgr,
		},
		informer: informer,
	}
	return mcMgr, nil
}

func (m *MultiClusterManager) ClusterAdded(cfg *rest.Config, name string) error {
	mgr, err := NewContextualManagerFromCfg(m.ctx, cfg, ContextualManagerOptions{})
	if err != nil {
		return err
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	go func() {
		if err := mgr.Manager().Start(mgr.Context().Done()); err != nil {
			contextutils.LoggerFrom(m.ctx).Errorf("could not start manager")
		}
	}()
	if err := m.informer.ClusterAdded(mgr, name); err != nil {
		return err
	}
	m.managers[name] = mgr
	return nil
}

func (m *MultiClusterManager) ClusterRemoved(name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	mgr, ok := m.managers[name]
	if !ok {
		return eris.Errorf("could not find manager for cluster %s", name)
	}
	go mgr.Stop()
	if err := m.informer.ClusterRemoved(name); err != nil {
		return err
	}
	delete(m.managers, name)
	return nil
}
