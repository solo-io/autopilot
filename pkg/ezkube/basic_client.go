package ezkube

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type restClient struct {
	mgr manager.Manager
}

var _ RestClient = &restClient{}

// NewRestClient creates the root REST client used to interact
// with Kubernetes. Wraps the controller-runtime client.Client.
// TODO: option to skip cache reads
func NewRestClient(mgr manager.Manager) *restClient {
	return &restClient{mgr: mgr}
}

func (c *restClient) Get(ctx context.Context, obj Object) error {
	objectKey := client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
	return c.mgr.GetClient().Get(ctx, objectKey, obj)
}

func (c *restClient) List(ctx context.Context, obj List, options ...client.ListOption) error {
	return c.mgr.GetClient().List(ctx, obj, options...)
}

func (c *restClient) Create(ctx context.Context, obj Object) error {
	return c.mgr.GetClient().Create(ctx, obj)
}

// if one or more reconcile funcs are passed, this will
// in
func (c *restClient) Update(ctx context.Context, obj Object, reconcileFuncs ...ReconcileFunc) error {

	return c.mgr.GetClient().Update(ctx, obj)
}

func (c *restClient) UpdateStatus(ctx context.Context, obj Object) error {
	return c.mgr.GetClient().Status().Update(ctx, obj)
}

func (c *restClient) Delete(ctx context.Context, obj Object) error {
	return c.mgr.GetClient().Delete(ctx, obj)
}

func (c *restClient) Manager() manager.Manager {
	return c.mgr
}
