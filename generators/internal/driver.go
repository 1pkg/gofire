package internal

import (
	"fmt"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

type Driver struct {
	params []generators.Parameter
}

func (d *Driver) VisitPlaceholder(p gofire.Placeholder) error {
	d.params = append(d.params, generators.Parameter{
		Name: fmt.Sprintf("p%d", len(d.params)),
		Type: p.Type,
	})
	return nil
}

func (d *Driver) VisitArgument(a gofire.Argument) error {
	d.params = append(d.params, generators.Parameter{
		Name: fmt.Sprintf("a%d", a.Index),
		Type: a.Type,
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
	var alt string
	if f.Short != "" {
		alt = fmt.Sprintf("%s%s", gname, f.Short)
	}
	var ref *generators.Reference
	if g != nil {
		ref = generators.NewReference(gname, f.Full)
	}
	d.params = append(d.params, generators.Parameter{
		Name: name,
		Alt:  alt,
		Type: f.Type,
		Doc:  doc,
		Ref:  ref,
	})
	return nil
}

func (d *Driver) Reset() error {
	d.params = nil
	return nil
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

func (d *Driver) Template() string {
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

func cached(p generators.Producer) generators.Producer {
	var (
		out  string
		err  error
		done bool
	)
	return generators.ProducerFunc(func(cmd gofire.Command) (string, error) {
		if !done {
			out, err = p.Output(cmd)
			done = true
		}
		return out, err
	})
}
