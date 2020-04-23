package annotations

import (
	"fmt"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
)

// ResourceHandler handles annotations related to external resources.
type ResourceHandler struct {
	// bindingInfo contains the binding details related to the annotation handler.
	bindingInfo *BindingInfo
	// client is the client used to retrieve a related secret.
	client dynamic.Interface
	// valuePath is the path that should be extracted from the secret.
	valuePath []string
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
	resourceNameValue, ok, err := nested.GetValueFromMap(
		h.resource.Object,
		strings.Split(h.relatedNamePath, ".")...,
	)
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

// discoverBindingType attempts to extract a binding type from the given annotation value val.
func discoverBindingType(val string) (bindingType, error) {
	re := regexp.MustCompile("^binding:(.*?):.*$")
	parts := re.FindStringSubmatch(val)
	if len(parts) == 0 {
		return "", fmt.Errorf("error extracting binding type")
	}
	t := bindingType(parts[1])
	_, ok := supportedBindingTypes[t]
	if !ok {
		return "", UnknownBindingTypeErr(t)
	}
	return t, nil
}

// Handle returns the value for an external resource strategy.
func (h *ResourceHandler) Handle() (Result, error) {
	name, err := h.discoverRelatedResourceName()
	if err != nil {
		return Result{}, err
	}

	ns := h.resource.GetNamespace()
	resource, err := h.client.Resource(h.relatedGroupVersionResource).Namespace(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return Result{}, fmt.Errorf("error handling annotation: %w", err)
	}

	val, ok, err := nested.GetValueFromMap(resource.Object, h.valuePath...)
	if !ok {
		return Result{}, InvalidArgumentErr(strings.Join(h.valuePath, ", "))
	}
	if err != nil {
		return Result{}, err
	}

	if mapVal, ok := val.(map[string]interface{}); ok {
		tmpVal := make(map[string]interface{})
		for k, v := range mapVal {
			decodedVal, err := h.valueDecoder(v)
			if err != nil {
				return Result{}, err
			}
			tmpVal[k] = decodedVal
		}
		val = tmpVal
	} else {
		val, err = h.valueDecoder(val)
		if err != nil {
			return Result{}, err
		}
	}

	typ, err := discoverBindingType(h.bindingInfo.Value)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Object: nested.ComposeValue(val, nested.NewPath(h.outputPath)),
		Type:   typ,
		Path:   h.outputPath,
	}, nil
}

// NewSecretHandler constructs a SecretHandler.
func NewResourceHandler(
	client dynamic.Interface,
	bindingInfo *BindingInfo,
	resource unstructured.Unstructured,
	relatedGroupVersionResource schema.GroupVersionResource,
	valuePathPrefix *string,
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

	valuePath := []string{}

	if len(bindingInfo.Path) > 0 && bindingInfo.FieldPath != bindingInfo.Path {
		valuePath = append(valuePath, bindingInfo.Path)
		outputPath = outputPath + "." + bindingInfo.Path
		if valuePathPrefix != nil && len(*valuePathPrefix) > 0 {
			valuePath = append([]string{*valuePathPrefix}, valuePath...)
		}
	} else if valuePathPrefix != nil && len(*valuePathPrefix) > 0 {
		valuePath = []string{*valuePathPrefix}
	}

	return &ResourceHandler{
		bindingInfo:                 bindingInfo,
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
