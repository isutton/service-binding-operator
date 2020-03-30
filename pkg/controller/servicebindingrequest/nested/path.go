package nested

import (
	"strconv"
	"strings"
)

// Field is the representation of a node in a path.
type Field struct {
	// Name is the field name.
	Name string
	// Index is the integer representation of the field name in the case it can be converted to an
	// integer value.
	Index *int
}

// NewField creates a new field with the given name.
func NewField(name string) Field {
	f := Field{Name: name}
	if i, err := strconv.Atoi(name); err == nil {
		f.Index = &i
	}
	return f
}

// Path represents a path inside a data-structure.
type Path []Field

// Head returns the path head if one exists.
func (p Path) Head() (Field, bool) {
	if len(p) > 0 {
		return p[0], true
	}
	return Field{}, false
}

// Tail returns the path tail if present.
func (p Path) Tail() Path {
	_, exists := p.Head()
	if !exists {
		return Path{}
	}
	return p[1:]
}

// HasTail asserts whether path has a tail.
func (p Path) HasTail() bool {
	return len(p.Tail()) > 0
}

// AdjustedPath adjusts the current path depending on the head element.
func (p Path) AdjustedPath() Path {
	head, exists := p.Head()
	if !exists {
		return Path{}
	}
	if head.Name == "*" {
		return p.Tail()
	}
	return p
}

func (p Path) LastField() (Field, bool) {
	if len(p) > 0 {
		return p[len(p)-1], true
	}
	return Field{}, false
}

func (p Path) BasePath() Path {
	if len(p) > 1 {
		return p[:len(p)-1]
	}
	return Path{}
}

func (p Path) Decompose() (Path, Field) {
	f, _ := p.LastField()
	b := p.BasePath()
	return b, f
}

func (p Path) Clean() Path {
	newPath := make(Path, 0)
	for _, f := range p.AdjustedPath() {
		if f.Index != nil {
			continue
		}
		if f.Name == "*" {
			continue
		}
		newPath = append(newPath, f)
	}
	return newPath
}

// NewPath creates a new path with the given string.
func NewPath(s string) Path {
	parts := strings.Split(s, ".")
	path := make(Path, len(parts))
	for i, p := range parts {
		path[i] = NewField(p)
	}
	return path
}

