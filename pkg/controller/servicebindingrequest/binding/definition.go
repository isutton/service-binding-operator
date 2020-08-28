package binding

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	// configMapObjectType indicates the path contains a name for a ConfigMap containing the binding
	// data.
	configMapObjectType objectType = "ConfigMap"
	// secretObjectType indicates the path contains a name for a Secret containing the binding data.
	secretObjectType objectType = "Secret"
	// stringObjectType indicates the path contains a value string.
	stringObjectType objectType = "string"
	// emptyObjectType is used as default value when the objectType key is present in the string
	// provided by the user but no value has been provided; can be used by the user to force the
	// system to use the default objectType.
	emptyObjectType objectType = ""

	// mapElementType indicates the value found at path is a map[string]interface{}.
	mapElementType elementType = "map"
	// sliceOfMapsElementType indicates the value found at path is a slice of maps.
	sliceOfMapsElementType elementType = "sliceOfMaps"
	// sliceOfStringsElementType indicates the value found at path is a slice of strings.
	sliceOfStringsElementType elementType = "sliceOfStrings"
	// stringElementType indicates the value found at path is a string.
	stringElementType elementType = "string"
)

type objectType string

type elementType string

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type Definition interface {
	Apply(u *unstructured.Unstructured) (Value, error)
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type DefinitionMapperOptions interface{}

type DefinitionMapper interface {
	Map(DefinitionMapperOptions) (Definition, error)
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

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

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type annotationToDefinitionMapper struct {
	kubeClient dynamic.Interface
}

var _ DefinitionMapper = (*annotationToDefinitionMapper)(nil)

type modelKey string

const (
	pathModelKey        modelKey = "path"
	objectTypeModelKey  modelKey = "objectType"
	sourceKeyModelKey   modelKey = "sourceKey"
	sourceValueModelKey modelKey = "sourceValue"
	elementTypeModelKey modelKey = "elementType"
)

const annotationPrefix = "service.binding"

func (m *annotationToDefinitionMapper) Map(mapperOpts DefinitionMapperOptions) (Definition, error) {
	opts, ok := mapperOpts.(AnnotationToDefinitionMapperOptions)
	if !ok {
		return nil, fmt.Errorf("provide an AnnotationToDefinitionMapperOptions")
	}

	name := opts.GetName()

	// bail out in the case the annotation name doesn't start with "service.binding"
	if name != annotationPrefix && !strings.HasPrefix(name, annotationPrefix+"/") {
		return nil, fmt.Errorf("can't process annotation with name %q", name)
	}

	outputName := ""
	if p := strings.SplitN(name, "/", 2); len(p) > 1 && len(p[1]) > 0 {
		outputName = p[1]
	}

	// re contains a regular expression to split the input string using '=' and ',' as separators
	re := regexp.MustCompile("[=,]")

	// split holds the tokens extracted from the input string
	split := re.Split(opts.GetValue(), -1)

	// its length should be even, since from this point on is assumed a sequence of key and value
	// pairs as model source
	if len(split)%2 != 0 {
		m := fmt.Sprintf("invalid input, odd number of tokens: %q", split)
		return nil, errors.New(m)
	}

	// extract the tokens into a map, iterating a pair at a time and using the Nth element as key and
	// Nth+1 as value
	raw := make(map[modelKey]string)
	for i := 0; i < len(split); i += 2 {
		k := modelKey(split[i])
		// invalid object type can be created here e.g. "foobar"; this does not pose a problem since
		// the value will be used in a switch statement further on
		v := split[i+1]
		raw[k] = v
	}

	// assert PathModelKey is present
	path, found := raw[pathModelKey]
	if !found {
		return nil, fmt.Errorf("path not found: '%s: %s'", name, opts.GetValue())
	}

	// ensure ObjectTypeModelKey has a default value
	var objType objectType
	if rawObjectType, found := raw[objectTypeModelKey]; !found {
		objType = stringObjectType
	} else {
		// in the case the key is present but the value isn't (for example, "objectType=,") the
		// default string object type should be set
		if objType = objectType(rawObjectType); objType == emptyObjectType {
			objType = stringObjectType
		}
	}

	// ensure sourceKey has a default value
	sourceKey, found := raw[sourceKeyModelKey]
	if !found {
		sourceKey = ""
	}

	// hasData indicates the configured or inferred objectType is either a Secret or ConfigMap
	hasData := (objType == secretObjectType || objType == configMapObjectType)
	// hasSourceKey indicates a value for sourceKey has been informed
	hasSourceKey := len(sourceKey) > 0

	var eltType elementType
	if rawEltType, found := raw[elementTypeModelKey]; found {
		// the input string contains an elementType configuration, use it
		eltType = elementType(rawEltType)
	} else if hasData && !hasSourceKey {
		// the input doesn't contain an elementType configuration, does contain a sourceKey
		// configuration, and is either a Secret or ConfigMap
		eltType = mapElementType
	} else {
		// elementType configuration hasn't been informed and there's no extra hints, assume it is a
		// string element
		eltType = stringElementType
	}

	// ensure SourceValueModelKey has a default value
	sourceValue, found := raw[sourceValueModelKey]
	if !found {
		sourceValue = ""
	}

	// ensure an error is returned if not all required information is available for sliceOfMaps
	// element type
	if eltType == sliceOfMapsElementType && (len(sourceValue) == 0 || len(sourceKey) == 0) {
		return nil, errors.New("sliceOfMaps elementType requires sourceKey and sourceValue to be present")
	}

	isStringElementType := eltType == stringElementType
	isStringObjectType := objType == stringObjectType
	isMapElementType := eltType == mapElementType
	isSliceOfMapsElementType := eltType == sliceOfMapsElementType
	isSliceOfStringsElementType := eltType == sliceOfStringsElementType
	hasDataField := (objType == secretObjectType || objType == configMapObjectType)

	path = strings.Trim(path, "{}.")
	pathParts := strings.Split(path, ".")

	if len(outputName) == 0 {
		outputName = pathParts[len(pathParts)-1]
	}

	switch {
	case isStringElementType && isStringObjectType:
		return &stringDefinition{
			outputName: outputName,
			path:       pathParts,
		}, nil

	case isStringElementType && hasDataField:
		return &stringFromDataFieldDefinition{
			kubeClient: m.kubeClient,
			objectType: objType,
			outputName: outputName,
			path:       pathParts,
			sourceKey:  sourceKey,
		}, nil

	case isMapElementType && hasDataField:
		return &mapFromDataFieldDefinition{
			kubeClient: m.kubeClient,
			objectType: objType,
			outputName: outputName,
			path:       pathParts,
		}, nil

	case isMapElementType && isStringObjectType:
		return &stringOfMapDefinition{
			outputName: outputName,
			path:       pathParts,
		}, nil

	case isSliceOfMapsElementType:
		return &sliceOfMapsFromPathDefinition{
			outputName:  outputName,
			path:        pathParts,
			sourceKey:   sourceKey,
			sourceValue: sourceValue,
		}, nil

	case isSliceOfStringsElementType:
		return &sliceOfStringsFromPathDefinition{
			outputName:  outputName,
			path:        pathParts,
			sourceValue: sourceValue,
		}, nil
	}

	panic("not implemented")
}

type stringDefinition struct {
	outputName string
	path       []string
}

var _ Definition = (*stringDefinition)(nil)

func (d *stringDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	val, ok, err := unstructured.NestedFieldCopy(u.Object, d.path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	outputName := d.outputName
	if len(outputName) == 0 {
		outputName = d.path[len(d.path)-1]
	}

	m := map[string]interface{}{
		outputName: fmt.Sprintf("%v", val),
	}
	return &value{v: m}, nil
}

type stringFromDataFieldDefinition struct {
	kubeClient dynamic.Interface
	objectType objectType
	outputName string
	path       []string
	sourceKey  string
}

var _ Definition = (*stringFromDataFieldDefinition)(nil)

func (d *stringFromDataFieldDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	if d.kubeClient == nil {
		return nil, errors.New("kubeClient required for this functionality")
	}

	var resource schema.GroupVersionResource
	if d.objectType == secretObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	} else if d.objectType == configMapObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	}

	resourceName, ok, err := unstructured.NestedString(u.Object, d.path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	otherObj, err := d.kubeClient.Resource(resource).Namespace(u.GetNamespace()).Get(resourceName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	val, ok, err := unstructured.NestedString(otherObj.Object, "data", d.sourceKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}
	if d.objectType == secretObjectType {
		n, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return nil, err
		}
		val = string(n)
	}
	v := map[string]interface{}{
		"": val,
	}
	return &value{v: v}, nil
}

type mapFromDataFieldDefinition struct {
	kubeClient dynamic.Interface
	objectType objectType
	outputName string
	path       []string
}

var _ Definition = (*mapFromDataFieldDefinition)(nil)

func (d *mapFromDataFieldDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	if d.kubeClient == nil {
		return nil, errors.New("kubeClient required for this functionality")
	}

	var resource schema.GroupVersionResource
	if d.objectType == secretObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	} else if d.objectType == configMapObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	}

	resourceName, ok, err := unstructured.NestedString(u.Object, d.path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	otherObj, err := d.kubeClient.Resource(resource).Namespace(u.GetNamespace()).
		Get(resourceName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	val, ok, err := unstructured.NestedStringMap(otherObj.Object, "data")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}
	if d.objectType == secretObjectType {
		for k, v := range val {
			n, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return nil, err
			}
			val[k] = string(n)
		}
	}

	return &value{v: val}, nil

}

