package binding

import (
	"testing"

	"github.com/stretchr/testify/require"
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
