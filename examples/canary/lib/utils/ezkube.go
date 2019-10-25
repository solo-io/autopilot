package utils

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type KubeObj interface {
	metav1.Object
	runtime.Object
}

type UpdateFunc func(obj1, obj2 KubeObj) (bool, error)

type EzKube interface {
	SetOwner(obj KubeObj)
	Ensure(ctx context.Context, obj KubeObj, update ...UpdateFunc) error
	UpdateStatus(ctx context.Context, obj KubeObj) error
	Get(ctx context.Context, obj KubeObj) error

	GetDeployment(ctx context.Context, name, namespace string) (*v1.Deployment, error)
}

type ezKube struct {
	controlObj KubeObj
	mgr        manager.Manager
}

func NewEzKube(mgr manager.Manager) *ezKube {
	return &ezKube{mgr: mgr}
}

func (m *ezKube) SetOwner(obj KubeObj) {
	m.controlObj = obj
}

// ensures the object is written. first attempts to create, if fail, fall back to update
// sets controller reference on the object
func (m *ezKube) Ensure(ctx context.Context, obj KubeObj, updateFuncs ...UpdateFunc) error {
	if err := ctrl.SetControllerReference(m.controlObj, obj, m.mgr.GetScheme()); err != nil {
		return err
	}

	orig := obj.DeepCopyObject().(KubeObj)
	if err := m.mgr.GetClient().Get(ctx, objKey(orig), orig); err != nil {
		if errors.IsNotFound(err) {
			return m.mgr.GetClient().Create(ctx, obj)
		}
		return err
	}

	for _, updateFn := range updateFuncs {
		shouldUpdate, err := updateFn(orig, obj)
		if err != nil {
			return err
		}
		if !shouldUpdate {
			return nil
		}
	}

	obj.SetResourceVersion(orig.GetResourceVersion())

	return m.mgr.GetClient().Update(ctx, obj)
}

func (m *ezKube) UpdateStatus(ctx context.Context, obj KubeObj) error {
	return m.mgr.GetClient().Status().Update(ctx, obj)
}

func (m *ezKube) Get(ctx context.Context, obj KubeObj) error {
	objectKey := client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
	return m.mgr.GetClient().Get(ctx, objectKey, obj)
}

func (m *ezKube) GetDeployment(ctx context.Context, namespace, name string) (*v1.Deployment, error) {
	obj := &v1.Deployment{}
	objectKey := client.ObjectKey{Namespace: namespace, Name: name}
	return obj, m.mgr.GetClient().Get(ctx, objectKey, obj)
}

func objKey(obj metav1.Object) client.ObjectKey {
	return client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}
