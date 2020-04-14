package annotations

import (
	"encoding/base64"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
)

const SecretValue = "binding:env:object:secret"

// IsSecret returns true if the annotation value should trigger the secret handler.
func IsSecret(s string) bool {
	return SecretValue == s
}

// decodeBase64String asserts whether val is a string and returns its decoded value.
func decodeBase64String(v interface{}) (string, error) {
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

// NewSecretHandler constructs a SecretHandler.
func NewSecretHandler(
	client dynamic.Interface,
	bindingInfo *bindinginfo.BindingInfo,
	resource unstructured.Unstructured,
) (Handler, error) {
	h, err := NewResourceHandler(
		client,
		bindingInfo,
		resource,
		schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		},
		&dataPath,
	)
	if err == nil {
		h.valueDecoder = decodeBase64String
	}
	return h, err
}
