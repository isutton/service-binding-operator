package annotations

import (
	"testing"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestAttributeHandler(t *testing.T) {
	type args struct {
		HandlerArgs
		expected map[string]interface{}
	}

	assertHandler := func(args args) func(t *testing.T) {
		return func(t *testing.T) {
			bindingInfo, err := bindinginfo.NewBindingInfo(args.Name, args.Value)
			require.NoError(t, err)
			handler := NewAttributeHandler(bindingInfo, *args.Resource)
			got, err := handler.Handle()
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, args.expected, got.Object)
		}
	}

	t.Run("attribute/scalar", assertHandler(args{
		expected: map[string]interface{}{
			"status": map[string]interface{}{
				"dbConnectionIP": "127.0.0.1",
			},
		},
		HandlerArgs: HandlerArgs{
			Name:  "servicebindingoperator.redhat.io/status.dbConnectionIP",
			Value: "binding:env:attribute",
			Resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"dbConnectionIP": "127.0.0.1",
					},
				},
			},
		},
	}))

	t.Run("attribute/scalar#alias", assertHandler(args{
		expected: map[string]interface{}{
			"alias": "127.0.0.1",
		},
		HandlerArgs: HandlerArgs{
			Name:  "servicebindingoperator.redhat.io/alias-status.dbConnectionIP",
			Value: "binding:env:attribute",
			Resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"dbConnectionIP": "127.0.0.1",
					},
				},
			},
		},
	}))

	t.Run("attribute/slice", assertHandler(args{
		expected: map[string]interface{}{
			"status": map[string]interface{}{
				"dbConnectionIPs": []string{"127.0.0.1", "1.1.1.1"},
			},
		},
		HandlerArgs: HandlerArgs{
			Name:  "servicebindingoperator.redhat.io/status.dbConnectionIPs",
			Value: "binding:env:attribute",
			Resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"dbConnectionIPs": []string{"127.0.0.1", "1.1.1.1"},
					},
				},
			},
		},
	}))

	t.Run("attribute/map", assertHandler(args{
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
			Resource: &unstructured.Unstructured{
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

	t.Run("attribute/map.key", assertHandler(args{
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
			Resource: &unstructured.Unstructured{
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

}
