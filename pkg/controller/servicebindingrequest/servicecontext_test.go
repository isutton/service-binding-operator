package servicebindingrequest

import (
	"testing"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
)

func TestBuildServiceContexts(t *testing.T) {
	ns := "planner"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, nil, resourceRef, "", deploymentsGVR, matchLabels)
	sbr.Spec.BackingServiceSelectors = &[]v1alpha1.BackingServiceSelector{
		*sbr.Spec.BackingServiceSelector,
	}
	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedSecret("db-credentials")

	t.Run("existing selectors", func(t *testing.T) {
		serviceCtxs, err := buildServiceContexts(
			f.FakeDynClient(), ns, extractServiceSelectors(sbr))
		require.NoError(t, err)
		require.NotEmpty(t, serviceCtxs)
	})

	t.Run("empty selectors", func(t *testing.T) {
		serviceCtxs, err := buildServiceContexts(f.FakeDynClient(), ns, nil)
		require.NoError(t, err)
		require.Empty(t, serviceCtxs)
	})

	t.Run("services in different namespace", func(t *testing.T) {
		serviceCtxs, err := buildServiceContexts(f.FakeDynClient(), ns, extractServiceSelectors(sbr))
		require.NoError(t, err)
		require.NotEmpty(t, serviceCtxs)
	})
}
