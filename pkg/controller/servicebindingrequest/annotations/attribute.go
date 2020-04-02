package annotations

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
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
	resource map[string]interface{}
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
func (h *AttributeHandler) Handle() (map[string]interface{}, error) {
	val, _, err := nested.GetValue(h.resource, h.inputPath, h.OutputPath())
	if err != nil {
		return nil, err
	}
	return val, nil
}

// NewAttributeHandler constructs an AttributeHandler.
func NewAttributeHandler(
	bindingInfo *bindinginfo.BindingInfo,
	resource map[string]interface{},
) *AttributeHandler {
	return &AttributeHandler{
		inputPath:  bindingInfo.Path,
		outputPath: bindingInfo.FieldPath,
		resource:   resource,
	}
}
