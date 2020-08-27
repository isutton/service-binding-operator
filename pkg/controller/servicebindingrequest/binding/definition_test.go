package binding

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestDefinitionMapper(t *testing.T) {
	type args struct {
		description   string
		name          string
		value         string
		expectedValue interface{}
	}

	testCases := []args{
		{
			description:   "string definition",
			expectedValue: &stringDefinition{},
			name:          "service.binding/username",
			value:         "path={.status.dbCredential.username}",
		},
		{
			description:   "map from data field definition#Secret",
			expectedValue: &mapFromDataFieldDefinition{},
			name:          "service.binding/username",
			value:         "path={.status.dbCredential.username},objectType=Secret",
		},
		{
			description:   "map from data field definition#ConfigMap",
			expectedValue: &mapFromDataFieldDefinition{},
			name:          "service.binding/username",
			value:         "path={.status.database.dsn},objectType=ConfigMap",
		},
		{
			description:   "string of map definition",
			expectedValue: &stringOfMapDefinition{},
			name:          "service.binding/username",
			value:         "path={.status.database},elementType=map",
		},
		{
			description:   "slice of maps from path definition",
			expectedValue: &sliceOfMapsFromPathDefinition{},
			name:          "service.binding",
			value:         "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
		},
		{
			description:   "slice of strings from path definition",
			expectedValue: &sliceOfStringsFromPathDefinition{},
			name:          "service.binding",
			value:         "path={.status.bootstrap},elementType=sliceOfStrings,sourceValue=url",
		},
	}

	mapper := &definitionMapper{}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			d, err := mapper.MapAnnotation(tc.name, tc.value)
			require.NoError(t, err)
			require.IsType(t, tc.expectedValue, d)
		})
	}
}

func TestStringDefinition(t *testing.T) {
	d := &stringDefinition{
		path: []string{"status", "dbCredential", "username"},
	}
	val, err := d.Apply(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredential": map[string]interface{}{
					"username": "AzureDiamond",
				},
			},
		},
	})
	require.NoError(t, err)
	v := map[string]interface{}{
		"": "AzureDiamond",
	}
	require.Equal(t, v, val.GetValue())
}

func TestStringOfMap(t *testing.T) {
	d := &stringOfMapDefinition{
		path: []string{"status", "dbCredential"},
	}
	val, err := d.Apply(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredential": map[string]interface{}{
					"username": "AzureDiamond",
					"password": "hunter2",
				},
			},
		},
	})
	require.NoError(t, err)
	v := map[string]interface{}{
		"": map[string]string{
			"username": "AzureDiamond",
			"password": "hunter2",
		},
	}
	require.Equal(t, v, val.GetValue())
}

func TestSliceOfStrings(t *testing.T) {
	d := &sliceOfStringsFromPathDefinition{
		sourceValue: "url",
		path:        []string{"status", "bootstrap"},
	}
	val, err := d.Apply(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"bootstrap": []interface{}{
					map[string]interface{}{
						"type": "http",
						"url":  "www.example.com",
					},
					map[string]interface{}{
						"type": "https",
						"url":  "secure.example.com",
					},
				},
			},
		},
	})
	require.NoError(t, err)
	v := []string{"www.example.com", "secure.example.com"}
	require.Equal(t, v, val.GetValue())
}

func TestSliceOfMaps(t *testing.T) {
	d := &sliceOfMapsFromPathDefinition{
		sourceKey:   "type",
		sourceValue: "url",
		path:        []string{"status", "bootstrap"},
	}
	val, err := d.Apply(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"bootstrap": []interface{}{
					map[string]interface{}{
						"type": "http",
						"url":  "www.example.com",
					},
					map[string]interface{}{
						"type": "https",
						"url":  "secure.example.com",
					},
				},
			},
		},
	})
	require.NoError(t, err)
	v := map[string]interface{}{
		"http":  "www.example.com",
		"https": "secure.example.com",
	}
	require.Equal(t, v, val.GetValue())

}
