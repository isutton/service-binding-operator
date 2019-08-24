package servicebindingrequest

import (
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SBRRequestMapper is the handler.SBRRequestMapper interface implementation. It should influence the
// enqueue process considering the resources informed.
type SBRRequestMapper struct{}

// Map execute the mapping of a resource with the requests it would produce. Here we inspect the
// given object trying to identify if this object is part of a SBR, or a actual SBR resource.
func (m *SBRRequestMapper) Map(obj handler.MapObject) []reconcile.Request {
	toReconcile := []reconcile.Request{}

	sbrNamespacedName, err := GetSBRNamespacedNameFromObject(obj.Object)
	if err != nil {
		// TODO: add proper logging for this error;
		return toReconcile
	}
	if IsSBRNamespacedNameEmpty(sbrNamespacedName) {
		return toReconcile
	}

	// TODO: add logging to show which SBR objects are taken from which other resources;
	return append(toReconcile, reconcile.Request{NamespacedName: sbrNamespacedName})
}
