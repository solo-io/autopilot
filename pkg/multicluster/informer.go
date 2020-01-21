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

type RemoveItemFunction func(name string) error
type AddConfigFunction func(cfg *rest.Config, name string) error
type AddManagerFunction func(mgr *ContextualManager, name string) error

// these functions aree intended to be used as callbacks for a resource watcher, where the
// resources represent KubeConfigs
type ClusterInformer struct {
	ClusterAdded   AddConfigFunction
	ClusterRemoved RemoveItemFunction
}

// These functions are intended to be used as an extension to the ClusterInformer.
// Only one manager needs to be created per cluster, so these callbacks will be
// called when a manager has been created for a given cluster
type ManagerInformer struct {
	ClusterAdded   AddManagerFunction
	ClusterRemoved RemoveItemFunction
}

type MultiClusterManager struct {
	ctx context.Context

	informerLock sync.RWMutex
	informers    map[string]ManagerInformer

	managerLock sync.RWMutex
	managers    map[string]*ContextualManager
}

func NewMultiClusterManager(ctx context.Context, mgr manager.Manager) (*MultiClusterManager, error) {
	ctxMgr := NewContextualManager(ctx, mgr)
	mcMgr := &MultiClusterManager{
		ctx: ctx,
		managers: map[string]*ContextualManager{
			"": ctxMgr,
		},
	}
	return mcMgr, nil
}



var (
	InformerAlreadyRegisteredError = eris.New("informer already registered with the given name")
	InformerNotRegisteredError     = eris.New("informer already registered with the given name")
)

func (m *MultiClusterManager) ClusterInformer() ClusterInformer {
	return ClusterInformer{
		ClusterAdded:   m.clusterAdded,
		ClusterRemoved: m.clusterRemoved,
	}
}

func (m *MultiClusterManager) AddInformer(informer ManagerInformer, name string) error {
	m.informerLock.Lock()
	defer m.informerLock.Unlock()
	_, ok := m.informers[name]
	if ok {
		return eris.Wrapf(InformerAlreadyRegisteredError, "informer name: %s", name)
	}
	m.informers[name] = informer
	return nil
}

func (m *MultiClusterManager) RemoveInformer(name string) error {
	m.informerLock.Lock()
	defer m.informerLock.Unlock()
	_, ok := m.informers[name]
	if ok {
		return eris.Wrapf(InformerNotRegisteredError, "informer name: %s", name)
	}
	delete(m.informers, name)
	return nil
}

func (m *MultiClusterManager) clusterAdded(cfg *rest.Config, name string) error {
	mgr, err := NewContextualManagerFromCfg(m.ctx, cfg, ContextualManagerOptions{})
	if err != nil {
		return err
	}
	go func() {
		if err := mgr.Manager().Start(mgr.Context().Done()); err != nil {
			contextutils.LoggerFrom(m.ctx).Errorf("could not start manager")
		}
	}()
	if synced := mgr.Manager().GetCache().WaitForCacheSync(m.ctx.Done()); !synced {
		go mgr.Stop()
		return eris.Errorf("unable to sync cache for cluster %s", name)
	}
	m.informerLock.RLock()
	for _, informer := range m.informers {
		if informer.ClusterAdded != nil {
			if err := informer.ClusterAdded(mgr, name); err != nil {
				return err
			}
		}
	}
	m.informerLock.RUnlock()
	m.managerLock.Lock()
	defer m.managerLock.Unlock()
	m.managers[name] = mgr
	return nil
}

func (m *MultiClusterManager) clusterRemoved(name string) error {
	m.managerLock.Lock()
	mgr, ok := m.managers[name]
	m.managerLock.Unlock()
	if !ok {
		return eris.Errorf("could not find manager for cluster %s", name)
	}
	go mgr.Stop()
	m.informerLock.RLock()
	for _, informer := range m.informers {
		if informer.ClusterRemoved != nil {
			if err := informer.ClusterRemoved(name); err != nil {
				return err
			}
		}
	}
	m.informerLock.RUnlock()
	delete(m.managers, name)
	return nil
}