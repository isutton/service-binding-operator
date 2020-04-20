package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestFindService(t *testing.T) {
	ns := "find-cr-tests"
	resourceRef := "db-testing"

	f := mocks.NewFake(t, ns)

	f.AddMockedUnstructuredCSV("cluster-service-version")
	db := f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()

	// NOTE(isuttonl): is there any utility to convert between schema and meta GroupVersionKind?
	gvk := metav1.GroupVersionKind{
		Group:   db.GetObjectKind().GroupVersionKind().Group,
		Version: db.GetObjectKind().GroupVersionKind().Version,
		Kind:    db.GetObjectKind().GroupVersionKind().Kind,
	}

	t.Run("missing service namespace", func(t *testing.T) {
		s := v1alpha1.BackingServiceSelector{
			GroupVersionKind: gvk,
			Namespace:        nil,
			ResourceRef:      resourceRef,
		}
		cr, err := findService(f.FakeDynClient(), s)
		require.Error(t, err)
		require.Equal(t, err, errBackingServiceNamespace)
		require.Nil(t, cr)
	})

	t.Run("golden path", func(t *testing.T) {
		s := v1alpha1.BackingServiceSelector{
			GroupVersionKind: gvk,
			Namespace:        &ns,
			ResourceRef:      resourceRef,
		}
		cr, err := findService(f.FakeDynClient(), s)
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
		cr, err := findService(f.FakeDynClient(), *sbr.Spec.BackingServiceSelector)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})
}

func TestFindServiceCRD(t *testing.T) {
	ns := "planner"
	f := mocks.NewFake(t, ns)
	expected := f.AddMockedUnstructuredDatabaseCRD()
	cr := f.AddMockedDatabaseCR("database", ns)

	t.Run("golden path", func(t *testing.T) {
		crd, err := findServiceCRD(f.FakeDynClient(), cr.GetObjectKind().GroupVersionKind())
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
