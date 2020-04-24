package servicebindingrequest

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/imdario/mergo"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/envvars"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// Retriever reads all data referred in plan instance, and store in a secret.
type Retriever struct {
	logger          *log.Log                     // logger instance
	data            map[string][]byte            // data retrieved
	Objects         []*unstructured.Unstructured // list of objects employed
	client          dynamic.Interface            // Kubernetes API client
	VolumeKeys      []string                     // list of keys found
	bindingPrefix   string                       // prefix for variable names
	envVarTemplates []corev1.EnvVar              // list of environment variable names and templates
	serviceCtxs     ServiceContextList           // list of service contexts associated with a SBR
}

// GetEnvVars returns the data read from related resources (see ReadBindableResourcesData and
// ReadCRDDescriptionData).
func (r *Retriever) GetEnvVars() (map[string][]byte, error) {
	svcCollectedKeys := make(map[string]interface{})
	customEnvVarCtx := make(map[string]interface{})

	for _, svcCtx := range r.serviceCtxs {
		// contribute service contributed env vars
		err := mergo.Merge(&svcCollectedKeys, svcCtx.EnvVars, mergo.WithAppendSlice, mergo.WithOverride)
		if err != nil {
			return nil, err
		}

		// contribute the entire resource to the context shared with the custom env parser
		gvk := svcCtx.Object.GetObjectKind().GroupVersionKind()
		err = unstructured.SetNestedField(
			customEnvVarCtx, svcCtx.Object.Object, gvk.Version, gvk.Group, gvk.Kind,
			svcCtx.Object.GetName())
		if err != nil {
			return nil, err
		}

		// FIXME(isuttonl): make volume keys a return value
		r.VolumeKeys = append(r.VolumeKeys, svcCtx.VolumeKeys...)
	}

	envVarTemplates := r.envVarTemplates
	envParser := NewCustomEnvParser(envVarTemplates, customEnvVarCtx)
	customEnvVars, err := envParser.Parse()
	if err != nil {
		r.logger.Error(
			err, "Creating envVars", "Templates", envVarTemplates, "TemplateContext", customEnvVarCtx)
		return nil, err
	}

	// convert values to a map[string][]byte
	envVars := make(map[string][]byte)
	for k, v := range customEnvVars {
		envVars[k] = []byte(v.(string))
	}

	svcEnvVars, err := envvars.Build(svcCollectedKeys, []string{})
	if err != nil {
		return nil, err
	}

	for k, v := range svcEnvVars {
		envVars[k] = []byte(v)
	}

	return envVars, nil
}

// ReadBindableResourcesData reads all related resources of a given sbr
// func (r *Retriever) ReadBindableResourcesData(
// 	sbr *v1alpha1.ServiceBindingRequest,
// 	crs []*unstructured.Unstructured,
// ) error {
// 	r.logger.Info("Detecting extra resources for binding...")
// 	for _, cr := range crs {
// 		b := NewDetectBindableResources(sbr, cr, []schema.GroupVersionResource{
// 			{Group: "", Version: "v1", Resource: "configmaps"},
// 			{Group: "", Version: "v1", Resource: "services"},
// 			{Group: "route.openshift.io", Version: "v1", Resource: "routes"},
// 		}, r.client)

// 		vals, err := b.GetBindableVariables()
// 		if err != nil {
// 			return err
// 		}
// 		for k, v := range vals {
// 			// r.store("", cr, k, []byte(fmt.Sprintf("%v", v)))
// 		}
// 	}

// 	return nil
// }

// NewRetriever instantiate a new retriever instance.
func NewRetriever(
	client dynamic.Interface,
	envVars []corev1.EnvVar,
	serviceContexts ServiceContextList,
	bindingPrefix string,
) *Retriever {
	return &Retriever{
		logger:          log.NewLog("retriever"),
		data:            make(map[string][]byte),
		Objects:         []*unstructured.Unstructured{},
		client:          client,
		VolumeKeys:      []string{},
		bindingPrefix:   bindingPrefix,
		envVarTemplates: envVars,
		serviceCtxs:     serviceContexts,
	}
}
