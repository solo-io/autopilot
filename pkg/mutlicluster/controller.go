package multicluster

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/autopilot/codegen/render/api/core/v1/controller"
	"github.com/solo-io/go-utils/contextutils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	MultiClusterLabel      = "solo.io/kubeconfig"
	MultiClusterController = "multi-cluster-controller"
)

//go:generate mockgen -destination ./mocks/mc_manager.go github.com/solo-io/autopilot/pkg/multicluster  ClusterInformer

type ClusterInformer interface {
	ClusterAdded(cfg *rest.Config, name string) error
	ClusterRemoved(name string) error
}

type Controller struct {
	ctrl *controller.SecretController
	mgr  manager.Manager
}

type multiClusterController struct {
	ctx       context.Context
	mcManager ClusterInformer
	namespace string
	cs        *ClusterStore
}

// RemoteCluster defines cluster structZZ
type RemoteCluster struct {
	secretName string
}

// ClusterStore is a collection of clusters
type ClusterStore struct {
	remoteClusters map[string]*RemoteCluster
}

// newClustersStore initializes data struct to store clusters information
func newClustersStore() *ClusterStore {
	remoteClusters := make(map[string]*RemoteCluster)
	return &ClusterStore{
		remoteClusters: remoteClusters,
	}
}

// NewController returns a new secret controller
func NewController(
	ctx context.Context,
	mgr manager.Manager,
	mcManager ClusterInformer) (*Controller, error) {

	ctrl := &multiClusterController{
		ctx:       ctx,
		cs:        newClustersStore(),
		mcManager: mcManager,
	}

	secretCtrl, err := controller.NewSecretController(MultiClusterController, mgr)
	if err != nil {
		return nil, err
	}

	if err := secretCtrl.AddEventHandler(ctrl, &multiClusterPredicate{}); err != nil {
		return nil, err
	}

	return &Controller{ctrl: secretCtrl, mgr: mgr}, nil
}

func (c *Controller) Start(ctx context.Context) error {
	return c.mgr.Start(ctx.Done())
}

func (c *multiClusterController) addMemberCluster(s *v1.Secret) error {
	logger := contextutils.LoggerFrom(c.ctx)
	for clusterID, kubeConfig := range s.Data {
		// clusterID must be unique even across multiple secrets
		if _, ok := c.cs.remoteClusters[clusterID]; !ok {
			if len(kubeConfig) == 0 {
				logger.Infof("Data '%s' in the secret %s in namespace %s is empty, and disregarded ",
					clusterID, s.GetName(), s.ObjectMeta.Namespace)
				continue
			}

			clusterConfig, err := clientcmd.Load(kubeConfig)
			if err != nil {
				logger.Infof("Data '%s' in the secret %s in namespace %s is not a kubeconfig: %v",
					clusterID, s.GetName(), s.ObjectMeta.Namespace, err)
				logger.Infof("KubeConfig: '%s'", string(kubeConfig))
				continue
			}

			clientConfig := clientcmd.NewDefaultClientConfig(*clusterConfig, &clientcmd.ConfigOverrides{})

			restConfig, err := clientConfig.ClientConfig()
			if err != nil {
				return eris.Errorf("error during conversion of secret to client config: %v", err)
			}

			err = c.mcManager.ClusterAdded(restConfig, clusterID)
			if err != nil {
				return eris.Errorf("error during create of clusterID: %s %v", clusterID, err)
			}

			logger.Infof("Adding new cluster member: %s", clusterID)
			c.cs.remoteClusters[clusterID] = &RemoteCluster{
				secretName: s.GetName(),
			}
		} else {
			logger.Infof("Cluster %s in the secret %s in namespace %s already exists",
				clusterID, c.cs.remoteClusters[clusterID].secretName, s.ObjectMeta.Namespace)
		}
	}
	logger.Infof("Number of remote clusters: %d", len(c.cs.remoteClusters))
	return nil
}

func (c *multiClusterController) deleteMemberCluster(s *v1.Secret) error {
	logger := contextutils.LoggerFrom(c.ctx)
	for clusterID, cluster := range c.cs.remoteClusters {
		if cluster.secretName == s.GetName() {
			logger.Infof("Deleting cluster member: %s", clusterID)
			err := c.mcManager.ClusterRemoved(clusterID)
			if err != nil {
				return eris.Errorf("error during cluster delete: %s %v", clusterID, err)
			}
			delete(c.cs.remoteClusters, clusterID)
		}
	}
	logger.Infof("Number of remote clusters: %d", len(c.cs.remoteClusters))
	return nil
}

func (c *multiClusterController) Create(s *v1.Secret) error {
	return c.addMemberCluster(s)
}

func (c *multiClusterController) Update(old, new *v1.Secret) error {
	// If mc label has been removed from secret, remove from remote clusters
	if hasOwnerLabels(old.GetObjectMeta()) && !hasOwnerLabels(new.GetObjectMeta()) {
		return c.deleteMemberCluster(new)
	}
	// if mc label has been added to secret, add to remote cluster list
	if !hasOwnerLabels(old.GetObjectMeta()) && hasOwnerLabels(new.GetObjectMeta()) {
		return c.addMemberCluster(new)
	}
	return nil
}

func (c *multiClusterController) Delete(s *v1.Secret) error {
	return c.deleteMemberCluster(s)
}

func (c *multiClusterController) Generic(obj *v1.Secret) error {
	contextutils.LoggerFrom(c.ctx).Warn("should never be called as generic events are not configured")
	return nil
}

type multiClusterPredicate struct{}

func hasOwnerLabels(metadata metav1.Object) bool {
	val, ok := metadata.GetLabels()[MultiClusterLabel]
	return ok && val == "true"
}

func (m *multiClusterPredicate) Create(e event.CreateEvent) bool {
	return hasOwnerLabels(e.Meta)
}

func (m *multiClusterPredicate) Delete(e event.DeleteEvent) bool {
	return hasOwnerLabels(e.Meta)
}

func (m *multiClusterPredicate) Update(e event.UpdateEvent) bool {
	return hasOwnerLabels(e.MetaNew) || hasOwnerLabels(e.MetaOld)
}

func (m *multiClusterPredicate) Generic(e event.GenericEvent) bool {
	return hasOwnerLabels(e.Meta)
}
