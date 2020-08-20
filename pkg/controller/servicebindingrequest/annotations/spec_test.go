package annotations

import (
	"testing"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestModel(t *testing.T) {
	t.Run("single token can't be parsed to a model, should return error", func(t *testing.T) {
		_, err := buildModel("path")
		require.Error(t, err)
	})

	t.Run("valid model, default elementType and objectType", func(t *testing.T) {
		for _, e := range []string{
			"path={.status.dbCredentials.username}",
			"path={.status.dbCredentials.username},objectType=",
			"path={.status.dbCredentials.username},objectType=string",
		} {
			t.Run(e, func(t *testing.T) {
				m, err := buildModel(e)
				require.NoError(t, err)

				assert.Equal(t, "{.status.dbCredentials.username}", m.path)
				assert.Equal(t, stringElementType, m.elementType)
				assert.Equal(t, stringObjectType, m.objectType)

				obj := unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"dbCredentials": map[string]interface{}{
								"username": "AzureDiamond",
							},
						},
					},
				}

				val, err := m.produce(obj, nil)
				require.NoError(t, err)
				require.Equal(t, "AzureDiamond", val)
			})
		}

		for _, e := range []string{
			"path={.status.dbCredentials}",
			"path={.status.dbCredentials},objectType=",
			"path={.status.dbCredentials},objectType=string",
		} {
			t.Run(e, func(t *testing.T) {
				m, err := buildModel(e)
				require.NoError(t, err)

				assert.Equal(t, "{.status.dbCredentials}", m.path)
				assert.Equal(t, stringElementType, m.elementType)
				assert.Equal(t, stringObjectType, m.objectType)

				obj := unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"dbCredentials": map[string]interface{}{
								"username": "AzureDiamond",
							},
						},
					},
				}
				val, err := m.produce(obj, nil)
				require.NoError(t, err)
				require.Equal(t, "map[username:AzureDiamond]", val)
			})
		}
	})

	t.Run("valid model, elementType map", func(t *testing.T) {
		for _, e := range []string{
			"path={.status.dbCredentials},elementType=map",
		} {
			t.Run(e, func(t *testing.T) {
				m, err := buildModel(e)
				require.NoError(t, err)

				assert.Equal(t, "{.status.dbCredentials}", m.path)
				assert.Equal(t, mapElementType, m.elementType)
				assert.Equal(t, stringObjectType, m.objectType)

				obj := unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"dbCredentials": map[string]interface{}{
								"username": "AzureDiamond",
							},
						},
					},
				}
				val, err := m.produce(obj, nil)
				require.NoError(t, err)
				expected := map[string]string{"username": "AzureDiamond"}
				require.Equal(t, expected, val)
			})
		}
	})

	t.Run("valid model, Secret objectType", func(t *testing.T) {
		for _, e := range []string{
			"path={.status.dbCredentials},objectType=Secret",
			"path={.status.dbCredentials},objectType=Secret,elementType=map",
		} {
			m, err := buildModel(e)
			require.NoError(t, err)

			assert.Equal(t, "{.status.dbCredentials}", m.path)
			assert.Equal(t, mapElementType, m.elementType)
			assert.Equal(t, secretObjectType, m.objectType)

			obj := unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"namespace": "test-namespace",
					},
					"status": map[string]interface{}{
						"dbCredentials": "external-secret",
					},
				},
			}

			f := mocks.NewFake(t, "test-namespace")
			f.AddMockedUnstructuredSecret("external-secret")

			kubeClient := f.FakeDynClient()
			val, err := m.produce(obj, kubeClient)
			require.NoError(t, err)
			expected := map[string]string{
				"username": "user",
				"password": "password",
			}
			require.Equal(t, expected, val)
		}
	})

	t.Run("valid model, ConfigMap objectType", func(t *testing.T) {
		for _, e := range []string{
			"path={.status.dbConfiguration},objectType=ConfigMap",
			"path={.status.dbConfiguration},objectType=ConfigMap,elementType=map",
		} {
			m, err := buildModel(e)
			require.NoError(t, err)

			assert.Equal(t, "{.status.dbConfiguration}", m.path)
			assert.Equal(t, mapElementType, m.elementType)
			assert.Equal(t, configMapObjectType, m.objectType)

			obj := unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"namespace": "test-namespace",
					},
					"status": map[string]interface{}{
						"dbConfiguration": "external-configmap",
					},
				},
			}

			f := mocks.NewFake(t, "test-namespace")
			f.AddMockedUnstructuredConfigMap("external-configmap")

			kubeClient := f.FakeDynClient()
			val, err := m.produce(obj, kubeClient)
			require.NoError(t, err)
			expected := map[string]string{
				"username": "user",
				"password": "password",
			}
			require.Equal(t, expected, val)
		}
	})

	t.Run("valid model, ConfigMap objectType, return single value key", func(t *testing.T) {
		for _, e := range []string{
			"path={.status.dbConfiguration},objectType=ConfigMap,sourceKey=username",
			"path={.status.dbConfiguration},objectType=ConfigMap,sourceKey=username,elementType=string",
		} {
			m, err := buildModel(e)
			require.NoError(t, err)

			assert.Equal(t, "{.status.dbConfiguration}", m.path)
			assert.Equal(t, stringElementType, m.elementType)
			assert.Equal(t, configMapObjectType, m.objectType)
			assert.Equal(t, "username", m.sourceKey)

			obj := unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"namespace": "test-namespace",
					},
					"status": map[string]interface{}{
						"dbConfiguration": "external-configmap",
					},
				},
			}

			f := mocks.NewFake(t, "test-namespace")
			f.AddMockedUnstructuredConfigMap("external-configmap")

			kubeClient := f.FakeDynClient()
			val, err := m.produce(obj, kubeClient)
			require.NoError(t, err)
			require.Equal(t, "user", val)
		}
	})

	t.Run("valid model, sliceOfMaps elementType", func(t *testing.T) {
		for _, e := range []string{
			"path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
			"path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url,objectType=string",
		} {
			m, err := buildModel(e)
			require.NoError(t, err)

			assert.Equal(t, "{.status.bootstrap}", m.path)
			assert.Equal(t, sliceOfMapsElementType, m.elementType)
			assert.Equal(t, stringObjectType, m.objectType)

			obj := unstructured.Unstructured{
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
			}

			val, err := m.produce(obj, nil)
			require.NoError(t, err)
			expected := map[string]interface{}{
				"http":  "www.example.com",
				"https": "secure.example.com",
			}
			require.Equal(t, expected, val)
		}
	})

	t.Run("valid model, sliceOfStrings elementType", func(t *testing.T) {
		for _, e := range []string{
			"path={.status.bootstrap},elementType=sliceOfStrings,sourceValue=url",
			"path={.status.bootstrap},elementType=sliceOfStrings,sourceValue=url,objectType=string",
		} {
			m, err := buildModel(e)
			require.NoError(t, err)

			assert.Equal(t, "{.status.bootstrap}", m.path)
			assert.Equal(t, sliceOfStringsElementType, m.elementType)
			assert.Equal(t, stringObjectType, m.objectType)

			obj := unstructured.Unstructured{
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
			}

			val, err := m.produce(obj, nil)
			require.NoError(t, err)
			expected := []string{"www.example.com", "secure.example.com"}
			require.Equal(t, expected, val)
		}
	})
}
