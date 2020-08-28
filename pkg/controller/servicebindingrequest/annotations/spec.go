package annotations

import (
	"errors"
	"strings"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/binding"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

type SpecHandler struct {
	kubeClient      dynamic.Interface
	obj             unstructured.Unstructured
	annotationKey   string
	annotationValue string
	restMapper      meta.RESTMapper
}

func (s *SpecHandler) Handle() (result, error) {
	mapper := &binding.AnnotationToDefinitionMapper{
		KubeClient: s.kubeClient,
	}
	d, err := mapper.Map(binding.NewAnnotationMapperOptions(s.annotationKey, s.annotationValue))
	if err != nil {
		return result{}, err
	}

	val, err := d.Apply(&s.obj)
	if err != nil {
		return result{}, err
	}

	data, ok := val.GetValue().(map[string]interface{})
	if !ok {
		return result{}, errors.New("not map[string]interface{}")
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
