package generators

import (
	"context"
	"fmt"
	"go/format"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/1pkg/gofire"
)

var (
	driversMu sync.Mutex
	drivers   = make(map[DriverName]Driver)
	strip     = regexp.MustCompile(`\n(\s)+\n`)
)

// Register makes a generator driver available by the provided name.
// If Register is called twice with the same name or if driver is nil, it panics.
func Register(name DriverName, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic(fmt.Errorf("register called twice for driver %q", name))
	}
	drivers[name] = driver
}

// Generate generates cli command using provided driver to provided writer output.
func Generate(ctx context.Context, drivern DriverName, cmd gofire.Command, w io.Writer) error {
	driversMu.Lock()
	driver, ok := drivers[drivern]
	defer driversMu.Unlock()
	if !ok {
		return fmt.Errorf("unknown driver %q (forgotten import?)", drivern)
	}
	if err := driver.Reset(); err != nil {
		return err
	}
	if err := cmd.Accept(driver); err != nil {
		return err
	}
	out, err := driver.Output()
	if err != nil {
		return err
	}
	g := Generator{driver: driver, cmd: cmd}
	rsrc := fmt.Sprintf(driver.Template(g), strings.Trim(strip.ReplaceAllString(string(out), "\n"), "\n\t "))
	src, err := format.Source([]byte(rsrc))
	if err != nil {
		return err
	}
	if _, err := w.Write(src); err != nil {
		return err
	}
	return nil
}

type Generator struct {
	driver Driver
	cmd    gofire.Command
}

func (g Generator) Package() string {
	return g.cmd.Package
}

func (g Generator) Function() string {
	return g.cmd.Function
}

func (g Generator) Doc() string {
	return g.cmd.Doc
}

func (g Generator) Import() string {
	imports := append(g.driver.Imports(), `"context"`)
	sort.Strings(imports)
	return strings.Join(imports, "\n")
}

func (g Generator) Return() string {
	// collect all return call signature param names.
	rnames := make([]string, 0, len(g.cmd.Results))
	for i := range g.cmd.Results {
		rnames = append(rnames, fmt.Sprintf("o%d", i))
	}
	// append an extra cmd error to the end of return signature.
	rtypes := append(g.cmd.Results, "error")
	rnames = append(rnames, "err")
	ret := make([]string, 0, len(rnames))
	for i := range rnames {
		ret = append(ret, fmt.Sprintf("%s %s", rnames[i], rtypes[i]))
	}
	return strings.Join(ret, ", ")
}

func (g Generator) Parameters() string {
	parameters := make([]string, 0, len(g.driver.Parameters()))
	for _, p := range g.driver.Parameters() {
		parameters = append(parameters, p.Name)
	}
	return strings.Join(parameters, ", ")
}

func (g Generator) Vars() string {
	vars := make([]string, 0, len(g.driver.Parameters()))
	for _, p := range g.driver.Parameters() {
		vars = append(vars, fmt.Sprintf("var %s %s", p.Name, p.Type.Type()))
	}
	return strings.Join(vars, "\n")
}

func (g Generator) Call() string {
	// define call expression context param aware template.
	var call string
	if g.cmd.Context {
		call = fmt.Sprintf("%s(ctx, %s)", g.cmd.Function, g.Parameters())
	} else {
		call = fmt.Sprintf("%s(%s)", g.cmd.Function, g.Parameters())
	}
	// collect all return call signature param names.
	rnames := make([]string, 0, len(g.cmd.Results))
	for i := range g.cmd.Results {
		rnames = append(rnames, fmt.Sprintf("o%d", i))
	}
	// enrich call expression template with collected return call signature params.
	return fmt.Sprintf("%s = %s", strings.Join(rnames, ", "), call)
}
