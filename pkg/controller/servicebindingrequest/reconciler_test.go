package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

const (
	reconcilerNs   = "testing"
	reconcilerName = "binding-request"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

// reconcileRequest creates a reconcile.Request object using global variables.
func reconcileRequest() reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: reconcilerNs,
			Name:      reconcilerName,
		},
	}
}

func TestReconcilerReconcileError(t *testing.T) {
	backingServiceResourceRef := "test-using-secret"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}

	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, "", matchLabels)

	fakeClient := f.FakeClient()
	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{client: fakeClient, dynClient: fakeDynClient, scheme: f.S}

	res, err := reconciler.Reconcile(reconcileRequest())

	// FIXME: decide this test's fate
	// I'm not very sure what this test was about, but in the case the SBR definition contains
	// references to objects that do not exist, the reconciliation process is supposed to be
	// successful. Commented below was the original test.
	//
	// require.Error(t, err)
	// require.True(t, res.Requeue)

	require.NoError(t, err)
	require.True(t, res.Requeue)
}

// TestReconcilerReconcile test the reconciliation process using a secret, expected to be
// the regular approach.
func TestReconcilerReconcile(t *testing.T) {
	// Arrange
	backingServiceResourceRef := "backingServiceRef"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}

	t.Run("reconcile-using-applicationResourceRef", func(t *testing.T) {
		// Arrange
		f := mocks.NewFake(t, reconcilerNs)
		addReconcilerCommonMocks(f, matchLabels, backingServiceResourceRef, "applicationRef", reconcilerName)
		f.AddMockedUnstructuredCSV("cluster-service-version-list")
		fakeClient := f.FakeClient()
		fakeDynClient := f.FakeDynClient()
		reconciler := &Reconciler{client: fakeClient, dynClient: fakeDynClient, scheme: f.S}

		// Act
		res, err := reconciler.Reconcile(reconcileRequest())

		// Assert
		updateActions := filterActions(fakeDynClient.Fake.Actions(), FilterOptions{Verb: Update})
		require.NoError(t, err)
		require.False(t, res.Requeue)
		assertServiceBindingRequestUpdate(t, updateActions)
	})

	t.Run("reconcile-using-volume", func(t *testing.T) {
		// Arrange
		f := mocks.NewFake(t, reconcilerNs)
		addReconcilerCommonMocks(f, matchLabels, backingServiceResourceRef, "", reconcilerName)
		f.AddMockedUnstructuredCSVWithVolumeMount("csv-with-volume-mount")
		fakeClient := f.FakeClient()
		fakeDynClient := f.FakeDynClient()
		reconciler := &Reconciler{client: fakeClient, dynClient: fakeDynClient, scheme: f.S}

		// Act
		res, err := reconciler.Reconcile(reconcileRequest())

		// Assert
		updateActions := filterActions(fakeDynClient.Fake.Actions(), FilterOptions{Verb: Update})
		require.NoError(t, err)
		require.False(t, res.Requeue)
		require.Len(t, updateActions, 7)
		assertDeploymentVolumesUpdate(t, updateActions)
		assertServiceBindingRequestUpdate(t, updateActions)
	})

	t.Run("reconcile-using-secret", func(t *testing.T) {
		// Arrange
		f := mocks.NewFake(t, reconcilerNs)
		addReconcilerCommonMocks(f, matchLabels, "reconcile-using-secret", "", reconcilerName)
		f.AddMockedUnstructuredCSV("csv-with-secret-mount")
		fakeClient := f.FakeClient()
		fakeDynClient := f.FakeDynClient()
		reconciler := &Reconciler{client: fakeClient, dynClient: fakeDynClient, scheme: f.S}

		// Act
		res, err := reconciler.Reconcile(reconcileRequest())

		// Assert
		updateActions := filterActions(fakeDynClient.Fake.Actions(), FilterOptions{Verb: Update})
		require.NoError(t, err)
		require.False(t, res.Requeue)
		require.Len(t, updateActions, 7)
		assertDeploymentContainersUpdate(t, updateActions)
		assertServiceBindingRequestUpdate(t, updateActions)
	})
}

// addReconcilerCommonMocks installs common mocks that should be present for the reconcile to work.
func addReconcilerCommonMocks(
	f *mocks.Fake,
	matchLabels map[string]string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	sbrName string,
) {
	f.AddMockedUnstructuredServiceBindingRequest(
		sbrName, backingServiceResourceRef, applicationResourceRef, matchLabels)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(sbrName, matchLabels)
	f.AddMockedSecret("db-credentials")
}
