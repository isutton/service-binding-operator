package servicebindingrequest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stesting "k8s.io/client-go/testing"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// assertDeploymentVolumesUpdate asserts whether containers in a deployment have their volume mounts
// and volumes properly updated.
func assertDeploymentVolumesUpdate(t *testing.T, actions []k8stesting.Action) {
	deploymentsResource := v1.SchemeGroupVersion.WithResource("deployments")
	options := FilterOptions{
		Verb:     Update,
		Resource: &deploymentsResource,
	}
	updates := filterActions(actions, options)
	require.Len(t, updates, 1)

	d := &v1.Deployment{}
	err := objectForAction(updates[0], d)
	require.NoError(t, err)

	containers := d.Spec.Template.Spec.Containers
	require.Equal(t, 1, len(containers))
	require.Equal(t, 1, len(containers[0].VolumeMounts))
	require.Equal(t, "/var/redhat", containers[0].VolumeMounts[0].MountPath)
	require.Equal(t, reconcilerName, containers[0].VolumeMounts[0].Name)

	volumes := d.Spec.Template.Spec.Volumes
	require.Equal(t, 1, len(volumes))
	require.Equal(t, reconcilerName, volumes[0].Name)
	require.Equal(t, reconcilerName, volumes[0].VolumeSource.Secret.SecretName)
}

// assertDeploymentContainersUpdate asserts whether containers in a deployment have their EnvFrom
// fields properly updated.
func assertDeploymentContainersUpdate(t *testing.T, actions []k8stesting.Action) {
	deploymentsResource := v1.SchemeGroupVersion.WithResource("deployments")
	options := FilterOptions{
		Verb:     Update,
		Resource: &deploymentsResource,
	}
	updates := filterActions(actions, options)
	require.Len(t, updates, 1)

	d := &v1.Deployment{}
	err := objectForAction(updates[0], d)
	require.NoError(t, err)

	containers := d.Spec.Template.Spec.Containers
	require.Equal(t, 1, len(containers))
	require.Equal(t, 1, len(containers[0].EnvFrom))
	require.NotNil(t, containers[0].EnvFrom[0].SecretRef)
	require.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)
}

// assertServiceBindingRequestUpdate asserts whether a ServiceBindingRequest have been properly
// updated.
func assertServiceBindingRequestUpdate(t *testing.T, actions []k8stesting.Action) {
	resource := v1alpha1.SchemeGroupVersion.WithResource("servicebindingrequests")
	subresource := ""
	options := FilterOptions{
		Verb:        Update,
		Resource:    &resource,
		Subresource: &subresource,
	}
	updates := filterActions(actions, options)
	require.Len(t, updates, 1)

	sbrOutput := &v1alpha1.ServiceBindingRequest{}
	err := objectForAction(updates[0], sbrOutput)
	require.NoError(t, err)

	require.Equal(t, "Success", sbrOutput.Status.BindingStatus)
	require.Equal(t, reconcilerName, sbrOutput.Status.Secret)
	require.Equal(t, 1, len(sbrOutput.Status.ApplicationObjects))
	require.Equal(t, fmt.Sprintf("%s/%s", reconcilerNs, reconcilerName), sbrOutput.Status.ApplicationObjects[0])
	require.Equal(t, reconcilerName, sbrOutput.Status.Secret)
}

// objectForAction attempts to return the runtime.Object associated with the given action, if any.
func objectForAction(action k8stesting.Action, out interface{}) error {
	var candidate runtime.Object

	switch a := action.(type) {
	case k8stesting.UpdateAction:
		candidate = a.GetObject()
	}

	if candidate != nil {
		u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(candidate)
		if err != nil {
			return err
		}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, out)
		if err != nil {
			return err
		}

		return nil
	}

	panic("ugh")
}

// FilterOptionsVerb is the enum of all action verbs.
type FilterOptionsVerb string

const (
	// Update includes all updates issued to an object.
	Update FilterOptionsVerb = "update"
)

// FilterOptions holds options to filter actions during unit tests assertion phase.
type FilterOptions struct {
	// Verb is one of the verbs encoded in actions, for example "create", "update" or "delete" among
	// others.
	Verb FilterOptionsVerb
	// Resource is used to filter events to a specific GroupVersionResource.
	Resource *schema.GroupVersionResource
	// Subresource is used to filter events to a specific sub-resource.
	Subresource *string
}

// filterActions attempts to filter the given slice of actions according to the given filter options.
func filterActions(actions []k8stesting.Action, options FilterOptions) []k8stesting.Action {
	var filteredActions []k8stesting.Action

	for _, a := range actions {
		if len(options.Verb) > 0 && FilterOptionsVerb(a.GetVerb()) != options.Verb {
			continue
		}
		if options.Resource != nil && !matchResources(a.GetResource(), *options.Resource) {
			continue
		}
		if options.Subresource != nil && a.GetSubresource() != *options.Subresource {
			continue
		}
		filteredActions = append(filteredActions, a)
	}

	return filteredActions
}

// matchResources returns the result of the comparison of compares all given resources' fields.
func matchResources(a, b schema.GroupVersionResource) bool {
	return strings.ToLower(a.Resource) == strings.ToLower(b.Resource) &&
		strings.ToLower(a.Group) == strings.ToLower(b.Group) &&
		strings.ToLower(a.Version) == strings.ToLower(b.Version)
}
