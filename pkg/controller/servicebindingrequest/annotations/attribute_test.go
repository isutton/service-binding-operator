package annotations

import (
	"testing"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
	"github.com/stretchr/testify/require"
)

func TestHandle(t *testing.T) {
	type args struct {
		HandlerArgs
		expected map[string]interface{}
	}

	assertDo := func(args args) func(t *testing.T) {
		return func(t *testing.T) {
			bindingInfo, err := bindinginfo.NewBindingInfo(args.Name, args.Value)
			require.NoError(t, err)
			cmd := NewAttributeHandler(bindingInfo, args.Resource)
			got, err := cmd.Handle()
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, args.expected, got)
		}
	}

	t.Run("attribute/scalar", assertDo(args{
		expected: map[string]interface{}{
			"status": map[string]interface{}{
				"dbConnectionIP": []string{"127.0.0.1"},
			},
		},
		HandlerArgs: HandlerArgs{
			Name:  "servicebindingoperator.redhat.io/status.dbConnectionIP",
			Value: "binding:env:attribute",
			Resource: map[string]interface{}{
				"status": map[string]interface{}{
					"dbConnectionIP": "127.0.0.1",
				},
			},
		},
	}))

	t.Run("attribute/scalar#alias", assertDo(args{
		expected: map[string]interface{}{
			"alias": []string{"127.0.0.1"},
		},
		HandlerArgs: HandlerArgs{
			Name:  "servicebindingoperator.redhat.io/alias-status.dbConnectionIP",
			Value: "binding:env:attribute",
			Resource: map[string]interface{}{
				"status": map[string]interface{}{
					"dbConnectionIP": "127.0.0.1",
				},
			},
		},
	}))

	t.Run("attribute/slice", assertDo(args{
		expected: map[string]interface{}{
			"status": map[string]interface{}{
				"dbConnectionIPs": []string{"127.0.0.1", "1.1.1.1"},
			},
		},
		HandlerArgs: HandlerArgs{
			Name:  "servicebindingoperator.redhat.io/status.dbConnectionIPs",
			Value: "binding:env:attribute",
			Resource: map[string]interface{}{
				"status": map[string]interface{}{
					"dbConnectionIPs": []string{"127.0.0.1", "1.1.1.1"},
				},
			},
		},
	}))

	t.Run("attribute/map", assertDo(args{
		expected: map[string]interface{}{
			"status": map[string]interface{}{
				"connection": []map[string]interface{}{
					{
						"host": "127.0.0.1",
						"port": "1234",
					},
				},
			},
		},
		HandlerArgs: HandlerArgs{
			Name:  "servicebindingoperator.redhat.io/status.connection",
			Value: "binding:env:attribute",
			Resource: map[string]interface{}{
				"status": map[string]interface{}{
					"connection": map[string]interface{}{
						"host": "127.0.0.1",
						"port": "1234",
					},
				},
			},
		},
	}))

	t.Run("attribute/map.key", assertDo(args{
		expected: map[string]interface{}{
			"status": map[string]interface{}{
				"connection": map[string]interface{}{
					"host": []string{"127.0.0.1"},
				},
			},
		},
		HandlerArgs: HandlerArgs{
			Name:  "servicebindingoperator.redhat.io/status.connection.host",
			Value: "binding:env:attribute",
			Resource: map[string]interface{}{
				"status": map[string]interface{}{
					"connection": map[string]interface{}{
						"host": "127.0.0.1",
						"port": "1234",
					},
				},
			},
		},
	}))

}
