package v1

import (
	"context"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// clienset for the things.test.io/v1 APIs
type Clientset interface {
	// clienset for the things.test.io/v1/v1 APIs
	Paints() PaintClient
}

type clientSet struct {
	client client.Client
}

func NewClientsetFromConfig(cfg *rest.Config) (*clientSet, error) {
	scheme := scheme.Scheme
	if err := AddToScheme(scheme); err != nil {
		return nil, err
	}
	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	return NewClientset(client), nil
}

func NewClientset(client client.Client) *clientSet {
	return &clientSet{client: client}
}

// clienset for the things.test.io/v1/v1 APIs
func (c *clientSet) Paints() PaintClient {
	return NewPaintClient(c.client)
}

// Reader knows how to read and list Paints.
type PaintReader interface {
	// Get retrieves a Paint for the given object key
	GetPaint(ctx context.Context, key client.ObjectKey) (*Paint, error)

	// List retrieves list of Paints for a given namespace and list options.
	ListPaint(ctx context.Context, opts ...client.ListOption) (*PaintList, error)
}

// Writer knows how to create, delete, and update Paints.
type PaintWriter interface {
	// Create saves the Paint object.
	CreatePaint(ctx context.Context, obj *Paint, opts ...client.CreateOption) error

	// Delete deletes the Paint object.
	DeletePaint(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error

	// Update updates the given Paint object.
	UpdatePaint(ctx context.Context, obj *Paint, opts ...client.UpdateOption) error

	// Patch patches the given Paint object.
	PatchPaint(ctx context.Context, obj *Paint, patch client.Patch, opts ...client.PatchOption) error

	// DeleteAllOf deletes all Paint objects matching the given options.
	DeleteAllOfPaint(ctx context.Context, opts ...client.DeleteAllOfOption) error
}

// StatusWriter knows how to update status subresource of a Paint object.
type PaintStatusWriter interface {
	// Update updates the fields corresponding to the status subresource for the
	// given Paint object.
	UpdatePaintStatus(ctx context.Context, obj *Paint, opts ...client.UpdateOption) error

	// Patch patches the given Paint object's subresource.
	PatchPaintStatus(ctx context.Context, obj *Paint, patch client.Patch, opts ...client.PatchOption) error
}

// Client knows how to perform CRUD operations on Paints.
type PaintClient interface {
	PaintReader
	PaintWriter
	PaintStatusWriter
}

type paintClient struct {
	client client.Client
}

func NewPaintClient(client client.Client) *paintClient {
	return &paintClient{client: client}
}

func (c *paintClient) GetPaint(ctx context.Context, key client.ObjectKey) (*Paint, error) {
	obj := &Paint{}
	if err := c.client.Get(ctx, key, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *paintClient) ListPaint(ctx context.Context, opts ...client.ListOption) (*PaintList, error) {
	list := &PaintList{}
	if err := c.client.List(ctx, list, opts...); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *paintClient) CreatePaint(ctx context.Context, obj *Paint, opts ...client.CreateOption) error {
	return c.client.Create(ctx, obj, opts...)
}

func (c *paintClient) DeletePaint(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error {
	obj := &Paint{}
	obj.SetName(key.Name)
	obj.SetNamespace(key.Namespace)
	return c.client.Delete(ctx, obj, opts...)
}

func (c *paintClient) UpdatePaint(ctx context.Context, obj *Paint, opts ...client.UpdateOption) error {
	return c.client.Update(ctx, obj, opts...)
}

func (c *paintClient) PatchPaint(ctx context.Context, obj *Paint, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c *paintClient) DeleteAllOfPaint(ctx context.Context, opts ...client.DeleteAllOfOption) error {
	obj := &Paint{}
	return c.client.DeleteAllOf(ctx, obj, opts...)
}

func (c *paintClient) UpdatePaintStatus(ctx context.Context, obj *Paint, opts ...client.UpdateOption) error {
	return c.client.Status().Update(ctx, obj, opts...)
}

func (c *paintClient) PatchPaintStatus(ctx context.Context, obj *Paint, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Status().Patch(ctx, obj, patch, opts...)
}
