package ezkube

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Our manager returns a manager.Manager
type ManagerGetter interface {
	Manager() manager.Manager
}

// ReconcileFuncs are used to dynamically update existing objects
// on calls to Update and Ensure.
// ReconcileFuncs perform a comparison of the the existing object with the desired object prior to writing to kube storage.
//
// The returned object will be applied to the cluster
// If a nil, nil is returned, the write will be skipped
type ReconcileFunc func(old Object, new Object) (*Object, error)

// RestClient is a wrapper for the dynamic controller-runtime client.Client
type RestClient interface {
	ManagerGetter

	// the Object passed will be updated to match the server version.
	// only key (namespace/name) is required
	Get(ctx context.Context, obj Object) error

	// Object passed will be updated to match the server version.
	// Object should be a List object
	// 	example:
	//
	// 	Client.List(ctx,
	// 		client.InNamespace("my-namespace",
	// 		client.MatchingLabels{"app": "petstore"})
	List(ctx context.Context, obj List, options ...client.ListOption) error

	// create the object passed
	Create(ctx context.Context, obj Object) error

	// update the object passed. does not update object subresources such as status.
	//
	// optional reconcile funcs can be passed which determine how
	// a child object should be reconciled with the existing object
	// when it already exists in the cluster
	Update(ctx context.Context, obj Object, reconcileFuncs ...ReconcileFunc) error

	// update the status of the object passed.
	UpdateStatus(ctx context.Context, obj Object) error

	// delete the object. only key (namespace/name) is required
	Delete(ctx context.Context, obj Object) error
}

// an Ensurer "ensures" that the given object will be created/applied to the cluster
// the object is applied after a resource version conflict.
// Warning: this can lead to race conditions if it is called asynchronously for the same resource.
// The ensured resource will have its controller reference set to the parent resource
type Ensurer interface {
	RestClient
	// optional reconcile funcs can be passed which determine how
	// a child object should be reconciled with the existing object
	// when it already exists in the cluster
	Ensure(ctx context.Context, parent Object, child Object, reconcileFuncs ...ReconcileFunc) error
}

// Ctl is a Control object for interacting with the Kubernetes API.
// Its meant as a generic, high-level abstraction over kubernetes client-go,
// similar to an in-process kubectl
type Ctl interface {
	Ensurer
}
