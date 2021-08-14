package generators

import "github.com/1pkg/gofire"

type DriverName string

const (
	DriverNameGofire DriverName = "gofire"
	DriverNameFlag  DriverName = "flag"
	DriverNamePFlag DriverName = "pflag"
)

type Parameter struct {
	Name string
	Type gofire.Typ
}

type Driver interface {
	Imports() []string
	Parameters() []Parameter
	Output() ([]byte, error)
	Reset() error
	gofire.Visitor
}
