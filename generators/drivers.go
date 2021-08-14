package generators

import (
	"fmt"

	"github.com/1pkg/gofire"
)

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

type BaseDriver struct {
	Params []Parameter
}

func (d *BaseDriver) VisitPlaceholder(p gofire.Placeholder) error {
	d.Params = append(d.Params, Parameter{
		Name: fmt.Sprintf("p%d", len(d.Params)),
		Type: p.Type,
	})
	return nil
}

func (d BaseDriver) Parameters() []Parameter {
	return d.Params
}

func (d *BaseDriver) Reset() error {
	d.Params = nil
	return nil
}
