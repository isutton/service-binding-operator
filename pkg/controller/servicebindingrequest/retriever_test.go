package servicebindingrequest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func TestRetriever(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	backingServiceNs := "backing-servicec-ns"
	crName := "db-testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddNamespacedMockedSecret("db-credentials", backingServiceNs)

	cr, err := mocks.UnstructuredDatabaseCRMock(backingServiceNs, crName)
	require.NoError(t, err)

	crInSameNamespace, err := mocks.UnstructuredDatabaseCRMock(ns, crName)
	require.NoError(t, err)

	serviceCtxs := ServiceContextList{
		{
			Object: cr,
		},
		{
			Object: crInSameNamespace,
		},
	}

	fakeDynClient := f.FakeDynClient()

	toTmpl := func(obj *unstructured.Unstructured) string {
		gvk := obj.GetObjectKind().GroupVersionKind()
		name := obj.GetName()
		return fmt.Sprintf(`{{ index . %q %q %q %q "metadata" "name" }}`, gvk.Version, gvk.Group, gvk.Kind, name)
	}

	retriever = NewRetriever(
		fakeDynClient,
		[]v1.EnvVar{
			{Name: "SAME_NAMESPACE", Value: toTmpl(crInSameNamespace)},
			{Name: "OTHER_NAMESPACE", Value: toTmpl(cr)},
			{Name: "DIRECT_ACCESS", Value: `{{ .v1alpha1.postgresql_baiju_dev.Database.db_testing.metadata.name }}`},
		},
		serviceCtxs,
		"SERVICE_BINDING",
	)
	require.NotNil(t, retriever)

	actual, err := retriever.GetEnvVars()
	require.NoError(t, err)
	require.Equal(t, map[string][]byte{
		"SERVICE_BINDING_SAME_NAMESPACE":  []byte(crInSameNamespace.GetName()),
		"SERVICE_BINDING_OTHER_NAMESPACE": []byte(cr.GetName()),
		"SERVICE_BINDING_DIRECT_ACCESS":   []byte(cr.GetName()),
	}, actual)
}

func TestBuildServiceEnvVars(t *testing.T) {

	type testCase struct {
		ctx                *ServiceContext
		globalEnvVarPrefix string
		expected           map[string]string
	}

	testCases := []testCase{
		{
			globalEnvVarPrefix: "GLOBAL",
			ctx: &ServiceContext{
				EnvVarPrefix: "SERVICE",
				EnvVars: map[string]interface{}{
					"status": map[string]interface{}{
						"host": "localhost",
					},
				},
			},
			expected: map[string]string{
				"GLOBAL_SERVICE_STATUS_HOST": "localhost",
			},
		},
		{
			globalEnvVarPrefix: "",
			ctx: &ServiceContext{
				EnvVarPrefix: "",
				EnvVars: map[string]interface{}{
					"status": map[string]interface{}{
						"host": "localhost",
					},
				},
			},
			expected: map[string]string{
				"STATUS_HOST": "localhost",
			},
		},
		{
			globalEnvVarPrefix: "",
			ctx: &ServiceContext{
				EnvVarPrefix: "SERVICE",
				EnvVars: map[string]interface{}{
					"status": map[string]interface{}{
						"host": "localhost",
					},
				},
			},
			expected: map[string]string{
				"SERVICE_STATUS_HOST": "localhost",
			},
		},
		{
			globalEnvVarPrefix: "GLOBAL",
			ctx: &ServiceContext{
				EnvVarPrefix: "",
				EnvVars: map[string]interface{}{
					"status": map[string]interface{}{
						"host": "localhost",
					},
				},
			},
			expected: map[string]string{
				"GLOBAL_STATUS_HOST": "localhost",
			},
		},
	}

	for _, tc := range testCases {
		actual, err := buildServiceEnvVars(tc.ctx, tc.globalEnvVarPrefix)
		require.NoError(t, err)
		require.Equal(t, tc.expected, actual)
	}
}
