package servicebindingrequest

import (
	"context"
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

var (
	plannerLog                 = log.NewLog("planner")
	errBackingServiceNamespace = errors.New("backing Service Namespace is unspecified")
)

// Planner plans resources needed to bind a given backend service, using OperatorLifecycleManager
// standards and CustomResourceDefinitionDescription data to understand which attributes are needed.
type Planner struct {
	ctx    context.Context                 // request context
	client dynamic.Interface               // kubernetes dynamic api client
	sbr    *v1alpha1.ServiceBindingRequest // instantiated service binding request
	logger *log.Log                        // logger instance
}

// Plan outcome, after executing planner.
type Plan struct {
	Ns              string                         // namespace name
	Name            string                         // plan name, same than ServiceBindingRequest
	SBR             v1alpha1.ServiceBindingRequest // service binding request
	ServiceContexts ServiceContextList             // CR and CRDDescription pairs SBR related
}

func findCR(client dynamic.Interface, selector v1alpha1.BackingServiceSelector) (*unstructured.Unstructured, error) {
	// gvr is the plural guessed resource for the given selector
	gvk := schema.GroupVersionKind{
		Group:   selector.Group,
		Version: selector.Version,
		Kind:    selector.Kind,
	}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)

	if selector.Namespace == nil {
		return nil, errBackingServiceNamespace
	}

	// delegate the search selector's namespaced resource client
	return client.Resource(gvr).Namespace(*selector.Namespace).Get(selector.ResourceRef, metav1.GetOptions{})
}

// CRDGVR is the plural GVR for Kubernetes CRDs.
var CRDGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1beta1",
	Resource: "customresourcedefinitions",
}

func findCRD(client dynamic.Interface, gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
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

// Plan by retrieving the necessary resources related to binding a service backend.
func (p *Planner) Plan() (*Plan, error) {
	sbr := p.sbr
	ns := sbr.GetNamespace()
	client := p.client
	inSelector := sbr.Spec.BackingServiceSelector
	inSelectors := sbr.Spec.BackingServiceSelectors
	var selectors []v1alpha1.BackingServiceSelector

	// FIXME(isuttonl): move the selectors compoosition to the caller.
	if inSelector != nil {
		selectors = append(selectors, *inSelector)
	}
	if inSelectors != nil {
		selectors = append(selectors, *inSelectors...)
	}
	if len(selectors) == 0 {
		return nil, EmptyBackingServiceSelectorsErr
	}

	ctxs, err := buildServiceContexts(client, ns, selectors)
	if err != nil {
		return nil, err
	}

	return &Plan{
		Name:            p.sbr.GetName(),
		Ns:              ns,
		ServiceContexts: ctxs,
		SBR:             *p.sbr,
	}, nil
}

// NewPlanner instantiate Planner type.
func NewPlanner(
	ctx context.Context,
	client dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
) *Planner {
	return &Planner{
		ctx:    ctx,
		client: client,
		sbr:    sbr,
		logger: plannerLog,
	}
}
