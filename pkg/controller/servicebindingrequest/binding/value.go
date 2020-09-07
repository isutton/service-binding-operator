package binding

type Value interface {
	GetValue() interface{}
	GetProjectionIdentifier() string
}

type value struct {
	v interface{}
	i string
}

var _ Value = (*value)(nil)

func (v *value) GetValue() interface{} {
	return v.v
}

func (v *value) GetProjectionIdentifier() string {
	return v.i
}
