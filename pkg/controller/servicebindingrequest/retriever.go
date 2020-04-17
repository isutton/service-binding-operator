package servicebindingrequest

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/imdario/mergo"
	"github.com/redhat-developer/service-binding-operator/pkg/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Retriever reads all data referred in plan instance, and store in a secret.
type Retriever struct {
	logger        *log.Log                     // logger instance
	data          map[string][]byte            // data retrieved
	Objects       []*unstructured.Unstructured // list of objects employed
	client        dynamic.Interface            // Kubernetes API client
	plan          *Plan                        // plan instance
	VolumeKeys    []string                     // list of keys found
	bindingPrefix string                       // prefix for variable names
	cache         map[string]interface{}       // store visited paths
}

// GetEnvVars returns the data read from related resources (see ReadBindableResourcesData and
// ReadCRDDescriptionData).
func (r *Retriever) GetEnvVars() (map[string][]byte, error) {
	envVarCtx := make(map[string]interface{})

	for _, resource := range r.plan.GetServiceContexts() {
		// contribute values extracted from the service related resources
		err := mergo.Merge(&envVarCtx, resource.EnvVars, mergo.WithAppendSlice, mergo.WithOverride)
		if err != nil {
			return nil, err
		}

		// contribute the entire CR to the context
		gvk := resource.CR.GetObjectKind().GroupVersionKind()
		err = unstructured.SetNestedField(
			envVarCtx, resource.CR.Object, gvk.Version, gvk.Group, gvk.Kind, resource.CR.GetName())
		if err != nil {
			return nil, err
		}

		// FIXME(isuttonl): make volume keys a return value
		r.VolumeKeys = append(r.VolumeKeys, resource.VolumeKeys...)
	}

	envVarTemplates := r.plan.SBR.Spec.CustomEnvVar
	envParser := NewCustomEnvParser(envVarTemplates, envVarCtx)
	envVars, err := envParser.Parse()
	if err != nil {
		return nil, err
	}

	// convert values to a map[string][]byte
	result := make(map[string][]byte)
	for k, v := range envVars {
		result[k] = []byte(v.(string))
	}

	return result, nil
}

// ReadBindableResourcesData reads all related resources of a given sbr
func (r *Retriever) ReadBindableResourcesData(
	sbr *v1alpha1.ServiceBindingRequest,
	crs []*unstructured.Unstructured,
) error {
	r.logger.Info("Detecting extra resources for binding...")
	for _, cr := range crs {
		b := NewDetectBindableResources(sbr, cr, []schema.GroupVersionResource{
			{Group: "", Version: "v1", Resource: "configmaps"},
			{Group: "", Version: "v1", Resource: "services"},
			{Group: "route.openshift.io", Version: "v1", Resource: "routes"},
		}, r.client)

		vals, err := b.GetBindableVariables()
		if err != nil {
			return err
		}
		for k, v := range vals {
			r.store(cr, k, []byte(fmt.Sprintf("%v", v)))
		}
	}

	return nil
}

// getNestedValue retrieve value from dotted key path
func (r *Retriever) getNestedValue(key string, sectionMap interface{}) (string, interface{}, error) {
	if !strings.Contains(key, ".") {
		value, exists := sectionMap.(map[string]interface{})[key]
		if !exists {
			return "", sectionMap, nil
		}
		return fmt.Sprintf("%v", value), sectionMap, nil
	}
	attrs := strings.SplitN(key, ".", 2)
	newSectionMap, exists := sectionMap.(map[string]interface{})[attrs[0]]
	if !exists {
		return "", newSectionMap, nil
	}
	return r.getNestedValue(attrs[1], newSectionMap.(map[string]interface{}))
}

// getCRKey retrieve key in section from CR object, part of the "plan" instance.
func (r *Retriever) getCRKey(u *unstructured.Unstructured, section string, key string) (string, interface{}, error) {
	obj := u.Object
	objName := u.GetName()
	log := r.logger.WithValues("CR.Name", objName, "CR.section", section, "CR.key", key)
	log.Debug("Reading CR attributes...")

	sectionMap, exists := obj[section]
	if !exists {
		return "", sectionMap, fmt.Errorf("Can't find '%s' section in CR named '%s'", section, objName)
	}

	log.WithValues("SectionMap", sectionMap).Debug("Getting values from sectionmap")
	v, _, err := r.getNestedValue(key, sectionMap)

	return v, sectionMap, err
}

// store key and value, formatting key to look like an environment variable.
func (r *Retriever) store(envVarPrefix *string, u *unstructured.Unstructured, key string, value []byte) {
	key = strings.ReplaceAll(key, ":", "_")
	key = strings.ReplaceAll(key, ".", "_")
	if envVarPrefix == nil {
		if r.bindingPrefix == "" {
			key = fmt.Sprintf("%s_%s", u.GetKind(), key)
		} else {
			key = fmt.Sprintf("%s_%s_%s", r.bindingPrefix, u.GetKind(), key)
		}
	} else if *envVarPrefix == "" {
		if r.bindingPrefix != "" {
			key = fmt.Sprintf("%s_%s", r.bindingPrefix, key)
		}
	} else {
		if r.bindingPrefix != "" {
			key = fmt.Sprintf("%s_%s_%s", r.bindingPrefix, *envVarPrefix, key)
		} else {
			key = fmt.Sprintf("%s_%s", *envVarPrefix, key)
		}
	}
	key = strings.ToUpper(key)
	r.data[key] = value
}

// NewRetriever instantiate a new retriever instance.
func NewRetriever(client dynamic.Interface, plan *Plan, bindingPrefix string) *Retriever {
	return &Retriever{
		logger:        log.NewLog("retriever"),
		data:          make(map[string][]byte),
		Objects:       []*unstructured.Unstructured{},
		client:        client,
		plan:          plan,
		VolumeKeys:    []string{},
		bindingPrefix: bindingPrefix,
		cache:         make(map[string]interface{}),
	}
}
