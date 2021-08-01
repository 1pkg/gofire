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
		panic("register called twice for driver " + name)
	}
	drivers[name] = driver
}

// Generate generates cli command using provided driver to provided writer output.
func Generate(ctx context.Context, dn DriverName, c gofire.Command, w io.Writer) error {
	driversMu.RLock()
	driver, ok := drivers[dn]
	driversMu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown driver %q (forgotten import?)", dn)
	}
	if err := c.Accept(driver); err != nil {
		return err
	}
	src := fmt.Sprintf(
		`
			package %s

			import(
				%s
			)
			
			func Command(ctx context.Context) error {
				%s
				return nil
			}
		`,
		c.Pckg,
		strings.Join(driver.Imports(), "\n"),
		strings.Trim(strip.ReplaceAllString(string(driver.Output()), "\n"), "\n\t "),
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
