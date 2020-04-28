package acc

import (
	"errors"

	"github.com/imdario/mergo"
)

const valuesKey = "values"

// Acc is a value accumulator.
type Acc map[string]interface{}

// UnsupportedTypeErr is returned when an unsupported type is encountered.
var UnsupportedTypeErr = errors.New("unsupported type")

// Accumulate accumulates the `val` value. An error is returned in the case
// `val` contains an unsupported type.
func (a Acc) Accumulate(val interface{}) error {
	b := NewAcc()
	switch v := val.(type) {
	case map[string]interface{}:
		b[valuesKey] = []map[string]interface{}{v}
	case string:
		b[valuesKey] = []string{v}
	case int:
		b[valuesKey] = []int{v}
	case []map[string]interface{}, []string, []int:
		b[valuesKey] = v
	default:
		return UnsupportedTypeErr
	}
	return mergo.Merge(&a, b, mergo.WithAppendSlice, mergo.WithOverride, mergo.WithTypeCheck)
}

// Value returns the accumulated values.
//
// It is guaranteed
func (a Acc) Value() interface{} {
	return a[valuesKey]
}

// NewAcc returns a new value accumulator
func NewAcc() Acc {
	return Acc{}
}
