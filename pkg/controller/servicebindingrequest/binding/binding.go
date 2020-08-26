package binding

const (
	// configMapObjectType indicates the path contains a name for a ConfigMap containing the binding
	// data.
	configMapObjectType objectType = "ConfigMap"
	// secretObjectType indicates the path contains a name for a Secret containing the binding data.
	secretObjectType objectType = "Secret"
	// stringObjectType indicates the path contains a value string.
	stringObjectType objectType = "string"
	// emptyObjectType is used as default value when the objectType key is present in the string
	// provided by the user but no value has been provided; can be used by the user to force the
	// system to use the default objectType.
	// emptyObjectType objectType = ""

	// mapElementType indicates the value found at path is a map[string]interface{}.
	mapElementType elementType = "map"
	// sliceOfMapsElementType indicates the value found at path is a slice of maps.
	sliceOfMapsElementType elementType = "sliceOfMaps"
	// sliceOfStringsElementType indicates the value found at path is a slice of strings.
	sliceOfStringsElementType elementType = "sliceOfStrings"
	// stringElementType indicates the value found at path is a string.
	stringElementType elementType = "string"
)

type objectType string

type elementType string

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
