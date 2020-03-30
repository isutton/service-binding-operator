package envvars

import (
	"errors"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
)

// UnsupportedTypeErr is returned when an unsupported type is encountered.
var UnsupportedTypeErr = errors.New("unsupported type")

// Build returns an environment variable dictionary with an entry for each
// leaf containing a scalar value.
func Build(obj interface{}, path []string) (map[string]string, error) {
	// perform the appropriate action depending on its type; maybe at some point
	// reflection might be required.
	switch val := obj.(type) {
	case map[string]interface{}:
		return buildMap(val, path)
	case []map[string]interface{}:
		return buildSliceOfMap(val, path)
	case string:
		return buildString(val, path), nil
	default:
		return nil, UnsupportedTypeErr
	}
}

// buildEnvVarName returns the environment variable name for the given `path`.
func buildEnvVarName(path []string) string {
	envVar := strings.Join(path, "_")
	envVar = strings.ToUpper(envVar)
	return envVar
}

// buildString returns a map containing the environment variable, named using
// the given `path` and the given `s` value.
func buildString(val string, path []string) map[string]string {
	return map[string]string{
		buildEnvVarName(path): val,
	}
}

// buildMap returns a map containing environment variables for all the leaves
// present in the given `obj` map.
func buildMap(obj map[string]interface{}, path []string) (map[string]string, error) {
	envVars := make(map[string]string)
	for k, v := range obj {
		if err := buildInner(path, k, v, envVars); err != nil {
			return nil, err
		}
	}
	return envVars, nil
}

// buildSliceOfMap returns a map containing environment variables for all the
// leaves present in the given `obj` slice.
func buildSliceOfMap(obj []map[string]interface{}, acc []string) (map[string]string, error) {
	envVars := make(map[string]string)
	for i, v := range obj {
		k := strconv.Itoa(i)
		if err := buildInner(acc, k, v, envVars); err != nil {
			return nil, err
		}
	}
	return envVars, nil
}

// buildInner builds recursively an environment variable map for the given value
// and merges it with the given `envVars` map.
func buildInner(
	path []string,
	key string,
	value interface{},
	envVars map[string]string,
) error {
	if envVar, err := Build(value, append(path, key)); err != nil {
		return err
	} else {
		return mergo.Merge(&envVars, envVar)
	}
}
