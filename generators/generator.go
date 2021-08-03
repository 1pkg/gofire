package generators

import (
	"context"
	"fmt"
	"go/format"
	"io"
	"regexp"
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
func Generate(ctx context.Context, driver DriverName, cmd gofire.Command, w io.Writer) error {
	importStr, parametersStr, definitionsStr, driverCode, err := applyDriver(ctx, driver, cmd)
	if err != nil {
		return err
	}
	callExprTemplate, retSign := callSignature(cmd)
	callExpr := fmt.Sprintf(callExprTemplate, parametersStr)
	src := fmt.Sprintf(
		`
			package %s

			import(
				%s
			)
			
			func Command%s(ctx context.Context) (%s) {
				%s
				if err = func(ctx context.Context) (err error) {
					%s
					return
				}(ctx); err != nil {
					return
				}
				%s
				return
			}
		`,
		cmd.Pckg,
		importStr,
		cmd.Name,
		retSign,
		definitionsStr,
		driverCode,
		callExpr,
	)
	b, err := format.Source([]byte(src))
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func applyDriver(ctx context.Context, name DriverName, cmd gofire.Command) (imports string, parameters string, definitions string, out string, err error) {
	driversMu.Lock()
	defer driversMu.Unlock()
	driver, ok := drivers[name]
	if !ok {
		err = fmt.Errorf("unknown driver %q (forgotten import?)", name)
		return
	}
	if err = driver.Reset(); err != nil {
		return
	}
	if err = cmd.Accept(driver); err != nil {
		return
	}
	imports = strings.Join(driver.Imports(), "\n")
	pnames := make([]string, 0, len(driver.Parameters()))
	for _, p := range driver.Parameters() {
		pnames = append(pnames, p.Name)
	}
	parameters = strings.Join(pnames, ", ")
	pdefinitions := make([]string, 0, len(driver.Parameters()))
	for _, p := range driver.Parameters() {
		pdefinitions = append(pdefinitions, fmt.Sprintf("var %s %s", p.Name, p.Type.Type()))
	}
	definitions = strings.Join(pdefinitions, "\n")
	parameters = strings.Join(pnames, ", ")
	out = strings.Trim(strip.ReplaceAllString(string(driver.Output()), "\n"), "\n\t ")
	return
}

func callSignature(cmd gofire.Command) (callExprTemplate string, retSign string) {
	// define call expression context param aware template.
	if cmd.Context {
		callExprTemplate = fmt.Sprintf("%s(ctx, %%s)", cmd.Name)
	} else {
		callExprTemplate = fmt.Sprintf("%s(%%s)", cmd.Name)
	}
	// collect all return call signature param names.
	rnames := make([]string, 0, len(cmd.Returns))
	for i := range cmd.Returns {
		rnames = append(rnames, fmt.Sprintf("o%d", i))
	}
	// enrich call expression template with collected return call signature params.
	callExprTemplate = fmt.Sprintf("%s = %s", strings.Join(rnames, ", "), callExprTemplate)
	// append an extra cmd error to the end of return signature.
	rtypes := append(cmd.Returns, "error")
	rnames = append(rnames, "err")
	for i := range rnames {
		retSign += fmt.Sprintf("%s %s, ", rnames[i], rtypes[i])
	}
	return
}
