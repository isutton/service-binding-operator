package binding

import (
	"testing"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestDefinitionMapperInvalidAnnotation(t *testing.T) {
	type args struct {
		description string
		opts        DefinitionMapperOptions
	}

	testCases := []args{
		{
			description: "prefix is service.binding but not followed by / or end of string",
			opts: &annotationToDefinitionMapperOptions{
				name: "service.bindingtrololol",
			},
		},
		{
			description: "invalid path",
			opts: &annotationToDefinitionMapperOptions{
				name:  "service.binding",
				value: "path=.status.secret",
			},
		},
		{
			description: "invalid path",
			opts: &annotationToDefinitionMapperOptions{
				name:  "service.binding",
				value: "path=.status.secret}",
			},
		},
		{
			description: "invalid path",
			opts: &annotationToDefinitionMapperOptions{
				name:  "service.binding",
				value: "path={.status.secret",
			},
		},
		{
			description: "other prefix supplied",
			opts: &annotationToDefinitionMapperOptions{
				name: "other.prefix",
			},
		},
	}

	mapper := &AnnotationToDefinitionMapper{}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := mapper.Map(tc.opts)
			require.Error(t, err)
		})
	}
}

func TestDefinitionMapperValidAnnotations(t *testing.T) {
	type args struct {
		description   string
		expectedValue Definition
		options       DefinitionMapperOptions
	}

	testCases := []args{
		{
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding/username",
				value: "path={.status.dbCredential.username}",
			},
			description: "string definition",
			expectedValue: &stringDefinition{
				outputName: "username",
				path:       []string{"status", "dbCredential", "username"},
			},
		},
		{
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding/anotherUsernameField",
				value: "path={.status.dbCredential.username}",
			},
			description: "string definition",
			expectedValue: &stringDefinition{
				outputName: "anotherUsernameField",
				path:       []string{"status", "dbCredential", "username"},
			},
		},
		{
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding",
				value: "path={.status.dbCredential.username}",
			},
			description: "string definition with default username",
			expectedValue: &stringDefinition{
				outputName: "username",
				path:       []string{"status", "dbCredential", "username"},
			},
		},
		{
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding/username",
				value: "path={.status.dbCredential.username},objectType=Secret",
			},
			description: "map from data field definition#Secret",
			expectedValue: &mapFromDataFieldDefinition{
				kubeClient: nil,
				objectType: secretObjectType,
				outputName: "username",
				path:       []string{"status", "dbCredential", "username"},
			},
		},
		{
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding/anotherUsernameField",
				value: "path={.status.dbCredential.username},objectType=Secret",
			},
			description: "map from data field definition#Secret",
			expectedValue: &mapFromDataFieldDefinition{
				kubeClient: nil,
				objectType: secretObjectType,
				outputName: "anotherUsernameField",
				path:       []string{"status", "dbCredential", "username"},
			},
		},
		{
			description: "map from data field definition#Secret",
			expectedValue: &mapFromDataFieldDefinition{
				kubeClient: nil,
				objectType: secretObjectType,
				outputName: "username",
				path:       []string{"status", "dbCredential", "username"},
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding",
				value: "path={.status.dbCredential.username},objectType=Secret",
			},
		},
		{
			description: "map from data field definition#ConfigMap",
			expectedValue: &mapFromDataFieldDefinition{
				kubeClient: nil,
				objectType: configMapObjectType,
				outputName: "username",
				path:       []string{"status", "dbCredential", "username"},
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding/username",
				value: "path={.status.dbCredential.username},objectType=ConfigMap",
			},
		},
		{
			description: "map from data field definition#ConfigMap",
			expectedValue: &mapFromDataFieldDefinition{
				kubeClient: nil,
				objectType: configMapObjectType,
				outputName: "anotherUsernameField",
				path:       []string{"status", "dbCredential", "username"},
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding/anotherUsernameField",
				value: "path={.status.dbCredential.username},objectType=ConfigMap",
			},
		},
		{
			description: "map from data field definition#ConfigMap",
			expectedValue: &mapFromDataFieldDefinition{
				kubeClient: nil,
				objectType: configMapObjectType,
				outputName: "username",
				path:       []string{"status", "dbCredential", "username"},
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding",
				value: "path={.status.dbCredential.username},objectType=ConfigMap",
			},
		},
		{
			description: "string of map definition",
			expectedValue: &stringOfMapDefinition{
				outputName: "database",
				path:       []string{"status", "database"},
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding/database",
				value: "path={.status.database},elementType=map",
			},
		},
		{
			description: "string of map definition",
			expectedValue: &stringOfMapDefinition{
				outputName: "anotherDatabaseField",
				path:       []string{"status", "database"},
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding/anotherDatabaseField",
				value: "path={.status.database},elementType=map",
			},
		},
		{
			description: "string of map definition",
			expectedValue: &stringOfMapDefinition{
				outputName: "database",
				path:       []string{"status", "database"},
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding",
				value: "path={.status.database},elementType=map",
			},
		},
		{
			description: "slice of maps from path definition",
			expectedValue: &sliceOfMapsFromPathDefinition{
				outputName:  "bootstrap",
				path:        []string{"status", "bootstrap"},
				sourceKey:   "type",
				sourceValue: "url",
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding",
				value: "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
			},
		},
		{
			description: "slice of maps from path definition",
			expectedValue: &sliceOfMapsFromPathDefinition{
				outputName:  "anotherBootstrapField",
				path:        []string{"status", "bootstrap"},
				sourceKey:   "type",
				sourceValue: "url",
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding/anotherBootstrapField",
				value: "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
			},
		},
		{
			description: "slice of strings from path definition",
			expectedValue: &sliceOfStringsFromPathDefinition{
				outputName:  "bootstrap",
				path:        []string{"status", "bootstrap"},
				sourceValue: "url",
			},
			options: &annotationToDefinitionMapperOptions{
				name:  "service.binding",
				value: "path={.status.bootstrap},elementType=sliceOfStrings,sourceValue=url",
			},
		},
	}

	mapper := &AnnotationToDefinitionMapper{}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			d, err := mapper.Map(tc.options)
			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, d)
		})
	}
}

