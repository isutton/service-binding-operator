package common

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddWatchesWithGVKs register watches for the given gvks in the controller
// using the requestsFunc mapping function.
func AddWatchesWithGVKs(c controller.Controller, gvks []schema.GroupVersionKind, requestsFunc handler.ToRequestsFunc) error {
	for _, gvk := range gvks {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		h := &handler.EnqueueRequestsFromMapFunc{ToRequests: requestsFunc}

		// FIXME: create a predicate to make sure we only allow reconciliation
		//        of objects that are having service-binding-operator
		//        annotations.
		if err := c.Watch(&source.Kind{Type: u}, h); err != nil {
			return err
		}
	}
	return nil
}
