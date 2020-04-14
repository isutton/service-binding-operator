package annotations

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// bindingType encodes the medium the binding should deliver the configuration value.
type bindingType string

const (
	// BindingTypeVolumeMount indicates the binding should happen through a volume mount.
	BindingTypeVolumeMount bindingType = "volumemount"
	// BindingTypeEnvVar indicates the binding should happen through environment variables.
	BindingTypeEnvVar bindingType = "env"
)

// supportedBindingTypes contains all currently supported binding types.
var supportedBindingTypes = map[bindingType]bool{
	BindingTypeVolumeMount: true,
	BindingTypeEnvVar:      true,
}

// Result contains meta-information regarding the result of Handle().
type Result struct {
	Object map[string]interface{}
	Type   bindingType
}

// Handler produces a unstructured object produced from the strategy encoded in
// an annotation value.
type Handler interface {
	Handle() (Result, error)
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