type stringOfMapDefinition struct {
	outputName string
	path       []string
}

var _ Definition = (*stringOfMapDefinition)(nil)

func (d *stringOfMapDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	val, ok, err := unstructured.NestedFieldNoCopy(u.Object, d.path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	outputName := d.outputName
	if len(outputName) == 0 {
		outputName = d.path[len(d.path)-1]
	}
	v := map[string]interface{}{
		outputName: val,
	}
	return &value{v: v}, nil

}

type sliceOfMapsFromPathDefinition struct {
	outputName  string
	path        []string
	sourceKey   string
	sourceValue string
}

var _ Definition = (*sliceOfMapsFromPathDefinition)(nil)

func (d *sliceOfMapsFromPathDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	val, ok, err := unstructured.NestedSlice(u.Object, d.path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	v := make(map[string]interface{})
	for _, e := range val {
		if mm, ok := e.(map[string]interface{}); ok {
			key := mm[d.sourceKey]
			ks := key.(string)
			value := mm[d.sourceValue]
			v[ks] = value
		}
	}

	return &value{v: v}, nil
}

type sliceOfStringsFromPathDefinition struct {
	outputName  string
	path        []string
	sourceValue string
}

var _ Definition = (*sliceOfStringsFromPathDefinition)(nil)

func (d *sliceOfStringsFromPathDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	val, ok, err := unstructured.NestedSlice(u.Object, d.path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	v := make([]string, 0, len(val))
	for _, e := range val {
		if mm, ok := e.(map[string]interface{}); ok {
			sourceValue := mm[d.sourceValue].(string)
			v = append(v, sourceValue)
		}
	}

	return &value{v: v}, nil
}
