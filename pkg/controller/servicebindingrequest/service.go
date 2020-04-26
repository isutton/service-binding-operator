package servicebindingrequest

import (
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
)

var (
	errBackingServiceNamespace = errors.New("backing Service Namespace is unspecified")
)

func findService(
	client dynamic.Interface,
	ns string,
	gvk schema.GroupVersionKind,
	resourceRef string,
) (
	*unstructured.Unstructured,
	error,
) {
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)

	if len(ns) == 0 {
		return nil, errBackingServiceNamespace
	}

	// delegate the search selector's namespaced resource client
	return client.
		Resource(gvr).
		Namespace(ns).
		Get(resourceRef, metav1.GetOptions{})
}

// CRDGVR is the plural GVR for Kubernetes CRDs.
var CRDGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1beta1",
	Resource: "customresourcedefinitions",
}

func findServiceCRD(client dynamic.Interface, gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	// gvr is the plural guessed resource for the given GVK
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	// crdName is the string'fied GroupResource, e.g. "deployments.apps"
	crdName := gvr.GroupResource().String()
	// delegate the search to the CustomResourceDefinition resource client
	return client.Resource(CRDGVR).Get(crdName, metav1.GetOptions{})
}

var EmptyBackingServiceSelectorsErr = errors.New("backing service selectors are empty")

func loadDescriptor(anns map[string]string, path string, descriptor string, root string) {
	if !strings.HasPrefix(descriptor, "binding:") {
		return
	}

	n := "servicebindingoperator.redhat.io/" + root + "." + path
	v := strings.Split(descriptor, ":")

	if strings.HasPrefix(descriptor, "binding:env:") {
		if len(v) > 4 {
			n = n + "-" + v[4]
			anns[n] = strings.Join(v[0:4], ":")
		}
		if len(v) == 4 {
			anns[n] = strings.Join(v[0:4], ":")
		}

	}

	if strings.HasPrefix(descriptor, "binding:volumemount:") {
		anns[n] = strings.Join(v[0:3], ":")
	}

}

func crdDescriptionToAnnotations(crdDescription *olmv1alpha1.CRDDescription) map[string]string {
	anns := make(map[string]string)
	for _, sd := range crdDescription.StatusDescriptors {
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "status")
		}
	}

	for _, sd := range crdDescription.SpecDescriptors {
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "spec")
		}
	}

	return anns
}

// findCRDDescription attempts to find the CRDDescription resource related CustomResourceDefinition.
func findCRDDescription(
	ns string,
	client dynamic.Interface,
	bssGVK schema.GroupVersionKind,
	crd *unstructured.Unstructured,
) (*olmv1alpha1.CRDDescription, error) {
	return NewOLM(client, ns).SelectCRDByGVK(bssGVK, crd)
}
