package generators

import "github.com/1pkg/gofire"

type DriverName string

const (
	DriverNameGofire DriverName = "gofire"
)

type Parameter struct {
	Name string
	Type gofire.Typ
}

type Driver interface {
	Imports() []string
	Parameters() []Parameter
	Output() []byte
	Reset() error
	gofire.Visitor
}
