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
	if g != nil {
		gname = g.Name
		d.groups[gname] = generators.Parameter{
			Name: fmt.Sprintf("g%s", gname),
			Type: g.Type,
		}
	}
	name := fmt.Sprintf("f%s%s", gname, f.Full)
	var alt string
	if f.Short != "" {
		alt = fmt.Sprintf("f%s%s", gname, f.Short)
	}
	d.params = append(d.params, generators.Parameter{
		Name: name,
		Alt:  alt,
		Type: f.Type,
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

func cached(driver generators.Driver) generators.Driver {
	var (
		out  string
		err  error
		done bool
		d    cd
	)
	d.Driver = driver
	d.f = func() (string, error) {
		if !done {
			out, err = driver.Output()
			done = true
		}
		return out, err
	}
	return d
}

type cd struct {
	generators.Driver
	f func() (string, error)
}

func (d cd) Output() (string, error) {
	return d.f()
}
