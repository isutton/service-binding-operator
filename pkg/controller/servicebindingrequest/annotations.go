package servicebindingrequest

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

const (
	sbrNamespaceAnnotation = "service-binding-operator.apps.openshift.io/binding-namespace"
	sbrNameAnnotation      = "service-binding-operator.apps.openshift.io/binding-name"
)

// extractNamespacedName returns a types.NamespacedName if the required service binding request keys
// are present in the given data
func extractNamespacedName(data map[string]string) types.NamespacedName {
	namespacedName := types.NamespacedName{}
	ns, exists := data[sbrNamespaceAnnotation]
	if !exists {
		return namespacedName
	}
	name, exists := data[sbrNameAnnotation]
	if !exists {
		return namespacedName
	}
	namespacedName.Namespace = ns
	namespacedName.Name = name
	return namespacedName
}

// GetSBRNamespacedNameFromObject returns a types.NamespacedName if the required service binding
// request annotations are present in the given runtime.Object, empty otherwise. When annotations are
// not present, it checks if the object is an actual SBR, returning the details when positive. An
// error can be returned in the case the object can't be decoded.
func GetSBRNamespacedNameFromObject(obj runtime.Object) (types.NamespacedName, error) {
	namespacedName := types.NamespacedName{}
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return namespacedName, err
	}

	u := &unstructured.Unstructured{Object: data}
	namespacedName = extractNamespacedName(u.GetAnnotations())
	if !IsSBRNamespacedNameEmpty(namespacedName) {
		return namespacedName, nil
	}

	// TODO: a constant to define the kind ServiceBindingRequest;
	// TODO: should we return error in this case?
	if u.GroupVersionKind() != v1alpha1.SchemeGroupVersion.WithKind("ServiceBindingRequest") {
		return namespacedName, nil
	}

	namespacedName.Namespace = u.GetNamespace()
	namespacedName.Name = u.GetName()
	return namespacedName, nil
}

// IsSBRNamespacedNameEmpty returns true if any of the fields from the given namespacedName is empty.
func IsSBRNamespacedNameEmpty(namespacedName types.NamespacedName) bool {
	return namespacedName.Namespace == "" || namespacedName.Name == ""
}

func SetSBRAnnotations() error {
	return nil
}
