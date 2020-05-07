package servicebindingrequest

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/envvars"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// Retriever reads all data referred in plan instance, and store in a secret.
type Retriever struct {
	logger             *log.Log                     // logger instance
	data               map[string][]byte            // data retrieved
	Objects            []*unstructured.Unstructured // list of objects employed
	client             dynamic.Interface            // Kubernetes API client
	VolumeKeys         []string                     // list of keys found
	globalEnvVarPrefix string                       // prefix for variable names
	envVarTemplates    []corev1.EnvVar              // list of environment variable names and templates
	serviceCtxs        ServiceContextList           // list of service contexts associated with a SBR
}

// createServiceIndexPath returns a string slice with fields representing a path to a resource in the
// environment variable context. This function cleans fields that might contain invalid characters to
// be used in Go template; for example, a Group might contain the "." character, which makes it
// harder to refer using Go template direct accessors and is substituted by an underbar "_".
func createServiceIndexPath(name string, gvk schema.GroupVersionKind) []string {
	return []string{
		gvk.Version,
		strings.ReplaceAll(gvk.Group, ".", "_"),
		gvk.Kind,
		strings.ReplaceAll(name, "-", "_"),
	}

}

func buildServiceEnvVars(svcCtx *ServiceContext, globalEnvVarPrefix string) (map[string]string, error) {
	prefixes := []string{}
	if len(globalEnvVarPrefix) > 0 {
		prefixes = append(prefixes, globalEnvVarPrefix)
	}
	if len(svcCtx.EnvVarPrefix) > 0 {
		prefixes = append(prefixes, svcCtx.EnvVarPrefix)
	}
	return envvars.Build(svcCtx.EnvVars, prefixes...)
}

// GetEnvVars returns the data read from related resources (see ReadBindableResourcesData and
// ReadCRDDescriptionData).
func (r *Retriever) GetEnvVars() (map[string][]byte, error) {
	customEnvVarCtx := make(map[string]interface{})
	envVars := make(map[string][]byte)

	for _, svcCtx := range r.serviceCtxs {
		svcEnvVars, err := buildServiceEnvVars(svcCtx, r.globalEnvVarPrefix)
		if err != nil {
			return nil, err
		}

		for k, v := range svcEnvVars {
			envVars[k] = []byte(v)
		}

		// contribute the entire resource to the context shared with the custom env parser
		gvk := svcCtx.Service.GetObjectKind().GroupVersionKind()

		// add an entry in the custom environment variable context, allowing the user to use the
		// following expression:
		//
		// `{{ index "v1alpha1" "postgresql.baiju.dev" "Database", "db-testing", "status", "connectionUrl" }}`
		err = unstructured.SetNestedField(
			customEnvVarCtx, svcCtx.Service.Object, gvk.Version, gvk.Group, gvk.Kind,
			svcCtx.Service.GetName())
		if err != nil {
			return nil, err
		}

		// add an entry in the custom environment variable context with modified key names (group
		// names have the "." separator changed to underbar and "-" in the resource name is changed
		// to underbar "_" as well).
		//
		// `{{ .v1alpha1.postgresql_baiju_dev.Database.db_testing.status.connectionUrl }}`
		err = unstructured.SetNestedField(
			customEnvVarCtx,
			svcCtx.Service.Object,
			createServiceIndexPath(svcCtx.Service.GetName(), svcCtx.Service.GroupVersionKind())...,
		)
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

	for k, v := range customEnvVars {
		if len(r.globalEnvVarPrefix) > 0 {
			k = strings.Join([]string{r.globalEnvVarPrefix, k}, "_")
		}
		envVars[k] = []byte(v.(string))
	}

	return envVars, nil
}

// NewRetriever instantiate a new retriever instance.
func NewRetriever(
	client dynamic.Interface,
	envVars []corev1.EnvVar,
	serviceContexts ServiceContextList,
	globalEnvVarPrefix string,
) *Retriever {
	return &Retriever{
		logger:             log.NewLog("retriever"),
		data:               make(map[string][]byte),
		Objects:            []*unstructured.Unstructured{},
		client:             client,
		VolumeKeys:         []string{},
		globalEnvVarPrefix: globalEnvVarPrefix,
		envVarTemplates:    envVars,
		serviceCtxs:        serviceContexts,
	}
}
