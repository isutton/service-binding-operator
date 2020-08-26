package binding

import (
	"encoding/base64"
	"errors"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Producer interface {
	Produce(obj map[string]interface{}) (Value, error)
}

type ToProducers interface {
	Map([]Definition) ([]Producer, error)
}

type producerMapper struct {
	kubeClient dynamic.Interface
}

var _ ToProducers = (*producerMapper)(nil)

func (m *producerMapper) mapProducer(d Definition) Producer {
	isStringElementType := d.GetElementType() == stringElementType
	isStringObjectType := d.GetObjectType() == stringObjectType
	isMapElementType := d.GetElementType() == mapElementType
	isSliceOfMapsElementType := d.GetElementType() == sliceOfMapsElementType
	isSliceOfStringsElementType := d.GetElementType() == sliceOfStringsElementType
	hasDataField := (d.GetObjectType() == secretObjectType || d.GetObjectType() == configMapObjectType)

	switch {
	case isStringElementType && isStringObjectType:
		return &stringProducer{definition: d}

	case isStringElementType && hasDataField:
		return &stringFromDataFieldProducer{
			definition: d,
			kubeClient: m.kubeClient,
		}

	case isMapElementType && hasDataField:
		return &mapFromDataFieldProducer{
			definition: d,
			kubeClient: m.kubeClient,
		}

	case isMapElementType && isStringObjectType:
		return &stringOfMapProducer{definition: d}

	case isSliceOfMapsElementType:
		return &sliceOfMapsFromPathProducer{definition: d}

	case isSliceOfStringsElementType:
		return &sliceOfStringsFromPathProducer{definition: d}
	}

	return nil
}

func (m *producerMapper) Map(definitions []Definition) ([]Producer, error) {
	producers := make([]Producer, 0)
	for _, d := range definitions {
		p := m.mapProducer(d)
		if p != nil {
			producers = append(producers, m.mapProducer(d))
		}
	}

	return producers, nil
}

type stringProducer struct {
	definition Definition
}

var _ Producer = (*stringProducer)(nil)

func (s *stringProducer) Produce(obj map[string]interface{}) (Value, error) {
	val, ok, err := unstructured.NestedFieldCopy(obj, s.definition.GetPath()...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	m := map[string]interface{}{
		"": fmt.Sprintf("%v", val),
	}
	return &value{v: m}, nil
}

type stringOfMapProducer struct {
	definition Definition
}

var _ Producer = (*stringOfMapProducer)(nil)

func (p *stringOfMapProducer) Produce(obj map[string]interface{}) (Value, error) {
	val, ok, err := unstructured.NestedStringMap(obj, p.definition.GetPath()...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}
	v := map[string]interface{}{
		"": val,
	}
	return &value{v: v}, nil
}

type stringFromDataFieldProducer struct {
	definition Definition
	kubeClient dynamic.Interface
	ns         string
}

var _ Producer = (*stringFromDataFieldProducer)(nil)

func (s *stringFromDataFieldProducer) Produce(obj map[string]interface{}) (Value, error) {
	if s.kubeClient == nil {
		return nil, errors.New("kubeClient required for this functionality")
	}

	var resource schema.GroupVersionResource
	if s.definition.GetObjectType() == secretObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	} else if s.definition.GetObjectType() == configMapObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	}

	resourceName, ok, err := unstructured.NestedString(obj, s.definition.GetPath()...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	otherObj, err := s.kubeClient.Resource(resource).Namespace(s.ns).Get(resourceName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	val, ok, err := unstructured.NestedString(otherObj.Object, "data", s.definition.GetSourceKey())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}
	if s.definition.GetObjectType() == secretObjectType {
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

type sliceOfMapsFromPathProducer struct {
	definition Definition
}

var _ Producer = (*sliceOfMapsFromPathProducer)(nil)

func (s *sliceOfMapsFromPathProducer) Produce(obj map[string]interface{}) (Value, error) {
	val, ok, err := unstructured.NestedSlice(obj, s.definition.GetPath()...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	v := make(map[string]interface{})
	for _, e := range val {
		if mm, ok := e.(map[string]interface{}); ok {
			key := mm[s.definition.GetSourceKey()]
			ks := key.(string)
			value := mm[s.definition.GetSourceValue()]
			v[ks] = value
		}
	}

	return &value{v: v}, nil
}

type sliceOfStringsFromPathProducer struct {
	definition Definition
}

var _ Producer = (*sliceOfStringsFromPathProducer)(nil)

func (s *sliceOfStringsFromPathProducer) Produce(obj map[string]interface{}) (Value, error) {
	val, ok, err := unstructured.NestedSlice(obj, s.definition.GetPath()...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	v := make([]string, 0, len(val))
	for _, e := range val {
		if mm, ok := e.(map[string]interface{}); ok {
			sourceValue := mm[s.definition.GetSourceValue()].(string)
			v = append(v, sourceValue)
		}
	}

	return &value{v: v}, nil
}

type mapFromDataFieldProducer struct {
	definition Definition
	kubeClient dynamic.Interface
	ns         string
}

var _ Producer = (*mapFromDataFieldProducer)(nil)

func (p *mapFromDataFieldProducer) Produce(obj map[string]interface{}) (Value, error) {
	if p.kubeClient == nil {
		return nil, errors.New("kubeClient required for this functionality")
	}

	var resource schema.GroupVersionResource
	if p.definition.GetObjectType() == secretObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	} else if p.definition.GetObjectType() == configMapObjectType {
		resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	}

	resourceName, ok, err := unstructured.NestedString(obj, p.definition.GetPath()...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	otherObj, err := p.kubeClient.Resource(resource).Namespace(p.ns).Get(resourceName, v1.GetOptions{})
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
	if p.definition.GetObjectType() == secretObjectType {
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
