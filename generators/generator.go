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
	driversMu sync.RWMutex
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
	importStr, parametersStr, driverCode, err := applyDriver(ctx, driver, cmd)
	if err != nil {
		return err
	}
	callExpr, retSign, docsStr := callSignature(cmd)
	src := fmt.Sprintf(
		`
			package %s

			import(
				%s
			)
			
			%s
			func Command%s(ctx context.Context) (%s) {
				if err = func(ctx context.Context) error {
					%s
				}(ctx); err != nil {
					return
				}
				%s
				return
			}
		`,
		cmd.Pckg,
		importStr,
		docsStr,
		cmd.Name,
		retSign,
		driverCode,
		fmt.Sprintf(callExpr, parametersStr),
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

func applyDriver(ctx context.Context, name DriverName, cmd gofire.Command) (imports string, parameters string, out string, err error) {
	driversMu.RLock()
	driver, ok := drivers[name]
	driversMu.RUnlock()
	if !ok {
		err = fmt.Errorf("unknown driver %q (forgotten import?)", name)
		return
	}
	if err = cmd.Accept(driver); err != nil {
		return
	}
	imports = strings.Join(driver.Imports(), "\n")
	parameters = strings.Join(driver.Parameters(), ",")
	out = strings.Trim(strip.ReplaceAllString(string(driver.Output()), "\n"), "\n\t ")
	return
}

func callSignature(cmd gofire.Command) (call string, ret string, docs string) {
	if cmd.Context {
		call = fmt.Sprintf("%s(ctx, %%s)", cmd.Name)
	} else {
		call = fmt.Sprintf("%s(%%s)", cmd.Name)
	}
	if len(cmd.Returns) != 0 {
		rnames := make([]string, 0, len(cmd.Returns))
		errIndex := -1
		for i, typ := range cmd.Returns {
			rnames = append(rnames, fmt.Sprintf("o%d", i))
			if typ == "error" {
				errIndex = i
			}
		}
		rtypes := cmd.Returns
		if errIndex > 0 {
			rnames[errIndex] = "err"
		}
		// regenerate here after err return name was updated.
		call = fmt.Sprintf("%s = %s", strings.Join(rnames, ","), call)
		if errIndex == 0 {
			rtypes = append(rnames, "err")
			rnames = append(rnames, "err")
		}
		for i := range rnames {
			ret += fmt.Sprintf("%s %s,", rnames[i], rtypes[i])
		}
	}
	if len(cmd.Returns) == 0 {
		ret = "err error"
	}
	doc := strings.Trim(cmd.Doc, "\n\t ")
	comments := strings.Split(doc, "\n")
	for i, comment := range comments {
		comment = strings.Trim(comment, "\t ")
		if !strings.HasPrefix(comment, "//") {
			comments[i] = fmt.Sprintf("// %s", comment)
		}
	}
	docs = strings.Join(comments, "\n")
	return
}
