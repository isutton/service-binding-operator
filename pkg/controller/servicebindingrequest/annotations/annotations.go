package annotations

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
)

// Handler produces a unstructured object produced from the strategy encoded in
// an annotation value.
type Handler interface {
	Handle() (map[string]interface{}, error)
}

// HandlerArgs are arguments that can be used by action constructors to perform
// its task.
type HandlerArgs struct {
	// Resource is the owner resource unstructured representation.
	Resource map[string]interface{}
	// Name is the annotation key, with prefix included.
	Name string
	// Value is the annotation value.
	Value string
}

// BuildHandler attempts to create an action for the given args.
func BuildHandler(args HandlerArgs) (Handler, error) {
	bindingInfo, err := bindinginfo.NewBindingInfo(args.Name, args.Value)
	if err != nil {
		return nil, err
	}
	switch bindingInfo.Value {
	case AttributeValue:
		return NewAttributeHandler(bindingInfo, args.Resource), nil
	default:
		panic("not implemented")
	}
}
