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
func Generate(ctx context.Context, dn DriverName, cmd gofire.Command, w io.Writer) error {
	driversMu.RLock()
	driver, ok := drivers[dn]
	driversMu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown driver %q (forgotten import?)", dn)
	}
	if err := cmd.Accept(driver); err != nil {
		return err
	}
	funcCall, funcRet := funcCallSinnature(cmd, driver.Parameters())
	src := fmt.Sprintf(
		`
			package %s

			import(
				%s
			)
			
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
		strings.Join(driver.Imports(), "\n"),
		cmd.Func.Name,
		funcRet,
		strings.Trim(strip.ReplaceAllString(string(driver.Output()), "\n"), "\n\t "),
		funcCall,
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

func funcCallSinnature(cmd gofire.Command, params []string) (call string, ret string) {
	if cmd.Func.Context {
		call = fmt.Sprintf("%s(ctx, %%s)", cmd.Func.Name)
	} else {
		call = fmt.Sprintf("%s(%%s)", cmd.Func.Name)
	}
	if len(cmd.Func.Returns) != 0 {
		rnames := make([]string, 0, len(cmd.Func.Returns))
		errIndex := -1
		for i, typ := range cmd.Func.Returns {
			rnames = append(rnames, fmt.Sprintf("o%d", i))
			if typ == "error" {
				errIndex = i
			}
		}
		rtypes := cmd.Func.Returns
		if errIndex > 0 {
			rnames[errIndex] = "err"
			call = fmt.Sprintf("%s = %s", strings.Join(rnames, ","), call)
		} else {
			call = fmt.Sprintf("%s = %s", strings.Join(rnames, ","), call)
			rtypes = append(rnames, "err")
			rnames = append(rnames, "err")
		}
		for i := range rnames {
			ret += fmt.Sprintf("%s %s,", rnames[i], rtypes[i])
		}
	} else {
		ret = "err error"
	}
	call = fmt.Sprintf(call, strings.Join(params, ","))
	return
}
