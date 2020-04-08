package annotations

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

type BindingType string

const (
	BindingTypeVolumeMount BindingType = "volumemount"
	BindingTypeEnvVar      BindingType = "env"
)

type Value struct {
	Result   map[string]interface{}
	BindType string
}

// Handler produces a unstructured object produced from the strategy encoded in
// an annotation value.
type Handler interface {
	Handle() (Value, error)
}

// HandlerArgs are arguments that can be used by action constructors to perform
// its task.
type HandlerArgs struct {
	// Resource is the owner resource unstructured representation.
	Resource unstructured.Unstructured
	// Name is the annotation key, with prefix included.
	Name string
	// Value is the annotation value.
	Value string

	Client dynamic.Interface
}

// BuildHandler attempts to create an action for the given args.
func BuildHandler(args HandlerArgs) (Handler, error) {
	bindingInfo, err := bindinginfo.NewBindingInfo(args.Name, args.Value)
	if err != nil {
		return nil, err
	}

	val := bindingInfo.Value

	switch {
	case IsAttribute(val):
		return NewAttributeHandler(bindingInfo, args.Resource), nil
	case IsSecret(val):
		return NewSecretHandler(args.Client, bindingInfo, args.Resource)
	case IsConfigMap(val):
		return NewConfigMapHandler(args.Client, bindingInfo, args.Resource)
	default:
		panic("not implemented")
	}
}
