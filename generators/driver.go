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

type Templater interface {
	Template(Generator) string
}

type Driver interface {
	Templater
	gofire.Visitor
	Imports() []string
	Parameters() []Parameter
	Output() (string, error)
	Reset() error
}

type Visitor struct {
	params []Parameter
	groups map[string]Parameter
}

func (d *Visitor) Template(g Generator) string {
	return fmt.Sprintf(
		`
			package %s

			import(
				%s
			)
			
			func Command%s(ctx context.Context) (%s) {
				%s
				if err = func(ctx context.Context) (err error) {
					%%s
					return
				}(ctx); err != nil {
					return
				}
				%s
				return
			}
		`,
		g.Package(),
		g.Import(),
		g.Function(),
		g.Return(),
		g.Vars(),
		g.Call(),
	)
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
		d.groups[gname] = Parameter{
			Name: fmt.Sprintf("g%s", gname),
			Type: g.Type,
		}
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
	groups := make([]Parameter, 0, len(d.groups))
	for _, p := range d.groups {
		groups = append(groups, p)
	}
	return append(d.params, groups...)
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