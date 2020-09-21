package binding

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
)

type AnnotationToDefinitionMapperOptions interface {
	DefinitionMapperOptions
	GetValue() string
	GetName() string
}

type annotationToDefinitionMapperOptions struct {
	name  string
	value string
}

func (o *annotationToDefinitionMapperOptions) GetName() string  { return o.name }
func (o *annotationToDefinitionMapperOptions) GetValue() string { return o.value }

func NewAnnotationMapperOptions(name, value string) AnnotationToDefinitionMapperOptions {
	return &annotationToDefinitionMapperOptions{name: name, value: value}
}

type AnnotationToDefinitionMapper struct {
	KubeClient dynamic.Interface
}

var _ DefinitionMapper = (*AnnotationToDefinitionMapper)(nil)

type modelKey string

const (
	pathModelKey        modelKey = "path"
	objectTypeModelKey  modelKey = "objectType"
	sourceKeyModelKey   modelKey = "sourceKey"
	sourceValueModelKey modelKey = "sourceValue"
	elementTypeModelKey modelKey = "elementType"
)

const AnnotationPrefix = "service.binding"

func parseName(name string) (string, error) {
	// bail out in the case the annotation name doesn't start with "service.binding"
	if name != AnnotationPrefix && !strings.HasPrefix(name, AnnotationPrefix+"/") {
		return "", fmt.Errorf("can't process annotation with name %q", name)
	}

	if p := strings.SplitN(name, "/", 2); len(p) > 1 && len(p[1]) > 0 {
		return p[1], nil
	}

	return "", nil
}

func (m *AnnotationToDefinitionMapper) Map(mapperOpts DefinitionMapperOptions) (Definition, error) {
	opts, ok := mapperOpts.(AnnotationToDefinitionMapperOptions)
	if !ok {
		return nil, fmt.Errorf("provide an AnnotationToDefinitionMapperOptions")
	}

	outputName, err := parseName(opts.GetName())
	if err != nil {
		return nil, err
	}

	mod, err := newModel(opts.GetValue())
	if err != nil {
		return nil, errors.Wrapf(err, "could not create binding model for annotation key %q",
			opts.GetName())
	}

	if len(outputName) == 0 {
		outputName = mod.path[len(mod.path)-1]
	}

	switch {
	case mod.isStringElementType() && mod.isStringObjectType():
		return &stringDefinition{
			outputName: outputName,
			path:       mod.path,
		}, nil

	case mod.isStringElementType() && mod.hasDataField():
		return &stringFromDataFieldDefinition{
			kubeClient: m.KubeClient,
			objectType: mod.objectType,
			outputName: outputName,
			path:       mod.path,
			sourceKey:  mod.sourceKey,
		}, nil

	case mod.isMapElementType() && mod.hasDataField():
		return &mapFromDataFieldDefinition{
			kubeClient:  m.KubeClient,
			objectType:  mod.objectType,
			outputName:  outputName,
			path:        mod.path,
			sourceValue: mod.sourceValue,
		}, nil

	case mod.isMapElementType() && mod.isStringObjectType():
		return &stringOfMapDefinition{
			outputName: outputName,
			path:       mod.path,
		}, nil

	case mod.isSliceOfMapsElementType():
		return &sliceOfMapsFromPathDefinition{
			outputName:  outputName,
			path:        mod.path,
			sourceKey:   mod.sourceKey,
			sourceValue: mod.sourceValue,
		}, nil

	case mod.isSliceOfStringsElementType():
		return &sliceOfStringsFromPathDefinition{
			outputName:  outputName,
			path:        mod.path,
			sourceValue: mod.sourceValue,
		}, nil
	}

	panic(fmt.Sprintf("not implemented: %s", opts))
}