func TestStringDefinition(t *testing.T) {
	type args struct {
		description   string
		outputName    string
		path          []string
		expectedValue interface{}
	}

	testCases := []args{
		{
			description: "outputName informed",
			outputName:  "username",
			path:        []string{"status", "dbCredentials", "username"},
			expectedValue: map[string]interface{}{
				"username": "AzureDiamond",
			},
		},
		{
			description: "outputName empty",
			path:        []string{"status", "dbCredentials", "username"},
			expectedValue: map[string]interface{}{
				"username": "AzureDiamond",
			},
		},
	}

	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"username": "AzureDiamond",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			d := &stringDefinition{
				outputName: tc.outputName,
				path:       tc.path,
			}
			val, err := d.Apply(u)
			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, val.GetValue())

		})
	}
}

func TestStringOfMap(t *testing.T) {
	type args struct {
		description   string
		outputName    string
		path          []string
		expectedValue interface{}
		object        *unstructured.Unstructured
	}

	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"username": "AzureDiamond",
					"password": "hunter2",
				},
			},
		},
	}

	expectedValue := map[string]interface{}{
		"dbCredentials": map[string]interface{}{
			"username": "AzureDiamond",
			"password": "hunter2",
		},
	}

	testCases := []args{
		{
			description:   "outputName informed",
			expectedValue: expectedValue,
			object:        u,
			outputName:    "dbCredentials",
			path:          []string{"status", "dbCredentials"},
		},
		{
			description:   "outputName empty",
			expectedValue: expectedValue,
			object:        u,
			outputName:    "",
			path:          []string{"status", "dbCredentials"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			d := &stringOfMapDefinition{
				outputName: tc.outputName,
				path:       tc.path,
			}
			val, err := d.Apply(tc.object)
			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, val.GetValue())
		})
	}
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
