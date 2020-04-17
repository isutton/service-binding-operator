package servicebindingrequest

import (
	"context"
	"errors"
	"strings"

	"github.com/imdario/mergo"

	"k8s.io/apimachinery/pkg/api/meta"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/annotations"
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
	ServiceContexts ServiceContexts                // CR and CRDDescription pairs SBR related
}

// GetCRs returns all collected service resources.
func (p *Plan) GetCRs() []*unstructured.Unstructured {
	return p.ServiceContexts.GetCRs()
}

// GetServiceContexts returns all collected service contexts.
func (p *Plan) GetServiceContexts() ServiceContexts {
	return p.ServiceContexts
}

// searchCR based on a CustomResourceDefinitionDescription and name, search for the object.
func (p *Planner) searchCR(selector v1alpha1.BackingServiceSelector) (*unstructured.Unstructured, error) {
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
	return p.client.Resource(gvr).Namespace(*selector.Namespace).Get(selector.ResourceRef, metav1.GetOptions{})
}

// CRDGVR is the plural GVR for Kubernetes CRDs.
var CRDGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1beta1",
	Resource: "customresourcedefinitions",
}

// searchCRD returns the CRD related to the gvk.
func (p *Planner) searchCRD(gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	// gvr is the plural guessed resource for the given GVK
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	// crdName is the string'fied GroupResource, e.g. "deployments.apps"
	crdName := gvr.GroupResource().String()
	// delegate the search to the CustomResourceDefinition resource client
	return p.client.Resource(CRDGVR).Get(crdName, metav1.GetOptions{})
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

func crdDescriptionToAnnotations(anns map[string]string, crdDescription *olmv1alpha1.CRDDescription) map[string]string {
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

func (p *Planner) getResourceInfo(
	cr *unstructured.Unstructured,
	bssGVK schema.GroupVersionKind,
) (
	*unstructured.Unstructured,
	*olmv1alpha1.CRDDescription,
	error,
) {
	crdDescription := &olmv1alpha1.CRDDescription{}

	// resolve the CRD using the service's GVK
	crd, err := p.searchCRD(bssGVK)
	if err != nil {
		// expected this to work, but didn't
		// if k8sError.IsNotFound(err) {...}
		p.logger.Error(err, "Probably not a CRD")

	} else {

		p.logger.Debug("Resolved CRD", "CRD", crd)

		olm := NewOLM(p.client, p.sbr.GetNamespace())

		// Parse annotations from the OLM descriptors or the CRD
		crdDescription, err = olm.SelectCRDByGVK(bssGVK, crd)
		if err != nil {
			p.logger.Error(err, "Probably not an OLM operator")
		}
		p.logger.Debug("Tentatively resolved CRDDescription", "CRDDescription", crdDescription)
	}

	err = buildCRDDescriptionFromCR(cr, crdDescription)
	if err != nil {
		return nil, nil, err
	}

	return crd, crdDescription, nil
}

// Plan by retrieving the necessary resources related to binding a service backend.
func (p *Planner) Plan() (*Plan, error) {
	ns := p.sbr.GetNamespace()

	var selectors []v1alpha1.BackingServiceSelector
	if p.sbr.Spec.BackingServiceSelector != nil {
		selectors = append(selectors, *p.sbr.Spec.BackingServiceSelector)
	}
	if p.sbr.Spec.BackingServiceSelectors != nil {
		selectors = append(selectors, *p.sbr.Spec.BackingServiceSelectors...)
	}

	if len(selectors) == 0 {
		return nil, EmptyBackingServiceSelectorsErr
	}

	ctxs := make([]*ServiceContext, 0)
	for _, s := range selectors {
		if s.Namespace == nil {
			s.Namespace = &ns
		}

		bssGVK := schema.GroupVersionKind{Kind: s.Kind, Version: s.Version, Group: s.Group}

		cr, err := p.searchCR(s)
		if err != nil {
			return nil, err
		}

		crd, crdDescription, err := p.getResourceInfo(cr, bssGVK)
		if err != nil {
			return nil, err
		}

		anns := crdDescriptionToAnnotations(crd.GetAnnotations(), crdDescription)
		volumeKeys := make([]string, 0)
		envVars := make(map[string]interface{})

		for n, v := range anns {
			h, err := annotations.BuildHandler(annotations.HandlerArgs{
				Name:     n,
				Value:    v,
				Resource: cr,
				Client:   p.client,
			})
			if err != nil {
				return nil, err
			}
			r, err := h.Handle()
			if err != nil {
				return nil, err
			}

			err = mergo.Merge(&envVars, r.Object, mergo.WithAppendSlice, mergo.WithOverride)
			if err != nil {
				return nil, err
			}

			// FIXME(isuttonl): rename volumeMounts to volumeKeys
			if r.Type == annotations.BindingTypeVolumeMount {
				volumeKeys = append(volumeKeys, r.Path)
			}
		}

		ctx := &ServiceContext{
			CRDDescription: crdDescription,
			CR:             cr,
			EnvVars:        envVars,
			VolumeKeys:     volumeKeys,
			EnvVarPrefix:   s.EnvVarPrefix,
		}

		ctxs = append(ctxs, ctx)
		p.logger.Debug("Resolved service context", "ServiceContext", ctx)
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
