package annotations

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const AttributeValue = "binding:env:attribute"

// AttributeHandler handles "binding:env:attribute" annotations.
type AttributeHandler struct {
	// inputPath is the path that should be extracted from the resource.
	inputPath string
	// outputPath is the path the extracted data should be placed under in the
	// resulting unstructured object in Handler.
	outputPath string
	// resource is the unstructured object to extract data using inputPath.
	resource unstructured.Unstructured
}

// OutputPath returns the path the extracted data should be placed under.
func (a *AttributeHandler) OutputPath() string {
	if len(a.outputPath) > 0 {
		return a.outputPath
	}
	return a.inputPath
}

// Handle returns a unstructured object according to the "binding:env:attribute"
// annotation strategy.
func (h *AttributeHandler) Handle() (Result, error) {
	val, _, err := nested.GetValue(h.resource.Object, h.inputPath, h.OutputPath())
	if err != nil {
		return Result{}, err
	}
	return Result{
		Object: val,
	}, nil
}

// IsAttribute returns true if the annotation value should trigger the attribute
// handler.
func IsAttribute(s string) bool {
	return AttributeValue == s
}

// NewAttributeHandler constructs an AttributeHandler.
func NewAttributeHandler(
	bindingInfo *BindingInfo,
	resource unstructured.Unstructured,
) *AttributeHandler {
	return &AttributeHandler{
		inputPath:  bindingInfo.SourcePath,
		outputPath: bindingInfo.ResourceReferencePath,
		resource:   resource,
	}
}
