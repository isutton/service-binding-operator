package servicebindingrequest

import (
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ObjectToSBRMapper struct{}

func (m *ObjectToSBRMapper) Map(obj handler.MapObject) []reconcile.Request {
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
