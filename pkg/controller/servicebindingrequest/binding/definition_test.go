package binding

import (
	"testing"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
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
			description: "string definition",
			expectedValue: &stringDefinition{
				path: []string{"status", "dbCredential", "username"},
			},
			name:  "service.binding/username",
			value: "path={.status.dbCredential.username}",
		},
		{
			description: "map from data field definition#Secret",
			expectedValue: &mapFromDataFieldDefinition{
				kubeClient: nil,
				objectType: secretObjectType,
				path:       []string{"status", "dbCredential", "username"},
			},
			name:  "service.binding/username",
			value: "path={.status.dbCredential.username},objectType=Secret",
		},
		{
			description: "map from data field definition#ConfigMap",
			expectedValue: &mapFromDataFieldDefinition{
				kubeClient: nil,
				objectType: configMapObjectType,
				path:       []string{"status", "dbCredential", "username"},
			},
			name:  "service.binding/username",
			value: "path={.status.dbCredential.username},objectType=ConfigMap",
		},
		{
			description: "string of map definition",
			expectedValue: &stringOfMapDefinition{
				path: []string{"status", "database"},
			},
			name:  "service.binding/username",
			value: "path={.status.database},elementType=map",
		},
		{
			description: "slice of maps from path definition",
			expectedValue: &sliceOfMapsFromPathDefinition{
				path:        []string{"status", "bootstrap"},
				sourceKey:   "type",
				sourceValue: "url",
			},
			name:  "service.binding",
			value: "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
		},
		{
			description: "slice of strings from path definition",
			expectedValue: &sliceOfStringsFromPathDefinition{
				path:        []string{"status", "bootstrap"},
				sourceValue: "url",
			},
			name:  "service.binding",
			value: "path={.status.bootstrap},elementType=sliceOfStrings,sourceValue=url",
		},
	}

	mapper := &definitionMapper{}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			d, err := mapper.MapAnnotation(tc.name, tc.value)
			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, d)
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

func TestSliceOfStringsFromPath(t *testing.T) {
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

func TestSliceOfMapsFromPath(t *testing.T) {
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

func TestMapFromSecretDataField(t *testing.T) {
	f := mocks.NewFake(t, "test-namespace")
	f.AddMockedUnstructuredSecret("dbCredentials-secret")
	d := &mapFromDataFieldDefinition{
		kubeClient: f.FakeDynClient(),
		objectType: secretObjectType,
		path:       []string{"status", "dbCredentials"},
	}
	val, err := d.Apply(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "dbCredentials-secret",
			},
		},
	})
	require.NoError(t, err)
	v := map[string]string{
		"username": "user",
		"password": "password",
	}
	require.Equal(t, v, val.GetValue())
}

func TestMapFromConfigMapDataField(t *testing.T) {
	f := mocks.NewFake(t, "test-namespace")
	f.AddMockedUnstructuredConfigMap("dbCredentials-configMap")
	d := &mapFromDataFieldDefinition{
		kubeClient: f.FakeDynClient(),
		objectType: configMapObjectType,
		path:       []string{"status", "dbCredentials"},
	}
	val, err := d.Apply(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "dbCredentials-configMap",
			},
		},
	})
	require.NoError(t, err)
	v := map[string]string{
		"username": "user",
		"password": "password",
	}
	require.Equal(t, v, val.GetValue())
}
