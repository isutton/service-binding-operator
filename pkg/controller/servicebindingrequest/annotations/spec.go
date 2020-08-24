package annotations

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type objectType string

type elementType string

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

type modelKey string

const (
	pathModelKey        modelKey = "path"
	objectTypeModelKey  modelKey = "objectType"
	sourceKeyModelKey   modelKey = "sourceKey"
	sourceValueModelKey modelKey = "sourceValue"
	elementTypeModelKey modelKey = "elementType"
)

type bindingDefinition struct {
	objectType objectType
	// path is a template represention of the path to an element in a Kubernetes resource. The
	// value of path is specified as JSONPath. Required.
	path string
	// elementType specifies the type of object in an array selected by the path. One of sliceOfMaps,
	// sliceOfStrings, string (default).
	elementType elementType
	// sourceKey specifies a particular key to select if a ConfigMap or Secret is selected by the
	// path. Specifies a value to use for the key for an entry in a binding Secret when elementType
	// is sliceOfMaps.
	sourceKey string
	// sourceValue specifies a particular value to use for the value for an entry in a binding Secret
	// when elementType is sliceOfMaps
	sourceValue string
}

func newBindingDefinitionFromAnnotation(in string) (*bindingDefinition, error) {
	// re contains a regular expression to split the input string using '=' and ',' as separators
	re := regexp.MustCompile("[=,]")

	// split holds the tokens extracted from the input string
	split := re.Split(in, -1)

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
		return nil, errors.New("path not found: " + in)
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

	return &bindingDefinition{
		path:        path,
		objectType:  objType,
		elementType: eltType,
		sourceKey:   sourceKey,
		sourceValue: sourceValue,
	}, nil
}

func (m *bindingDefinition) getPath() []string {
	p := strings.Trim(m.path, "{}.")
	return strings.Split(p, ".")
}

func produceStringValue(obj map[string]interface{}, path []string) (string, error) {
	val, ok, err := unstructured.NestedFieldCopy(obj, path...)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("not found")
	}
	return fmt.Sprintf("%v", val), nil
}

func produceStringOfMapValue(obj map[string]interface{}, path []string) (map[string]string, error) {
	val, ok, err := unstructured.NestedStringMap(obj, path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}
	return val, nil
}

func produceStringValueFromDataField(
	kubeClient dynamic.Interface,
	ns string,
	objType objectType,
	obj map[string]interface{},
	path []string,
	sourceKey string,
) (string, error) {
	if kubeClient == nil {
		return "", errors.New("kubeClient required for this functionality")
	}

	var resource schema.GroupVersionResource
	if objType == secretObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	} else if objType == configMapObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	}

	resourceName, ok, err := unstructured.NestedString(obj, path...)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("not found")
	}

	otherObj, err := kubeClient.Resource(resource).Namespace(ns).Get(resourceName, v1.GetOptions{})
	if err != nil {
		return "", err
	}

	val, ok, err := unstructured.NestedString(otherObj.Object, "data", sourceKey)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("not found")
	}
	if objType == secretObjectType {
		n, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return "", err
		}
		val = string(n)
	}
	return val, nil
}

