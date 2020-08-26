package binding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProducerMapper(t *testing.T) {
	type args struct {
		description   string
		definition    Definition
		expectedValue interface{}
	}

	testCases := []args{
		{
			definition: &definition{
				elementType: stringElementType,
				objectType:  stringObjectType,
				path:        []string{"status", "dbCredential", "username"},
				sourceKey:   "",
				sourceValue: "",
			},
			expectedValue: &stringProducer{},
		},
		{
			definition: &definition{
				elementType: mapElementType,
				objectType:  stringObjectType,
				path:        []string{"status", "dbCredential"},
				sourceKey:   "",
				sourceValue: "",
			},
			expectedValue: &stringOfMapProducer{},
		},
		{
			definition: &definition{
				elementType: sliceOfMapsElementType,
				objectType:  stringObjectType,
				path:        []string{"status", "bootstrap"},
				sourceKey:   "type",
				sourceValue: "url",
			},
			expectedValue: &sliceOfMapsFromPathProducer{},
		},
		{
			definition: &definition{
				elementType: sliceOfStringsElementType,
				objectType:  stringObjectType,
				path:        []string{"status", "bootstrap"},
				sourceKey:   "",
				sourceValue: "url",
			},
			expectedValue: &sliceOfStringsFromPathProducer{},
		},
	}

	mapper := &producerMapper{}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			producers, err := mapper.Map([]Definition{tc.definition})
			require.NoError(t, err)
			require.Len(t, producers, 1)
			require.IsType(t, tc.expectedValue, producers[0])
		})
	}
}
