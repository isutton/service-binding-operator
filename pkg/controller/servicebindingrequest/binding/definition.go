package binding

type Definition interface {
	GetObjectType() objectType
	GetElementType() elementType
	GetProjectionIdentifier() string
	GetPath() []string
	GetSourceKey() string
	GetSourceValue() string
}

type definition struct {
	objectType           objectType
	elementType          elementType
	path                 []string
	sourceKey            string
	sourceValue          string
	projectionIdentifier string
}

var _ Definition = (*definition)(nil)

func (d *definition) GetElementType() elementType     { return d.elementType }
func (d *definition) GetObjectType() objectType       { return d.objectType }
func (d *definition) GetPath() []string               { return d.path }
func (d *definition) GetSourceKey() string            { return d.sourceKey }
func (d *definition) GetSourceValue() string          { return d.sourceValue }
func (d *definition) GetProjectionIdentifier() string { return d.projectionIdentifier }

type DefinitionMapper interface {
	Map(val interface{}) (Definition, error)
}

type definitionMapper struct{}

var _ DefinitionMapper = (*definitionMapper)(nil)

func (m *definitionMapper) Map(val interface{}) (Definition, error) {
	return nil, nil
}
