package ezkube

import "context"

// package scoped utility functions for interacting with kubernetes REST APIs easily

// Uses the provided client to update the resource of the object
// by retrieving it from the server
func UpdateResourceVersion(c RestClient, ctx context.Context, obj Object) error {
	clone := obj.DeepCopyObject().(Object)

	if err := c.Get(ctx, clone); err != nil {
		return err
	}

	obj.SetResourceVersion(clone.GetResourceVersion())
	return nil
}
