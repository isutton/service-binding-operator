package servicebindingrequest

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// GetCRs returns all collected service resources.
func (p *Plan) GetCRs() []*unstructured.Unstructured {
	return p.ServiceContexts.GetCRs()
}

// GetServiceContexts returns all collected service contexts.
func (p *Plan) GetServiceContexts() ServiceContexts {
	return p.ServiceContexts
}