func produceMapValueFromPath(
	obj map[string]interface{},
	path []string,
	sourceKey string,
	sourceValue string,
) (map[string]interface{}, error) {
	val, ok, err := unstructured.NestedSlice(obj, path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	r := make(map[string]interface{})
	for _, e := range val {
		if mm, ok := e.(map[string]interface{}); ok {
			k := mm[sourceKey]
			ks := k.(string)
			v := mm[sourceValue]
			r[ks] = v
		}
	}

	return r, nil
}

func produceSliceOfStringsValueFromPath(
	obj map[string]interface{},
	path []string,
	sourceValue string,
) ([]string, error) {
	val, ok, err := unstructured.NestedSlice(obj, path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	r := make([]string, 0, len(val))
	for _, e := range val {
		if mm, ok := e.(map[string]interface{}); ok {
			v := mm[sourceValue].(string)
			r = append(r, v)
		}
	}

	return r, nil
}

func produceMapValueFromDataField(
	kubeClient dynamic.Interface,
	ns string,
	objType objectType,
	obj map[string]interface{},
	path []string,
	sourceKey string,

) (map[string]string, error) {
	if kubeClient == nil {
		return nil, errors.New("kubeClient required for this functionality")
	}

	var resource schema.GroupVersionResource
	if objType == secretObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	} else if objType == configMapObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	}

	resourceName, ok, err := unstructured.NestedString(obj, path...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	otherObj, err := kubeClient.Resource(resource).Namespace(ns).Get(resourceName, v1.GetOptions{})
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
	if objType == secretObjectType {
		for k, v := range val {
			n, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return nil, err
			}
			val[k] = string(n)
		}
	}
	return val, nil
}

type bindingValue interface{}

func produceValue(
	m *bindingDefinition,
	obj unstructured.Unstructured,
	kubeClient dynamic.Interface,
) (bindingValue, error) {
	path := m.getPath()
	isStringElementType := m.elementType == stringElementType
	isStringObjectType := m.objectType == stringObjectType
	isMapElementType := m.elementType == mapElementType
	isSliceOfMapsElementType := m.elementType == sliceOfMapsElementType
	isSliceOfStringsElementType := m.elementType == sliceOfStringsElementType
	hasDataField := (m.objectType == secretObjectType || m.objectType == configMapObjectType)

	switch {
	case isStringElementType && isStringObjectType:
		return produceStringValue(obj.Object, path)

	case isStringElementType && hasDataField:
		return produceStringValueFromDataField(
			kubeClient, obj.GetNamespace(), m.objectType, obj.Object, path, m.sourceKey)

	case isMapElementType && hasDataField:
		return produceMapValueFromDataField(
			kubeClient, obj.GetNamespace(), m.objectType, obj.Object, path, m.sourceKey)

	case isMapElementType && isStringObjectType:
		return produceStringOfMapValue(obj.Object, path)

	case isSliceOfMapsElementType:
		return produceMapValueFromPath(obj.Object, path, m.sourceKey, m.sourceValue)

	case isSliceOfStringsElementType:
		return produceSliceOfStringsValueFromPath(obj.Object, path, m.sourceValue)
	}

	return nil, errors.New("not implemented")
}

type SpecHandler struct {
	kubeClient      dynamic.Interface
	obj             unstructured.Unstructured
	annotationKey   string
	annotationValue string
	restMapper      meta.RESTMapper
}

func (s *SpecHandler) Handle() (result, error) {
	m, err := newBindingDefinitionFromAnnotation(s.annotationValue)
	if err != nil {
		return result{}, err
	}
	val, err := produceValue(m, s.obj, s.kubeClient)
	if err != nil {
		return result{}, err
	}

	data := map[string]interface{}{}
	p := strings.SplitN(s.annotationKey, "/", 2)
	if len(p) > 1 && (len(p[1]) > 0) {
		data[p[1]] = val
	} else {
		switch val2 := val.(type) {
		case map[string]interface{}:
			for k, v := range val2 {
				data[k] = v
			}

		}
		if val2, ok := val.(map[string]interface{}); ok {
			for k, v := range val2 {
				data[k] = v
			}
		}
	}

	return result{Data: data}, nil
}

func newSpecHandler(
	kubeClient dynamic.Interface,
	annotationKey string,
	annotationValue string,
	obj unstructured.Unstructured,
	restMapper meta.RESTMapper,
) (handler, error) {
	return &SpecHandler{
		kubeClient:      kubeClient,
		obj:             obj,
		annotationKey:   annotationKey,
		annotationValue: annotationValue,
		restMapper:      restMapper,
	}, nil
}

func isSpec(annotationKey string) bool {
	return strings.HasPrefix(annotationKey, "service.binding")
}
