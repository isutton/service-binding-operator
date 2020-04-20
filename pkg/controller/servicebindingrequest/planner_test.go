package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestPlanner(t *testing.T) {
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

	// Out of the box, our mocks don't set the namespace
	// ensure SearchCR fails.
	t.Run("findCR missing service namespace", func(t *testing.T) {
		cr, err := findCR(f.FakeDynClient(), *sbr.Spec.BackingServiceSelector)
		require.Error(t, err)
		require.Equal(t, err, errBackingServiceNamespace)
		require.Nil(t, cr)
	})

	// FIXME(isuttonl): move this test to servicecontext
	t.Run("extract existing selectors", func(t *testing.T) {
		serviceCtxs, err := buildServiceContexts(
			f.FakeDynClient(), ns, extractServiceSelectors(sbr))
		require.NoError(t, err)
		require.NotEmpty(t, serviceCtxs)
	})

	// The searchCR contract only cares about the backingServiceNamespace
	sbr.Spec.BackingServiceSelector.Namespace = &ns
	t.Run("findCR", func(t *testing.T) {
		cr, err := findCR(f.FakeDynClient(), *sbr.Spec.BackingServiceSelector)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})
}

func TestPlannerWithExplicitBackingServiceNamespace(t *testing.T) {
	ns := "planner"
	backingServiceNamespace := "backing-service-namespace"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, &backingServiceNamespace, resourceRef, "", deploymentsGVR, matchLabels)
	require.NotNil(t, sbr.Spec.BackingServiceSelector.Namespace)

	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef, backingServiceNamespace)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddNamespacedMockedSecret("db-credentials", backingServiceNamespace)

	t.Run("findCR", func(t *testing.T) {
		cr, err := findCR(f.FakeDynClient(), *sbr.Spec.BackingServiceSelector)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})
}

func TestPlannerAnnotation(t *testing.T) {
	ns := "planner"
	f := mocks.NewFake(t, ns)
	expected := f.AddMockedUnstructuredDatabaseCRD()
	cr := f.AddMockedDatabaseCR("database", ns)

	t.Run("findCRD", func(t *testing.T) {
		crd, err := findCRD(f.FakeDynClient(), cr.GetObjectKind().GroupVersionKind())
		require.NoError(t, err)
		require.NotNil(t, crd)
		require.Equal(t, expected, crd)
	})
}

func TestPlannerLoadDescriptor(t *testing.T) {
	type args struct {
		path       string
		descriptor string
		root       string
		expected   map[string]string
	}

	assertLoadDescriptor := func(args args) func(t *testing.T) {
		return func(t *testing.T) {
			anns := map[string]string{}
			loadDescriptor(anns, args.path, args.descriptor, args.root)
			require.Equal(t, args.expected, anns)
		}
	}

	t.Run("", assertLoadDescriptor(args{
		descriptor: "binding:volumemount:secret:user",
		root:       "status",
		path:       "user",
		expected: map[string]string{
			"servicebindingoperator.redhat.io/status.user": "binding:volumemount:secret",
		},
	}))

}
