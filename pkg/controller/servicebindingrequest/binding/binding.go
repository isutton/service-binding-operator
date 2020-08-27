package binding

type ProjectionMapper interface {
	Map([]Value) ([]Projection, error)
}

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

type Projection interface {
	Apply(obj map[string]interface{}) (map[string]interface{}, error)
}

var _ Projection = (*projection)(nil)

type projection struct{}

func (p *projection) Apply(
	obj map[string]interface{},
) (map[string]interface{}, error) {
	return nil, nil
}

type projectionSorter struct{}

var _ ProjectionMapper = (*projectionSorter)(nil)

func (p *projectionSorter) Map([]Value) ([]Projection, error) {
	return nil, nil
}
