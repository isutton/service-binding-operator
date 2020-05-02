package annotations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestAttributeHandler exercises the AttributeHandler's ability to extract values according to the
// given annotation name and value.
func TestAttributeHandler(t *testing.T) {
	type args struct {
		HandlerArgs
		expected map[string]interface{}
	}

	assertHandler := func(args args) func(t *testing.T) {
		return func(t *testing.T) {
			bindingInfo, err := NewBindingInfo(args.Name, args.Value)
			require.NoError(t, err)
			handler := NewAttributeHandler(bindingInfo, *args.Object)
			got, err := handler.Handle()
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, args.expected, got.Object)
		}
	}

	// "scalar" tests whether a single deep scalar value can be extracted from the given object.
	t.Run("should extract a single value from .status.dbConnectionIP into .status.dbConnectionIP",
		assertHandler(args{
			expected: map[string]interface{}{
				"status": map[string]interface{}{
					"dbConnectionIP": "127.0.0.1",
				},
			},
			HandlerArgs: HandlerArgs{
				Name:  "servicebindingoperator.redhat.io/status.dbConnectionIP",
				Value: "binding:env:attribute",
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"dbConnectionIP": "127.0.0.1",
						},
					},
				},
			},
		}),
	)

	// "scalar#alias" tests whether a single deep scalar value can be extracted from the given object
	// returning a different name than the original given path.
	t.Run("should extract a single value from .status.dbConnectionIP into .alias",
		assertHandler(args{
			expected: map[string]interface{}{
				"alias": "127.0.0.1",
			},
			HandlerArgs: HandlerArgs{
				Name:  "servicebindingoperator.redhat.io/alias-status.dbConnectionIP",
				Value: "binding:env:attribute",
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"dbConnectionIP": "127.0.0.1",
						},
					},
				},
			},
		}),
	)

	// tests whether a deep slice value can be extracted from the given object.
	t.Run("should extract a slice from .status.dbConnectionIPs into .status.dbConnectionIPs",
		assertHandler(args{
			expected: map[string]interface{}{
				"status": map[string]interface{}{
					"dbConnectionIPs": []string{"127.0.0.1", "1.1.1.1"},
				},
			},
			HandlerArgs: HandlerArgs{
				Name:  "servicebindingoperator.redhat.io/status.dbConnectionIPs",
				Value: "binding:env:attribute",
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"dbConnectionIPs": []string{"127.0.0.1", "1.1.1.1"},
						},
					},
				},
			},
		}),
	)

	// tests whether a deep map value can be extracted from the given object.
	t.Run("should extract a map from .status.connection into .status.connection", assertHandler(args{
		expected: map[string]interface{}{
			"status": map[string]interface{}{
				"connection": map[string]interface{}{
					"host": "127.0.0.1",
					"port": "1234",
				},
			},
		},
		HandlerArgs: HandlerArgs{
			Name:  "servicebindingoperator.redhat.io/status.connection",
			Value: "binding:env:attribute",
			Object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"connection": map[string]interface{}{
							"host": "127.0.0.1",
							"port": "1234",
						},
					},
				},
			},
		},
	}))

	// "map.key" tests whether a deep map key can be extracted from the given object.
	t.Run("should extract a single map key from .status.connection into .status.connection",
		assertHandler(args{
			expected: map[string]interface{}{
				"status": map[string]interface{}{
					"connection": map[string]interface{}{
						"host": "127.0.0.1",
					},
				},
			},
			HandlerArgs: HandlerArgs{
				Name:  "servicebindingoperator.redhat.io/status.connection.host",
				Value: "binding:env:attribute",
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"connection": map[string]interface{}{
								"host": "127.0.0.1",
								"port": "1234",
							},
						},
					},
				},
			},
		}),
	)
}
