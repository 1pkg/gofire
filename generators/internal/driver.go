package internal

import (
	"fmt"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

type Driver struct {
	params []generators.Parameter
}

func (d *Driver) Reset() error {
	d.params = nil
	return nil
}

func (d Driver) Imports() []string {
	return []string{}
}

func (d Driver) Parameters() []generators.Parameter {
	return d.params
}

func (d Driver) Last() *generators.Parameter {
	l := len(d.params)
	if l == 0 {
		return nil
	}
	return &d.params[l-1]
}

func (d Driver) Template() string {
	return `
		package {{.Package}}

		import(
			{{.Import}}
		)
		
		{{.Doc}}
		func {{.Function}}(ctx context.Context) ({{.Return}}) {
			{{.Vars}}
			if err = func(ctx context.Context) (err error) {
				{{.Body}}
				{{.Groups}}
				return
			}(ctx); err != nil {
				return
			}
			{{.Call}}
			return
		}
	`
}

func (d *Driver) VisitPlaceholder(p gofire.Placeholder) error {
	d.params = append(d.params, generators.Parameter{
		Name: fmt.Sprintf("p%d", len(d.params)),
		Type: p.Type,
	})
	return nil
}

func (d *Driver) VisitArgument(a gofire.Argument) error {
	// For ellipsis argument we need to produce slice like parameter.
	typ := a.Type
	if a.Ellipsis {
		typ = gofire.TSlice{ETyp: typ}
	}
	d.params = append(d.params, generators.Parameter{
		Name:     fmt.Sprintf("a%d", a.Index),
		Ellipsis: a.Ellipsis,
		Type:     typ,
	})
	return nil
}

func (d *Driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	var gname string
	var gdoc string
	if g != nil {
		gname = g.Name
		gdoc = g.Doc
	}
	name := fmt.Sprintf("%s%s", gname, f.Full)
	doc := fmt.Sprintf("%s %s", gdoc, f.Doc)
	var ref *generators.Reference
	if g != nil {
		ref = generators.NewReference(g.Type.Type(), gname, f.Full)
	}
	d.params = append(d.params, generators.Parameter{
		Name:  name,
		Full:  f.Full,
		Short: f.Short,
		Type:  f.Type,
		Doc:   doc,
		Ref:   ref,
	})
	return nil
}
