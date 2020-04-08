package annotations

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
)

// ResourceHandler handles annotations related to external resources.
type ResourceHandler struct {
	// client is the client used to retrieve a related secret.
	client dynamic.Interface
	// valuePath is the path that should be extracted from the secret.
	valuePath string
	// relatedNamePath is the path the related resource name can be found in the resource.
	relatedNamePath string
	// relatedGroupVersionResource is the related resource GVR, used to retrieve the related resource
	// using the client.
	relatedGroupVersionResource schema.GroupVersionResource
	// outputPath is the path the extracted value will be placed under.
	outputPath string
	// resource is the unstructured object to extract data using inputPath.
	resource unstructured.Unstructured

	valueDecoder func(interface{}) (string, error)
}

// discoverRelatedResourceName returns the resource name referenced by the handler. Can return an
// error in the case the expected information doesn't exist in the handler's resource object.
func (h *ResourceHandler) discoverRelatedResourceName() (string, error) {
	resourceNameValue, ok, err := nested.GetValueFromMap(h.resource.Object, h.relatedNamePath)
	if !ok {
		return "", ResourceNameFieldNotFoundErr
	}
	if err != nil {
		return "", err
	}
	name, ok := resourceNameValue.(string)
	if !ok {
		return "", InvalidArgumentErr(h.relatedNamePath)
	}
	return name, nil
}

// Handle returns a unstructured object according to the "binding:env:object:secret" annotation
// strategy.
func (h *ResourceHandler) Handle() (map[string]interface{}, error) {
	name, err := h.discoverRelatedResourceName()
	if err != nil {
		return nil, err
	}

	ns := h.resource.GetNamespace()
	resource, err := h.client.Resource(h.relatedGroupVersionResource).Namespace(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	val, ok, err := nested.GetValueFromMap(resource.Object, h.valuePath)
	if !ok {
		return nil, InvalidArgumentErr(h.valuePath)
	}
	if err != nil {
		return nil, err
	}

	if mapVal, ok := val.(map[string]interface{}); ok {
		tmpVal := make(map[string]interface{})
		for k, v := range mapVal {
			decodedVal, err := h.valueDecoder(v)
			if err != nil {
				return nil, err
			}
			tmpVal[k] = decodedVal
		}
		val = tmpVal
	} else {
		val, err = h.valueDecoder(val)
		if err != nil {
			return nil, err
		}
	}

	return nested.ComposeValue(val, nested.NewPath(h.outputPath)), nil
}

// NewSecretHandler constructs a SecretHandler.
func NewResourceHandler(
	client dynamic.Interface,
	bindingInfo *bindinginfo.BindingInfo,
	resource unstructured.Unstructured,
	relatedGroupVersionResource schema.GroupVersionResource,
) (*ResourceHandler, error) {
	if client == nil {
		return nil, InvalidArgumentErr("client")
	}

	if bindingInfo == nil {
		return nil, InvalidArgumentErr("bindingInfo")
	}

	if len(bindingInfo.Path) == 0 {
		return nil, InvalidArgumentErr("bindingInfo.Path")
	}

	relatedNamePath := bindingInfo.FieldPath
	outputPath := relatedNamePath
	valuePath := "data"

	if len(bindingInfo.Path) > 0 && bindingInfo.FieldPath != bindingInfo.Path {
		valuePath = bindingInfo.Path
		outputPath = outputPath + "." + valuePath
	}

	return &ResourceHandler{
		client:                      client,
		valuePath:                   valuePath,
		relatedNamePath:             relatedNamePath,
		outputPath:                  outputPath,
		resource:                    resource,
		relatedGroupVersionResource: relatedGroupVersionResource,
		valueDecoder: func(v interface{}) (string, error) {
			s, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("value is not a string")
			}
			return s, nil
		},
	}, nil
}
