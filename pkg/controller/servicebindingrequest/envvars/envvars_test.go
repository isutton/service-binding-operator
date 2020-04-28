package envvars

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
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

	expected := map[string]string{
		"STATUS_LISTENERS_0_TYPE":             "secure",
		"STATUS_LISTENERS_0_ADDRESSES_0_HOST": "my-cluster-kafka-bootstrap.coffeeshop.svc",
		"STATUS_LISTENERS_0_ADDRESSES_0_PORT": "9093",
	}

	actual, err := Build(src, []string{})
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	tmpl, err := template.New("specific").
		Parse(`{{ index . "status" "listeners" 0 "type" }}`)
	require.NoError(t, err)

	buf := bytes.NewBuffer(nil)
	err = tmpl.Execute(buf, src)

	require.NoError(t, err)
	require.Equal(t, "secure", buf.String())
}
