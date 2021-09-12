package generators

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/1pkg/gofire"
	"golang.org/x/tools/imports"
)

var (
	driverMu sync.Mutex
	drivers  = make(map[DriverName]Driver)
	stripnl  = regexp.MustCompile(`\n(\s)+\n`)
	strips   = regexp.MustCompile(`[ \t]+`)
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
	proxy, err := proxify(driver, cmd)
	if err != nil {
		return err
	}
	tmpl, err := template.New("gen").Parse(driver.Template())
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, proxy); err != nil {
		return err
	}
	src := []byte(
		strings.Trim(
			strings.TrimSpace(
				stripnl.ReplaceAllString(
					strips.ReplaceAllString(buf.String(), " "),
					"\n",
				),
			),
			"\n",
		),
	)
	if src, err = imports.Process("", src, nil); err != nil {
		return err
	}
	if _, err := w.Write(src); err != nil {
		return err
	}
	return nil
}
