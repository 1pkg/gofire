package internal

import (
	"fmt"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

type Driver struct {
	params []generators.Parameter
	groups map[string]generators.Parameter
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
		d.groups[gname] = generators.Parameter{
			Name: fmt.Sprintf("g%s", gname),
			Type: g.Type,
		}
	}
	name := fmt.Sprintf("%s%s", gname, f.Full)
	doc := fmt.Sprintf("%s %s", gdoc, f.Doc)
	var alt string
	if f.Short != "" {
		alt = fmt.Sprintf("%s%s", gname, f.Short)
	}
	d.params = append(d.params, generators.Parameter{
		Name:    name,
		Alt:     alt,
		Type:    f.Type,
		Default: f.Default,
		Doc:     doc,
	})
	return nil
}

func (d *Driver) Reset() error {
	d.params = nil
	return nil
}

func (d Driver) Parameters() []generators.Parameter {
	groups := make([]generators.Parameter, 0, len(d.groups))
	for _, p := range d.groups {
		groups = append(groups, p)
	}
	return append(d.params, groups...)
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
