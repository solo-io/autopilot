package ezkube

import (
	"context"

	"github.com/solo-io/autopilot/pkg/utils"
	"k8s.io/client-go/util/retry"

	"k8s.io/apimachinery/pkg/api/errors"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

type ensurer struct {
	RestClient
}

var _ Ensurer = &ensurer{}

// Instantiate an Ensurer from a RestClient
func NewEnsurer(restClient RestClient) *ensurer {
	return &ensurer{RestClient: restClient}
}

func (c *ensurer) Ensure(ctx context.Context, parent Object, child Object, reconcileFuncs ...ReconcileFunc) error {
	if parent != nil {
		if err := controllerruntime.SetControllerReference(parent, child, c.Manager().GetScheme()); err != nil {
			return err
		}
	}

	existing := child.DeepCopyObject().(Object)

	if err := c.Get(ctx, existing); err != nil {
		if errors.IsNotFound(err) {
			return c.Create(ctx, child)
		}
		return err
	}

	child.SetResourceVersion(existing.GetResourceVersion())

	for _, reconcile := range reconcileFuncs {
		reconciledObj, err := reconcile(existing, child)
		if err != nil {
			return err
		}
		if reconciledObj == nil {
			return nil
		}
		child = *reconciledObj
	}

	// retry on resource version conflict
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		err := c.Update(ctx, child)
		if errors.IsConflict(err) {
			utils.LoggerFromContext(ctx).Info("retrying on resource conflict")
			if err := UpdateResourceVersion(c, ctx, child); err != nil {
				return err
			}
		}
		return err
	})
}
