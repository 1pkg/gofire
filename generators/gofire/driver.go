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
	if _, err := d.WriteString("args, flags, err := tokenize(os.Args[1:]);"); err != nil {
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
	if a.Ellipsis {
		return fmt.Errorf(
			"driver %s: ellipsis argument types are not supported, got an argument %s %s",
			generators.DriverNameGofire,
			p.Name,
			a.Type.Type(),
		)
	}
	switch p.Type.Kind() {
	case gofire.Bool:
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
	case gofire.Float32, gofire.Float64:
	case gofire.Complex64, gofire.Complex128:
	case gofire.String:
	case gofire.Array:
	case gofire.Slice:
	case gofire.Map:
	default:
		return fmt.Errorf(
			"driver %s: argument type %s is not supported for an argument %s",
			generators.DriverNameGofire,
			p.Type.Type(),
			p.Name,
		)
	}
	if _, err := fmt.Fprintf(d,
		`
			{
				i := %d
				if len(args) <= i {
					return fmt.Errorf("argument %%d-th is required", i)
				}
				v, _, err := parsers.ParseTypeValue(%#v, args[i])
				if err != nil {
					return fmt.Errorf("argument %s value %%v can't be parsed %%v", args[i], err)
				}
				if err := mapstructure.Decode(v, &%s); err != nil {
					return fmt.Errorf("argument %s value %%v can't be decoded %%v", v, err)
				}
			}
		`,
		a.Index,
		p.Type,
		p.Name,
		p.Name,
		p.Name,
	); err != nil {
		return err
	}
	return nil
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	_ = d.Driver.VisitFlag(f, g)
	p := d.Last()
	typ := p.Type
	var amp string
	tprt, ptr := typ.(gofire.TPtr)
	if ptr {
		typ = tprt.ETyp
		amp = "&"
	}
	switch typ.Kind() {
	case gofire.Bool:
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
	case gofire.Float32, gofire.Float64:
	case gofire.Complex64, gofire.Complex128:
	case gofire.String:
	case gofire.Array:
	case gofire.Slice:
	case gofire.Map:
	default:
		return fmt.Errorf(
			"driver %s: flag type %s is not supported for a flag %s",
			generators.DriverNameGofire,
			p.Type.Type(),
			p.Name,
		)
	}
	if _, err := fmt.Fprintf(d,
		`
			{
				f, ok := flags[%q]
				v, set, err := parsers.ParseTypeValue(%#v, f)
				if err != nil {
					return fmt.Errorf("flag %s value %%v can't be parsed %%v", f, err)
				}
				if !ok || !set {
					t := %s
					v = %st
				}
				var t %s
				if err := mapstructure.Decode(v, &t); err != nil {
					return fmt.Errorf("flag %s value %%v can't be decoded %%v", v, err)
				}
				%s = %st
			}
		`,
		p.Full,
		typ,
		p.Name,
		typ.Format(f.Default),
		amp,
		typ.Type(),
		p.Name,
		p.Name,
		amp,
	); err != nil {
		return err
	}
	return nil
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
	s := strings.SplitN(string(b), "// #include", 2)
	if _, err := d.WriteString(s[1]); err != nil {
		return err
	}
	return nil
}
