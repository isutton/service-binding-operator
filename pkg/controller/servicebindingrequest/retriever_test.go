package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func TestRetriever(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	backingServiceNs := "backing-servicec-ns"
	crName := "db-testing"
	testEnvVarPrefix := "TEST_PREFIX"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddNamespacedMockedSecret("db-credentials", backingServiceNs)

	crdDescription := mocks.CRDDescriptionMock()
	cr, err := mocks.UnstructuredDatabaseCRMock(backingServiceNs, crName)
	require.NoError(t, err)

	crInSameNamespace, err := mocks.UnstructuredDatabaseCRMock(ns, crName)
	require.NoError(t, err)

	plan := &Plan{
		Ns:   ns,
		Name: "retriever",
		ServiceContexts: []*ServiceContext{
			{
				CRDDescription: &crdDescription,
				CR:             cr,
			},
			{
				CRDDescription: &crdDescription,
				CR:             crInSameNamespace,
			},
		},
	}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("getCRKey", func(t *testing.T) {
		imageName, _, err := retriever.getCRKey(cr, "spec", "imageName")
		require.NoError(t, err)
		require.Equal(t, "postgres", imageName)
	})

	t.Run("store", func(t *testing.T) {
		retriever.store(&testEnvVarPrefix, cr, "test", []byte("test"))
		require.Contains(t, retriever.data, "SERVICE_BINDING_TEST_PREFIX_TEST")
		require.Equal(t, []byte("test"), retriever.data["SERVICE_BINDING_TEST_PREFIX_TEST"])
	})
}

func TestRetrieverWithNestedCRKey(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddMockedSecret("db-credentials")

	crdDescription := mocks.CRDDescriptionMock()
	cr, err := mocks.UnstructuredNestedDatabaseCRMock(ns, crName)
	require.NoError(t, err)

	plan := &Plan{
		Ns:   ns,
		Name: "retriever",
		ServiceContexts: []*ServiceContext{
			{
				CRDDescription: &crdDescription,
				CR:             cr,
			},
		},
	}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("Second level", func(t *testing.T) {
		imageName, _, err := retriever.getCRKey(cr, "spec", "image.name")
		require.NoError(t, err)
		require.Equal(t, "postgres", imageName)
	})

	t.Run("Second level error", func(t *testing.T) {
		// FIXME: if attribute isn't available in CR we would not throw any error.
		t.Skip()
		_, _, err := retriever.getCRKey(cr, "spec", "image..name")
		require.NotNil(t, err)
	})

	t.Run("Third level", func(t *testing.T) {
		something, _, err := retriever.getCRKey(cr, "spec", "image.third.something")
		require.NoError(t, err)
		require.Equal(t, "somevalue", something)
	})
}
