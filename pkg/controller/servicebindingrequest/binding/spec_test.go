package binding

import (
	"testing"

	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSpecHandler(t *testing.T) {
	type args struct {
		name            string
		value           string
		service         map[string]interface{}
		resources       []runtime.Object
		expectedData    map[string]interface{}
		expectedRawData map[string]interface{}
	}

	assertHandler := func(args args) func(*testing.T) {
		return func(t *testing.T) {
			f := mocks.NewFake(t, "test")

			for _, r := range args.resources {
				f.AddMockResource(r)
			}

			restMapper := testutils.BuildTestRESTMapper()

			handler, err := NewSpecHandler(
				f.FakeDynClient(),
				args.name,
				args.value,
				unstructured.Unstructured{Object: args.service},
				restMapper,
			)
			require.NoError(t, err)
			got, err := handler.Handle()
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, args.expectedData, got.Data, "Data does not match expected")
			require.Equal(t, args.expectedRawData, got.RawData, "RawData does not match expected")
		}
	}

	t.Run("should return the from the related config map", assertHandler(args{
		name:  "service.binding/password",
		value: "path={.status.dbCredentials.password}",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"password": "hunter2",
				},
			},
		},
		resources: []runtime.Object{
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind: "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},

				Data: map[string]string{
					"password": "hunter2",
				},
			},
		},
		expectedData: map[string]interface{}{
			"password": "hunter2",
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"password": "hunter2",
				},
			},
		},
	}))
}
