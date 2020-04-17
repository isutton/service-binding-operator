package servicebindingrequest

import (
	"fmt"

	"github.com/imdario/mergo"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// GetEnvVars returns the data read from related resources (see ReadBindableResourcesData and
// ReadCRDDescriptionData).
func (r *Retriever) GetEnvVars() (map[string][]byte, error) {
	envVarCtx := make(map[string]interface{})

	for _, resource := range r.plan.GetRelatedResources() {
		// contribute values extracted from the service related resources
		err := mergo.Merge(&envVarCtx, resource.EnvVars, mergo.WithAppendSlice, mergo.WithOverride)
		if err != nil {
			return nil, err
		}

		// contribute the entire CR to the context
		gvk := resource.CR.GetObjectKind().GroupVersionKind()
		err = unstructured.SetNestedField(
			envVarCtx, resource.CR.Object, gvk.Version, gvk.Group, gvk.Kind, resource.CR.GetName())
		if err != nil {
			return nil, err
		}

		// FIXME(isuttonl): make volume keys a return value
		r.VolumeKeys = append(r.VolumeKeys, resource.VolumeKeys...)
	}

	envVarTemplates := r.plan.SBR.Spec.CustomEnvVar
	envParser := NewCustomEnvParser(envVarTemplates, envVarCtx)
	envVars, err := envParser.Parse()
	if err != nil {
		return nil, err
	}

	// convert values to a map[string][]byte
	result := make(map[string][]byte)
	for k, v := range envVars {
		result[k] = []byte(v.(string))
	}

	return result, nil
}

// ReadBindableResourcesData reads all related resources of a given sbr
func (r *Retriever) ReadBindableResourcesData(
	sbr *v1alpha1.ServiceBindingRequest,
	relatedResources RelatedResources,
) error {
	r.logger.Info("Detecting extra resources for binding...")
	for _, rs := range ([]*RelatedResource)(relatedResources) {
		b := NewDetectBindableResources(sbr, rs.CR, []schema.GroupVersionResource{
			{Group: "", Version: "v1", Resource: "configmaps"},
			{Group: "", Version: "v1", Resource: "services"},
			{Group: "route.openshift.io", Version: "v1", Resource: "routes"},
		}, r.client)

		vals, err := b.GetBindableVariables()
		if err != nil {
			return err
		}
		for k, v := range vals {
			r.store(rs.EnvVarPrefix, rs.CR, k, []byte(fmt.Sprintf("%v", v)))
		}
	}

	return nil
}
