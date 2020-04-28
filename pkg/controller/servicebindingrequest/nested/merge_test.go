package nested

import (
	"testing"

	"github.com/imdario/mergo"
	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	dst := map[string]interface{}{
		"status": map[string]interface{}{
			"listeners": []map[string]interface{}{
				{
					"type": "plain",
					"addresses": []map[string]interface{}{
						{
							"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
							"port": "9092",
						},
					},
				},
			},
		},
	}

	src := map[string]interface{}{
		"status": map[string]interface{}{
			"listeners": []map[string]interface{}{
				{
					"type": "secure",
					"addresses": []map[string]interface{}{
						{
							"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
							"port": "9093",
						},
					},
				},
			},
		},
	}

	expected := map[string]interface{}{
		"status": map[string]interface{}{
			"listeners": []map[string]interface{}{
				{
					"type": "plain",
					"addresses": []map[string]interface{}{
						{
							"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
							"port": "9092",
						},
					},
				},
				{
					"type": "secure",
					"addresses": []map[string]interface{}{
						{
							"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
							"port": "9093",
						},
					},
				},
			},
		},
	}

	err := mergo.Merge(&dst, src, WithSmartMerge)
	require.NoError(t, err)
	require.Equal(t, expected, dst)
}
