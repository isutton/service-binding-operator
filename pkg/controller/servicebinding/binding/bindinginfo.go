package binding

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// ServiceBindingOperatorAnnotationPrefix is the prefix of Service Binding Operator related annotations.
	ServiceBindingOperatorAnnotationPrefix = "servicebindingoperator.redhat.io/"
)

// bindingInfo represents the pieces of a binding as parsed from an annotation.
type bindingInfo struct {
	// ResourceReferencePath is the field in the Backing Service CR referring to a bindable property, either
	// embedded or a reference to a related object..
	ResourceReferencePath string
	// SourcePath is the field that will be collected from the Backing Service CR or a related object.
	SourcePath string
	// Descriptor is the field reference to another manifest.
	Descriptor string
	// Value is the original annotation value.
	Value string
}

type ErrInvalidAnnotationPrefix string

func (e ErrInvalidAnnotationPrefix) Error() string {
	return fmt.Sprintf("invalid annotation prefix: %s", string(e))
}

func IsErrInvalidAnnotationPrefix(err error) bool {
	_, ok := err.(ErrInvalidAnnotationPrefix)
	return ok
}

var ErrInvalidAnnotationName = errors.New("invalid annotation name")

type ErrEmptyAnnotationName string

func (e ErrEmptyAnnotationName) Error() string {
	return fmt.Sprintf("empty annotation name: %s", string(e))
}

func IsErrEmptyAnnotationName(err error) bool {
	_, ok := err.(ErrEmptyAnnotationName)
	return ok
}

// NewBindingInfo parses the encoded in the annotation name, returning its parts.
func NewBindingInfo(name string, value string) (*bindingInfo, error) {
	// do not process unknown annotations
	if !strings.HasPrefix(name, ServiceBindingOperatorAnnotationPrefix) {
		return nil, ErrInvalidAnnotationPrefix(name)
	}

	cleanName := strings.TrimPrefix(name, ServiceBindingOperatorAnnotationPrefix)
	if len(cleanName) == 0 {
		return nil, ErrEmptyAnnotationName(cleanName)
	}
	parts := strings.SplitN(cleanName, "-", 2)

	resourceReferencePath := parts[0]
	// sourcePath := parts[0]
	sourcePath := ""
	descriptor := ""
	// the annotation is a reference to another manifest
	if len(parts) == 2 {
		sourcePath = parts[1]
		descriptor = strings.Join([]string{value, sourcePath}, ":")
	} else {
		descriptor = strings.Join([]string{value, resourceReferencePath}, ":")
	}

	return &bindingInfo{
		ResourceReferencePath: resourceReferencePath,
		SourcePath:            sourcePath,
		Descriptor:            descriptor,
		Value:                 value,
	}, nil
}
