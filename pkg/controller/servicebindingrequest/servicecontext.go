package servicebindingrequest

import (
	"github.com/imdario/mergo"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/annotations"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// ServiceContext contains information related to a service.
type ServiceContext struct {
	// Object is the resource being used as reference.
	Object *unstructured.Unstructured
	// EnvVars contains the service's contributed environment variables.
	EnvVars map[string]interface{}
	// VolumeKeys contains the keys that should be mounted as volume from the binding secret.
	VolumeKeys []string
	// EnvVarPrefix indicates the prefix to use in environment variables.
	EnvVarPrefix *string
}

// ServiceContextList is a list of ServiceContext values.
type ServiceContextList []*ServiceContext

// GetObjects returns a slice of service unstructured objects contained in the collection.
func (sc ServiceContextList) GetObjects() []*unstructured.Unstructured {
	var crs []*unstructured.Unstructured
	for _, s := range sc {
		crs = append(crs, s.Object)
	}
	return crs
}

// buildServiceContexts return a collection of ServiceContext values from the given service
// selectors.
//
// TODO(isuttonl): implement tests
func buildServiceContexts(
	client dynamic.Interface,
	ns string,
	selectors []v1alpha1.BackingServiceSelector,
) (ServiceContextList, error) {
	serviceCtxs := make([]*ServiceContext, 0)
	for _, s := range selectors {
		if s.Namespace == nil {
			s.Namespace = &ns
		}

		gvk := schema.GroupVersionKind{Kind: s.Kind, Version: s.Version, Group: s.Group}

		serviceCtx, err := buildServiceContext(client, *s.Namespace, gvk, s.ResourceRef)
		if err != nil {
			return nil, err
		}
		serviceCtxs = append(serviceCtxs, serviceCtx)
	}

	return serviceCtxs, nil
}

func buildServiceContext(
	client dynamic.Interface,
	ns string,
	gvk schema.GroupVersionKind,
	resourceRef string,
) (*ServiceContext, error) {
	obj, err := findService(client, ns, gvk, resourceRef)
	if err != nil {
		return nil, err
	}

	anns := map[string]string{}

	// attempt to search the CRD of given gvk and bail out right away if a CRD can't be found; this
	// means also a CRDDescription can't exist or if it does exist it is not meaningful.
	crd, err := findServiceCRD(client, gvk)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if !errors.IsNotFound(err) {
		// attempt to search the a CRDDescription related to the obtained CRD.
		crdDescription, err := findCRDDescription(ns, client, gvk, crd)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		// start with annotations extracted from CRDDescription
		err = mergo.Merge(
			&anns, convertCRDDescriptionToAnnotations(crdDescription), mergo.WithOverride)
		if err != nil {
			return nil, err
		}
		// then override collected annotations with CR annotations
		err = mergo.Merge(&anns, crd.GetAnnotations(), mergo.WithOverride)
		if err != nil {
			return nil, err
		}
	}

	// and finally override collected annotations with own annotations
	err = mergo.Merge(&anns, obj.GetAnnotations(), mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	volumeKeys := make([]string, 0)
	envVars := make(map[string]interface{})

	for n, v := range anns {
		args := annotations.HandlerArgs{Client: client, Name: n, Object: obj, Value: v}
		h, err := annotations.BuildHandler(args)
		if err != nil {
			if err == annotations.InvalidAnnotationPrefixErr {
				continue
			}
			return nil, err
		}
		r, err := h.Handle()
		if err != nil {
			continue
		}

		err = mergo.Merge(&envVars, r.Object, mergo.WithAppendSlice, mergo.WithOverride)
		if err != nil {
			return nil, err
		}

		if r.Type == annotations.BindingTypeVolumeMount {
			volumeKeys = append(volumeKeys, r.Path)
		}
	}

	serviceCtx := &ServiceContext{
		Object:     obj,
		EnvVars:    envVars,
		VolumeKeys: volumeKeys,
	}

	return serviceCtx, nil
}
