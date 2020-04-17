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
	// CR is the resource being used as reference.
	CR *unstructured.Unstructured
	// EnvVars is the composition of all collected data for the reference CR.
	EnvVars map[string]interface{}
	// VolumeKeys contains the keys that should be mounted as volume from the binding secret.
	VolumeKeys   []string
	EnvVarPrefix *string
}

// ServiceContexts contains a collection of service contexts.
type ServiceContexts []*ServiceContext

// GetCRs returns a slice of unstructured CRs contained in the collection.
func (rr ServiceContexts) GetCRs() []*unstructured.Unstructured {
	var crs []*unstructured.Unstructured
	for _, r := range rr {
		crs = append(crs, r.CR)
	}
	return crs
}
