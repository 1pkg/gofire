package generators

import (
	"fmt"
	"sort"
	"strings"

	"github.com/1pkg/gofire"
)

type proxy struct {
	driver  Driver
	command gofire.Command
}

func proxify(driver Driver, cmd gofire.Command) (interface{}, error) {
	driver = &cached{Driver: driver}
	if _, err := driver.Output(cmd); err != nil {
		return nil, err
	}
	return proxy{driver: driver, command: cmd}, nil
}

func (p proxy) Package() string {
	return p.command.Package
}

func (p proxy) Function() string {
	return p.command.Function
}

func (p proxy) Doc() string {
	return p.command.Doc
}

func (p proxy) Import() string {
	imports := append(p.driver.Imports(), `"context"`)
	sort.Strings(imports)
	return strings.Join(imports, "\n")
}

func (p proxy) Return() string {
	// collect all return call signature param namep.
	rnames := make([]string, 0, len(p.command.Results))
	for i := range p.command.Results {
		rnames = append(rnames, fmt.Sprintf("o%d", i))
	}
	// append an extra cmd error to the end of return signature.
	rtypes := append(p.command.Results, "error")
	rnames = append(rnames, "err")
	ret := make([]string, 0, len(rnames))
	for i := range rnames {
		ret = append(ret, fmt.Sprintf("%s %s", rnames[i], rtypes[i]))
	}
	return strings.Join(ret, ", ")
}

func (p proxy) Parameters() string {
	parameters := make([]string, 0, len(p.driver.Parameters()))
	for _, p := range p.driver.Parameters() {
		name := p.Name
		if p.Ellipsis {
			name = fmt.Sprintf("%s...", name)
		}
		parameters = append(parameters, name)
	}
	return strings.Join(parameters, ", ")
}

func (p proxy) Vars() string {
	vars := make([]string, 0, len(p.driver.Parameters()))
	groups := make(map[string]bool)
	for _, p := range p.driver.Parameters() {
		vars = append(vars, fmt.Sprintf("var %s %s", p.Name, p.Type.Type()))
		// for each struct group generate separate var too.
		if g := p.Ref.Group(); g != "" && !groups[g] {
			vars = append(vars, fmt.Sprintf("var g%s %s", g, g))
			groups[g] = true
		}
	}
	return strings.Join(vars, "\n")
}

func (p proxy) Body() string {
	out, _ := p.driver.Output(p.command)
	return out
}

func (p proxy) Groups() string {
	// collect all group assigns and append them to generated body.
	var gassigns []string
	for _, p := range p.driver.Parameters() {
		if p.Ref != nil {
			gassigns = append(gassigns, fmt.Sprintf("g%s=%s", *p.Ref, p.Name))
		}
	}
	return strings.Join(gassigns, "\n")
}

func (p proxy) Call() string {
	// define call expression context param aware template.
	var call string
	if p.command.Context {
		call = fmt.Sprintf("%s(ctx, %s)", p.command.Function, p.Parameters())
	} else {
		call = fmt.Sprintf("%s(%s)", p.command.Function, p.Parameters())
	}
	// collect all return call signature param namep.
	rnames := make([]string, 0, len(p.command.Results))
	for i := range p.command.Results {
		rnames = append(rnames, fmt.Sprintf("o%d", i))
	}
	// enrich call expression template with collected return call signature paramp.
	return fmt.Sprintf("%s = %s", strings.Join(rnames, ", "), call)
}

type cached struct {
	Driver
	cache string
}

func (d *cached) Output(cmd gofire.Command) (string, error) {
	if d.cache != "" {
		return d.cache, nil
	}
	out, err := d.Driver.Output(cmd)
	if err != nil {
		return "", nil
	}
	d.cache = out
	return out, nil
}

func (d *cached) Reset() error {
	if err := d.Driver.Reset(); err != nil {
		return err
	}
	d.cache = ""
	return nil
}