package internal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

func init() {
	generators.Bind = func(driver generators.Driver, cmd gofire.Command) (interface{}, error) {
		driver = cached(driver)
		_, err := driver.Output(cmd)
		if err != nil {
			return nil, err
		}
		return generator{driver: driver, command: cmd}, nil
	}
}

type generator struct {
	driver  generators.Driver
	command gofire.Command
}

func (g generator) Package() string {
	return g.command.Package
}

func (g generator) Function() string {
	return g.command.Function
}

func (g generator) Doc() string {
	return g.command.Doc
}

func (g generator) Import() string {
	imports := append(g.driver.Imports(), `"context"`)
	sort.Strings(imports)
	return strings.Join(imports, "\n")
}

func (g generator) Return() string {
	// collect all return call signature param names.
	rnames := make([]string, 0, len(g.command.Results))
	for i := range g.command.Results {
		rnames = append(rnames, fmt.Sprintf("o%d", i))
	}
	// append an extra cmd error to the end of return signature.
	rtypes := append(g.command.Results, "error")
	rnames = append(rnames, "err")
	ret := make([]string, 0, len(rnames))
	for i := range rnames {
		ret = append(ret, fmt.Sprintf("%s %s", rnames[i], rtypes[i]))
	}
	return strings.Join(ret, ", ")
}

func (g generator) Parameters() string {
	parameters := make([]string, 0, len(g.driver.Parameters()))
	for _, p := range g.driver.Parameters() {
		parameters = append(parameters, p.Name)
	}
	return strings.Join(parameters, ", ")
}

func (g generator) Vars() string {
	vars := make([]string, 0, len(g.driver.Parameters()))
	for _, p := range g.driver.Parameters() {
		vars = append(vars, fmt.Sprintf("var %s %s", p.Name, p.Type.Type()))
	}
	return strings.Join(vars, "\n")
}

func (g generator) Body() string {
	out, _ := g.driver.Output(g.command)
	return out
}

func (g generator) Call() string {
	// define call expression context param aware template.
	var call string
	if g.command.Context {
		call = fmt.Sprintf("%s(ctx, %s)", g.command.Function, g.Parameters())
	} else {
		call = fmt.Sprintf("%s(%s)", g.command.Function, g.Parameters())
	}
	// collect all return call signature param names.
	rnames := make([]string, 0, len(g.command.Results))
	for i := range g.command.Results {
		rnames = append(rnames, fmt.Sprintf("o%d", i))
	}
	// enrich call expression template with collected return call signature params.
	return fmt.Sprintf("%s = %s", strings.Join(rnames, ", "), call)
}
