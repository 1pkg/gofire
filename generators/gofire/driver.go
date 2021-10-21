package gofire

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/generators/internal"
)

func init() {
	generators.Register(generators.DriverNameGofire, internal.Cached(internal.Annotated(new(driver))))
}

type driver struct {
	internal.Driver
	bytes.Buffer
}

func (d driver) Output(gofire.Command) (string, error) {
	return d.String(), nil
}

func (d *driver) Reset() error {
	_ = d.Driver.Reset()
	d.Buffer.Reset()
	if err := d.include("tokenize.go"); err != nil {
		return err
	}
	if _, err := d.WriteString("args, flags, err := tokenize(os.Args)"); err != nil {
		return err
	}
	if err := d.include("parsetv.go"); err != nil {
		return err
	}
	return nil
}

func (d driver) Imports() []string {
	return []string{
		`"errors"`,
		`"fmt"`,
		`"strconv"`,
		`"strings"`,
		`"unicode"`,
		`"os"`,
		`"github.com/1pkg/gofire"`,
		`"github.com/1pkg/gofire/parsers"`,
		`"github.com/mitchellh/mapstructure"`,
	}
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	_ = d.Driver.VisitArgument(a)
	p := d.Last()
	return d.visit(p.Name, p.Name, p.Type, false, "", fmt.Sprintf("args[%d]", a.Index))
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	_ = d.Driver.VisitFlag(f, g)
	p := d.Last()
	typ := p.Type
	tprt, ptr := typ.(gofire.TPtr)
	if ptr {
		typ = tprt.ETyp
	}
	return d.visit(p.Name, p.Full, typ, ptr, f.Default, fmt.Sprintf("flags[%s]", f.Full))
}

func (d *driver) visit(name, param string, t gofire.Typ, ptr bool, val interface{}, assing string) error {
	var amp string
	if ptr {
		amp = "&"
	}
	k := t.Kind()
	switch k {
	case gofire.Bool:
		fallthrough
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		fallthrough
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		fallthrough
	case gofire.Float32, gofire.Float64:
		fallthrough
	case gofire.Complex64, gofire.Complex128:
		fallthrough
	case gofire.String:
		fallthrough
	case gofire.Array:
		fallthrough
	case gofire.Slice:
		fallthrough
	case gofire.Map:
		if _, err := fmt.Fprintf(d,
			`
				{
					p := %s
					v, err := parsetv(%#v, p, %q)
					if err != nil {
						return fmt.Errorf("parameter %s value %%s can't be parsed %%v", p, err)
					}
					var t %s
					if err := mapstructure.Decode(v, &t); err != nil {
						return fmt.Errorf("parameter %s value %%s can't be decoded %%v", p, err)
					}
					%s = %st
				}
			`,
			assing,
			t,
			val,
			param,
			t.Type(),
			param,
			name,
			amp,
		); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf(
			"type %s is not supported for parameter %s",
			t.Type(),
			param,
		)
	}
}

func (d *driver) include(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	s := strings.SplitN(string(b), "// ---", 2)
	if _, err := d.WriteString(s[1]); err != nil {
		return err
	}
	return nil
}
