package generators

import (
	"fmt"
	"strings"

	"github.com/1pkg/gofire"
)

type DriverName string

const (
	DriverNameGofire DriverName = "gofire"
	DriverNameFlag  DriverName = "flag"
	DriverNamePFlag DriverName = "pflag"
)

type Reference string

func NewReference(g, f string) *Reference {
	ref := Reference(fmt.Sprintf("%s.%s", g, f))
	return &ref
}

func (g *Reference) Group() string {
	if g == nil {
		return ""
	}
	return strings.Split(string(*g), ".")[0]
}

func (g *Reference) Field() string {
	if g == nil {
		return ""
	}
	return strings.Split(string(*g), ".")[1]
}

type Parameter struct {
	Name     string
	Alt      string
	Type     gofire.Typ
	Ellipsis bool
	Doc      string
	Ref      *Reference
}

type Driver interface {
	Producer
	Imports() []string
	Parameters() []Parameter
	Template() string
	gofire.Visitor
}

type Producer interface {
	Output(gofire.Command) (string, error)
	Reset() error
}

type BaseDriver struct {
	params []Parameter
}

func (d *BaseDriver) Reset() error {
	d.params = nil
	return nil
}

func (d BaseDriver) Parameters() []Parameter {
	return d.params
}

func (d BaseDriver) Last() *Parameter {
	l := len(d.params)
	if l == 0 {
		return nil
	}
	return &d.params[l-1]
}

func (d BaseDriver) Template() string {
	return `
		package {{.Package}}

		import(
			{{.Import}}
		)
		
		func Command{{.Function}}(ctx context.Context) ({{.Return}}) {
			{{.Vars}}
			if err = func(ctx context.Context) (err error) {
				{{.Body}}
				return
			}(ctx); err != nil {
				return
			}
			{{.Call}}
			return
		}
	`
}

func (d BaseDriver) VisitPlaceholder(p gofire.Placeholder) error {
	d.params = append(d.params, Parameter{
		Name: fmt.Sprintf("p%d", len(d.params)),
		Type: p.Type,
	})
	return nil
}

func (d *BaseDriver) VisitArgument(a gofire.Argument) error {
	// For ellipsis argument we need to produce slice like parameter.
	typ := a.Type
	if a.Ellipsis {
		typ = gofire.TSlice{ETyp: typ}
	}
	d.params = append(d.params, Parameter{
		Name:     fmt.Sprintf("a%d", a.Index),
		Ellipsis: a.Ellipsis,
		Type:     typ,
	})
	return nil
}

func (d *BaseDriver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	var gname string
	var gdoc string
	if g != nil {
		gname = g.Name
		gdoc = g.Doc
	}
	name := fmt.Sprintf("%s%s", gname, f.Full)
	doc := fmt.Sprintf("%s %s", gdoc, f.Doc)
	var alt string
	if f.Short != "" {
		alt = fmt.Sprintf("%s%s", gname, f.Short)
	}
	var ref *Reference
	if g != nil {
		ref = NewReference(gname, f.Full)
	}
	d.params = append(d.params, Parameter{
		Name: name,
		Alt:  alt,
		Type: f.Type,
		Doc:  doc,
		Ref:  ref,
	})
	return nil
}
