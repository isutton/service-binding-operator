package servicebindingrequest

import (
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ServiceContext contains information related to a service.
type ServiceContext struct {
	// CRDDescription is the description of the resources CRD, either built from a manifest from the
	// cluster or composed through annotations in the CRD.
	CRDDescription *v1alpha1.CRDDescription
	// Object is the resource being used as reference.
	Object *unstructured.Unstructured
	// EnvVars contains the service's contributed environment variables.
	EnvVars map[string]interface{}
	// VolumeKeys contains the keys that should be mounted as volume from the binding secret.
	VolumeKeys   []string
	EnvVarPrefix *string
}

// ServiceContexts contains a collection of service contexts.
type ServiceContexts []*ServiceContext

// GetObjects returns a slice of service unstructured objects contained in the collection.
func (sc ServiceContexts) GetObjects() []*unstructured.Unstructured {
	var crs []*unstructured.Unstructured
	for _, s := range sc {
		crs = append(crs, s.Object)
	}
	return crs
}
