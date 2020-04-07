package annotations

import (
	"encoding/base64"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
)

const SecretValue = "binding:env:object:secret"

// SecretHandler handles "binding:env:object:secret" annotations.
type SecretHandler struct {
	// client is the client used to retrieve a related secret.
	client dynamic.Interface
	// valuePath is the path that should be extracted from the secret.
	valuePath string
	// secretNamePath is the path the secret can be found in the resource.
	secretNamePath string
	// outputPath is the path the extracted value will be placed under.
	outputPath string
	// resource is the unstructured object to extract data using inputPath.
	resource unstructured.Unstructured
}

// discoverSecretName returns the secret name referenced by the handler. Can return an error in the
// case the expected information doesn't exist in the handler's resource object.
func (h *SecretHandler) discoverSecretName() (string, error) {
	secretNameValue, ok, err := nested.GetValueFromMap(h.resource.Object, h.secretNamePath)
	if !ok {
		return "", SecretNameFieldNotFoundErr
	}
	if err != nil {
		return "", err
	}
	name, ok := secretNameValue.(string)
	if !ok {
		return "", InvalidArgumentErr(h.secretNamePath)
	}
	return name, nil
}

// base64ToString asserts whether val is a string and returns its decoded value.
func base64ToString(v interface{}) (string, error) {
	stringVal, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("should be a string")
	}
	decodedVal, err := base64.StdEncoding.DecodeString(stringVal)
	if err != nil {
		return "", err
	}
	return string(decodedVal), nil
}

// Handle returns a unstructured object according to the "binding:env:object:secret" annotation
// strategy.
func (h *SecretHandler) Handle() (map[string]interface{}, error) {
	name, err := h.discoverSecretName()
	if err != nil {
		return nil, err
	}

	ns := h.resource.GetNamespace()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}

	secret, err := h.client.Resource(gvr).Namespace(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	val, ok, err := nested.GetValueFromMap(secret.Object, h.valuePath)
	if !ok {
		return nil, InvalidArgumentErr(h.valuePath)
	}
	if err != nil {
		return nil, err
	}

	if mapVal, ok := val.(map[string]interface{}); ok {
		tmpVal := make(map[string]interface{})
		for k, v := range mapVal {
			decodedVal, err := base64ToString(v)
			if err != nil {
				return nil, err
			}
			tmpVal[k] = string(decodedVal)
		}
		val = tmpVal
	} else {
		val, err = base64ToString(val)
		if err != nil {
			return nil, err
		}
	}

	return nested.ComposeValue(val, nested.NewPath(h.outputPath)), nil
}

// IsSecret returns true if the annotation value should trigger the secret handler.
func IsSecret(s string) bool {
	return SecretValue == s
}

// NewSecretHandler constructs a SecretHandler.
func NewSecretHandler(
	client dynamic.Interface,
	bindingInfo *bindinginfo.BindingInfo,
	resource unstructured.Unstructured,
) (*SecretHandler, error) {
	if client == nil {
		return nil, InvalidArgumentErr("client")
	}

	if bindingInfo == nil {
		return nil, InvalidArgumentErr("bindingInfo")
	}

	if len(bindingInfo.Path) == 0 {
		return nil, InvalidArgumentErr("bindingInfo.Path")
	}

	secretNamePath := bindingInfo.FieldPath
	outputPath := secretNamePath
	valuePath := "data"

	if len(bindingInfo.Path) > 0 && bindingInfo.FieldPath != bindingInfo.Path {
		valuePath = bindingInfo.Path
		outputPath = outputPath + "." + valuePath
	}

	return &SecretHandler{
		client:         client,
		valuePath:      valuePath,
		secretNamePath: secretNamePath,
		outputPath:     outputPath,
		resource:       resource,
	}, nil
}
