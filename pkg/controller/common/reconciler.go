package common

import (
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// ReconcileRelatedSBR attempts to map an arbitrary object into a
// ServiceBindingRequest if the object contains the required information used
// to identify one.
func ReconcileRelatedSBR(o handler.MapObject) []reconcile.Request {
	logger := log.Log.WithName("catchall")

	var result []reconcile.Request

	sbrSelector, err := GetSBRNamespacedNameFromObject(o.Object)
	if err != nil {
		logger.Error(err, "error on extracting SBR namespaced-name from annotations")
	}

	if sbrSelector != nil {
		result = append(result, reconcile.Request{NamespacedName: *sbrSelector})
	}

	return result
}
