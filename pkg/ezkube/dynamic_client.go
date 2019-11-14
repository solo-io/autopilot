package ezkube

import (
	"context"

	"github.com/solo-io/autopilot/pkg/utils"
	"k8s.io/client-go/util/retry"

	"k8s.io/apimachinery/pkg/api/errors"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Our manager returns a manager.Manager
type Manager interface {
	Manager() manager.Manager
}

// the reconcile func determines how to reconcile the old object with the new
// when the Ensurer performs an automatic update.
// The returned object will be applied to the cluster
// Any errors returned will be returned by the Ensure call
// If a nil, nil is returned, Ensure will skip the update
type ReconcileFunc func(old Object, new Object) (*Object, error)

// an Ensurer "ensures" that the given object will be created/applied to the cluster
// the object is applied after a resource version conflict.
// Warning: this can lead to race conditions if it is called asynchronously for the same resource.
// The ensured resource will have its controller reference set to the parent resource
type Ensurer interface {
	// optional reconcile funcs can be passed which determine how
	// a child object should be reconciled with the existing object
	// when it already exists in the cluster
	Ensure(ctx context.Context, parent Object, child Object, reconcileFuncs ...ReconcileFunc) error
}

// Client is an interface for interacting with the k8s rest api
// It is functional with any kubernetes runtime.Object with k8s metadata
type Client interface {
	Manager
	Ensurer

	// the Object passed will be updated to match the server version.
	// only key (namespace/name) is required
	Get(ctx context.Context, obj Object) error

	// Object passed will be updated to match the server version.
	// Object should be a List object
	// 	example:
	//
	// 	simpleClient.List(ctx,
	// 		client.InNamespace("my-namespace",
	// 		client.MatchingLabels{"app": "petstore"})
	List(ctx context.Context, obj List, options ...client.ListOption) error

	// create the object passed
	Create(ctx context.Context, obj Object) error

	// update the object passed. does not update the object status
	Update(ctx context.Context, obj Object) error

	// update the status of the object passed.
	UpdateStatus(ctx context.Context, obj Object) error

	// delete the object. only key (namespace/name) is required
	Delete(ctx context.Context, obj Object) error
}

type simpleClient struct {
	mgr manager.Manager
}

var _ Client = &simpleClient{}

// NewClient creates an implementation of a Ensurer/Client based on a manager
func NewClient(mgr manager.Manager) *simpleClient {
	return &simpleClient{mgr: mgr}
}

func (c *simpleClient) Get(ctx context.Context, obj Object) error {
	objectKey := client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
	return c.mgr.GetClient().Get(ctx, objectKey, obj)
}

func (c *simpleClient) List(ctx context.Context, obj List, options ...client.ListOption) error {
	return c.mgr.GetClient().List(ctx, obj, options...)
}

func (c *simpleClient) Create(ctx context.Context, obj Object) error {
	return c.mgr.GetClient().Create(ctx, obj)
}

func (c *simpleClient) Update(ctx context.Context, obj Object) error {
	return c.mgr.GetClient().Update(ctx, obj)
}

func (c *simpleClient) UpdateStatus(ctx context.Context, obj Object) error {
	return c.mgr.GetClient().Status().Update(ctx, obj)
}

func (c *simpleClient) Delete(ctx context.Context, obj Object) error {
	return c.mgr.GetClient().Delete(ctx, obj)
}

func (c *simpleClient) Manager() manager.Manager {
	return c.mgr
}

func (c *simpleClient) Ensure(ctx context.Context, parent Object, child Object, reconcileFuncs ...ReconcileFunc) error {
	if parent != nil {
		if err := controllerruntime.SetControllerReference(parent, child, c.mgr.GetScheme()); err != nil {
			return err
		}
	}

	orig := child.DeepCopyObject().(Object)

	for _, reconcile := range reconcileFuncs {
		reconciledObj, err := reconcile(orig, child)
		if err != nil {
			return err
		}
		if reconciledObj == nil {
			return nil
		}
		child = *reconciledObj
	}

	if err := c.Get(ctx, orig); err != nil {
		if errors.IsNotFound(err) {
			return c.Create(ctx, child)
		}
		return err
	}

	child.SetResourceVersion(orig.GetResourceVersion())

	// retry on resource version conflict
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		err := c.Update(ctx, child)
		if errors.IsConflict(err) {
			utils.LoggerFromContext(ctx).Info("retrying on resource conflict")
			if err := c.UpdateResourceVersion(ctx, child); err != nil {
				return err
			}
		}
		return err
	})
}

func (c *simpleClient) UpdateResourceVersion(ctx context.Context, obj Object) error {

	clone := obj.DeepCopyObject().(Object)

	if err := c.Get(ctx, clone); err != nil {
		return err
	}

	obj.SetResourceVersion(clone.GetResourceVersion())
	return nil
}
