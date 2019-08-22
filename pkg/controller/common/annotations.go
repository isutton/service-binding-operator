package common

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	sbrNamespaceAnnotation = "service-binding-operator.apps.openshift.io/binding-namespace"
	sbrNameAnnotation      = "service-binding-operator.apps.openshift.io/binding-name"
)

// extractNamespacedName returns a types.NamespacedName if the required service
// binding request keys are present in the given data
func extractNamespacedName(data map[string]string) *types.NamespacedName {
	ns, exists := data[sbrNamespaceAnnotation]
	if !exists {
		return nil
	}
	name, exists := data[sbrNameAnnotation]
	if !exists {
		return nil
	}
	return &types.NamespacedName{Namespace: ns, Name: name}
}

// GetSBRNamespacedNameFromObject returns a types.NamespacedName if the
// required service binding request annotations are present in the given
// runtime.Object, nil otherwise. An error can be returned in the case the
// object can't be decoded.
func GetSBRNamespacedNameFromObject(obj runtime.Object) (*types.NamespacedName, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	u := &unstructured.Unstructured{Object: data}
	return extractNamespacedName(u.GetAnnotations()), nil
}

// IsSBRNamespacedNameEmpty returns true if any of the fields from the given
// namespacedName is empty.
func IsSBRNamespacedNameEmpty(namespacedName types.NamespacedName) bool {
	return namespacedName.Namespace == "" || namespacedName.Name == ""
}

func SetSBRSelectorInObject() error {
	return nil
}
