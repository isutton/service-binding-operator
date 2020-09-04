package annotations

import (
	"fmt"
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

// dataPath is the path ConfigMap and Secret resources use to scope their data.
//
// note: it is currently used to provide a pointer to the "data" string, which is the location
// ConfigMap and Secret resources keep user data.
var dataPath = "data"

// result contains data that has been collected by an annotation handler.
type result struct {
	// Data contains the annotation data collected by an annotation handler inside a deep structure
	// with its root being the value specified in the Path field.
	Data map[string]interface{}
	// Type indicates where the Object field should be injected in the application; can be either
	// "env" or "volumemount".
	Type bindingType
	// Path is the nested location the collected data can be found in the Data field.
	Path string
	// RawData contains the annotation data collected by an annotation handler
	// inside a deep structure with its root being composed by the path where
	// the external resource name was extracted and the path within the external
	// resource.
	RawData map[string]interface{}
}

// handler should be implemented by types that want to offer a mechanism to provide binding data to
// the system.
type handler interface {
	// Handle returns binding data.
	Handle() (result, error)
}

type errHandlerNotFound string

func (e errHandlerNotFound) Error() string {
	return fmt.Sprintf("could not find handler for annotation value %q", string(e))
}

func IsErrHandlerNotFound(err error) bool {
	_, ok := err.(errHandlerNotFound)
	return ok
}
