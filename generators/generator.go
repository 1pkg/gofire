package generators

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"io"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/1pkg/gofire"
)

var (
	driverMu sync.Mutex
	drivers  = make(map[DriverName]Driver)
	strip    = regexp.MustCompile(`\n(\s)+\n`)
)

// Register makes a generator driver available by the provided name.
// If Register is called twice with the same name or if driver is nil, it panics.
func Register(name DriverName, driver Driver) {
	driverMu.Lock()
	defer driverMu.Unlock()
	if driver == nil {
		panic("register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic(fmt.Errorf("register called twice for driver %q", name))
	}
	drivers[name] = driver
}

// Bind binds actual driver to provided cli command and returns template bag for it to execute on.
var Bind = func(Driver, gofire.Command) (interface{}, error) {
	return nil, nil
}

// Generate generates cli command using provided driver to provided writer output.
func Generate(ctx context.Context, name DriverName, cmd gofire.Command, w io.Writer) error {
	driverMu.Lock()
	driver, ok := drivers[name]
	driverMu.Unlock()
	if !ok {
		return fmt.Errorf("unknown driver %q (forgotten import?)", name)
	}
	if err := driver.Reset(); err != nil {
		return err
	}
	if err := cmd.Accept(driver); err != nil {
		return err
	}
	g, err := Bind(driver, cmd)
	if err != nil {
		return err
	}
	tmpl, err := template.New("gen").Parse(driver.Template())
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g); err != nil {
		return err
	}
	src := []byte(strings.Trim(strip.ReplaceAllString(buf.String(), "\n"), "\n\t "))
	if src, err = format.Source(src); err != nil {
		return err
	}
	if _, err := w.Write(src); err != nil {
		return err
	}
	return nil
}
