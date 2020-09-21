package servicebinding

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	sbrNamespaceAnnotation = "service-binding-operator.operators.coreos.com/binding-namespace"
	sbrNameAnnotation      = "service-binding-operator.operators.coreos.com/binding-name"
)

var (
	annotationsLog = log.NewLog("annotations")
)

// extractSBRNamespacedName returns a types.NamespacedName if the required service binding request keys
// are present in the given data
func extractSBRNamespacedName(data map[string]string) types.NamespacedName {
	namespacedName := types.NamespacedName{}
	ns, exists := data[sbrNamespaceAnnotation]
	if !exists || len(ns) == 0 {
		return namespacedName
	}
	name, exists := data[sbrNameAnnotation]
	if !exists || len(name) == 0 {
		return namespacedName
	}
	namespacedName.Namespace = ns
	namespacedName.Name = name
	return namespacedName
}

// getSBRNamespacedNameFromObject returns a types.NamespacedName if the required service binding
// request annotations are present in the given runtime.Object, empty otherwise. When annotations are
// not present, it checks if the object is an actual SBR, returning the details when positive. An
// error can be returned in the case the object can't be decoded.
func getSBRNamespacedNameFromObject(obj runtime.Object) (types.NamespacedName, error) {
	sbrNamespacedName := types.NamespacedName{}
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return sbrNamespacedName, err
	}

	u := &unstructured.Unstructured{Object: data}

	sbrNamespacedName = extractSBRNamespacedName(u.GetAnnotations())
	log := annotationsLog.WithValues(
		"Resource.GVK", u.GroupVersionKind(),
		"Resource.Namespace", u.GetNamespace(),
		"Resource.Name", u.GetName(),
		"SBR.NamespacedName", sbrNamespacedName.String(),
	)

	if isNamespacedNameEmpty(sbrNamespacedName) {
		log.Debug("SBR information not present in annotations, continue inspecting object")
	} else {
		log.Trace("SBR information found in annotations, returning it")
		return sbrNamespacedName, nil
	}

	if u.GroupVersionKind() == v1alpha1.SchemeGroupVersion.WithKind(serviceBindingRequestKind) {
		log.Debug("Object is a SBR, returning its namespaced name")
		sbrNamespacedName.Namespace = u.GetNamespace()
		sbrNamespacedName.Name = u.GetName()
		return sbrNamespacedName, nil
	}

	log.Trace("Object is not a SBR, returning an empty namespaced name")
	return types.NamespacedName{}, nil
}

// updateUnstructuredObj generic call to update the unstructured resource informed. It can return
// error when API update call does.
func updateUnstructuredObj(client dynamic.Interface, obj *unstructured.Unstructured) error {
	gvk := obj.GroupVersionKind()
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	opts := metav1.UpdateOptions{}

	log := annotationsLog.WithValues(
		"SBR.Namespace", obj.GetNamespace(),
		"SBR.Name", obj.GetName(),
		"Resource.GVK", gvk,
		"Resource.Namespace", obj.GetNamespace(),
		"Resource.Name", obj.GetName(),
	)
	log.Debug("Updating resource annotations...")

	_, err := client.Resource(gvr).Namespace(obj.GetNamespace()).Update(obj, opts)
	if err != nil {
		log.Error(err, "unable to set/update annotations in object")
	}
	return err
}

// setSBRAnnotations set annotations to include SBR information and return a new object.
func setSBRAnnotations(namespacedName types.NamespacedName,
	obj *unstructured.Unstructured) *unstructured.Unstructured {
	newObj := obj.DeepCopy()
	annotations := newObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[sbrNamespaceAnnotation] = namespacedName.Namespace
	annotations[sbrNameAnnotation] = namespacedName.Name
	newObj.SetAnnotations(annotations)
	return newObj
}

// setAndUpdateSBRAnnotations update existing annotations to include operator's. The annotations added are
// referring to a existing SBR namespaced name.
func setAndUpdateSBRAnnotations(
	client dynamic.Interface,
	namespacedName types.NamespacedName,
	objs []*unstructured.Unstructured,
) error {
	for _, obj := range objs {
		newObj := setSBRAnnotations(namespacedName, obj)
		equal, err := nestedUnstructuredComparison(obj, newObj, []string{"metadata", "annotations"}...)
		if err != nil {
			return err
		}
		if !equal.Success {
			if err := updateUnstructuredObj(client, newObj); err != nil {
				return err
			}
		}
	}
	return nil
}

// removeSBRAnnotations removes SBR related annotations and return a new object.
func removeSBRAnnotations(obj *unstructured.Unstructured) *unstructured.Unstructured {
	newObj := obj.DeepCopy()
	annotations := newObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	delete(annotations, sbrNameAnnotation)
	delete(annotations, sbrNamespaceAnnotation)
	newObj.SetAnnotations(annotations)
	return newObj
}

// removeAndUpdateSBRAnnotations removes SBR related annotations from all the objects and updates them using
// the given client.
func removeAndUpdateSBRAnnotations(client dynamic.Interface, objs []*unstructured.Unstructured) error {
	for _, obj := range objs {
		newObj := removeSBRAnnotations(obj)
		equal, err := nestedUnstructuredComparison(obj, newObj, []string{"metadata", "annotations"}...)
		if err != nil {
			return err
		}
		if !equal.Success {
			if err := updateUnstructuredObj(client, newObj); err != nil {
				return err
			}
		}
	}
	return nil
}
