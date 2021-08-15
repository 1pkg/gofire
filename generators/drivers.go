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
	Alt  string
	Type gofire.Typ
}

type Driver interface {
	Imports() []string
	Parameters() []Parameter
	Output() ([]byte, error)
	Reset() error
	gofire.Visitor
}

type Visitor struct {
	params []Parameter
}

func (d *Visitor) VisitPlaceholder(p gofire.Placeholder) error {
	d.params = append(d.params, Parameter{
		Name: fmt.Sprintf("p%d", len(d.params)),
		Type: p.Type,
	})
	return nil
}

func (d *Visitor) VisitArgument(a gofire.Argument) error {
	d.params = append(d.params, Parameter{
		Name: fmt.Sprintf("a%d", a.Index),
		Type: a.Type,
	})
	return nil
}

func (d *Visitor) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	var gname string
	if g != nil {
		gname = g.Name
	}
	name := fmt.Sprintf("f%s%s", gname, f.Full)
	var alt string
	if f.Short != "" {
		alt = fmt.Sprintf("f%s%s", gname, f.Short)
	}
	d.params = append(d.params, Parameter{
		Name: name,
		Alt:  alt,
		Type: f.Type,
	})
	return nil
}

func (d Visitor) Parameters() []Parameter {
	return d.params
}

func (d *Visitor) Reset() error {
	d.params = nil
	return nil
}

func (d Visitor) Last() *Parameter {
	l := len(d.params)
	if l == 0 {
		return nil
	}
	return &d.params[l-1]
}
