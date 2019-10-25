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

type EzKube interface {
	Ensure(ctx context.Context, obj KubeObj) error
	UpdateStatus(ctx context.Context, obj KubeObj) error
	Get(ctx context.Context, obj KubeObj) error

	GetDeployment(ctx context.Context, name, namespace string) (*v1.Deployment, error)
}

type ezKube struct {
	controlObj KubeObj
	mgr        manager.Manager
}

func NewEzKube(controlObj KubeObj, mgr manager.Manager) *ezKube {
	return &ezKube{controlObj: controlObj, mgr: mgr}
}

// ensures the object is written. first attempts to create, if fail, fall back to update
// sets controller reference on the object
func (m *ezKube) Ensure(ctx context.Context, obj KubeObj) error {
	if err := ctrl.SetControllerReference(m.controlObj, obj, m.mgr.GetScheme()); err != nil {
		return err
	}
	err := m.mgr.GetClient().Create(ctx, obj)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			updated, err := m.updateResourceVersion(ctx, obj)
			if err != nil {
				return err
			}
			if updated {
				if err := m.mgr.GetClient().Update(ctx, obj); err != nil {
					return err
				}
			}
			return nil
		}
		return err
	}

	return nil
}

func (m *ezKube) UpdateStatus(ctx context.Context, obj KubeObj) error {
	return m.mgr.GetClient().Status().Update(ctx, obj)
}

func (m *ezKube) Get(ctx context.Context, obj KubeObj) error {
	objectKey := client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
	return  m.mgr.GetClient().Get(ctx, objectKey, obj)
}

// do an HTTP GET to update the resource version of the desired object
func (m *ezKube) updateResourceVersion(ctx context.Context, obj KubeObj) (bool, error) {
	current := obj.DeepCopyObject().(KubeObj)
	objectKey := client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
	if err := m.mgr.GetClient().Get(ctx, objectKey, current); err != nil {
		return false, err
	}
	updated := obj.GetResourceVersion() != current.GetResourceVersion()
	if updated {
		obj.SetResourceVersion(current.GetResourceVersion())
	}
	return updated, nil
}

func (m *ezKube) GetDeployment(ctx context.Context, namespace, name string) (*v1.Deployment, error) {
	obj := &v1.Deployment{}
	objectKey := client.ObjectKey{Namespace: namespace, Name: name}
	return obj, m.mgr.GetClient().Get(ctx, objectKey, obj)
}
