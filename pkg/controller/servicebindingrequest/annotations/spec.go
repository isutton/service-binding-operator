package annotations

import (
	"strings"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/binding"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
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
	opts := binding.NewAnnotationMapperOptions(s.annotationKey, s.annotationValue)
	d, err := mapper.Map(opts)
	if err != nil {
		return result{}, err
	}

	val, err := d.Apply(&s.obj)
	if err != nil {
		return result{}, err
	}

	v := val.GetValue()

	path := strings.Join(d.GetPath(), ".")

	out := make(map[string]interface{})

	switch t := v.(type) {
	case map[string]string:
		for k, v := range t {
			out[k] = v
		}
	case map[string]interface{}:
		for k, v := range t {
			out[k] = v
		}
	}

	return result{
		Data:    out,
		RawData: nested.ComposeValue(out, nested.NewPath(path)),
	}, nil
}

func NewSpecHandler(
	kubeClient dynamic.Interface,
	annotationKey string,
	annotationValue string,
	obj unstructured.Unstructured,
	restMapper meta.RESTMapper,
) (*SpecHandler, error) {
	return &SpecHandler{
		kubeClient:      kubeClient,
		obj:             obj,
		annotationKey:   annotationKey,
		annotationValue: annotationValue,
		restMapper:      restMapper,
	}, nil
}

func IsSpec(annotationKey string) bool {
	return strings.HasPrefix(annotationKey, "service.binding")
}
