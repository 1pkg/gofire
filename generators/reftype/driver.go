package reftype

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/generators/internal"
)

func init() {
	generators.Register(generators.DriverNameRefType, internal.Cached(internal.Annotated(new(driver))))
}

type driver struct {
	internal.Driver
	bytes.Buffer
	usageList []string
	printList []string
}

func (d driver) Output(cmd gofire.Command) (string, error) {
	var buf bytes.Buffer
	sort.Strings(d.usageList)
	sort.Strings(d.printList)
	u := strings.Join(d.usageList, " ")
	p := strings.Join(d.printList, " ")
	if _, err := fmt.Fprintf(
		&buf,
		`
			help := func() {
				doc, usage, list := %q, %q, %q
				if doc != "" {
					_, _ = fmt.Fprintln(flag.CommandLine.Output(), doc)
				}
				if usage != "" {
					_, _ = fmt.Fprintln(flag.CommandLine.Output(), usage)
				}
				if list != "" {
					_, _ = fmt.Fprintln(flag.CommandLine.Output(), list)
				}
			}
		`,
		cmd.Doc,
		fmt.Sprintf("%s %s [--help]", cmd.Function, u),
		fmt.Sprintf("%s, %s", cmd.Definition, p),
	); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(
		`
			defer func() {
				if err != nil {
					help()
				}
			}()
		`,
	); err != nil {
		return "", err
	}
	if _, err := buf.WriteString("args, flags, err := reftype.Tokenize(os.Args[1:]);"); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(
		`
			if err != nil {
				return err
			}
			if args == nil {
				args = nil
			}
			if flags == nil {
				flags = nil
			}
			if flags["help"] == "true" {
				return errors.New("help requested")
			}
		`,
	); err != nil {
		return "", err
	}
	if _, err := buf.ReadFrom(&d.Buffer); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (d *driver) Reset() error {
	_ = d.Driver.Reset()
	d.Buffer.Reset()
	d.usageList = nil
	d.printList = nil
	return nil
}

func (driver) Name() generators.DriverName {
	return generators.DriverNameRefType
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
		`"github.com/1pkg/gofire/generators/reftype"`,
		`"github.com/mitchellh/mapstructure"`,
	}
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	_ = d.Driver.VisitArgument(a)
	p := d.Last()
	if a.Ellipsis {
		return fmt.Errorf(
			"driver %s: ellipsis argument types are not supported, got an argument %s %s",
			d.Name(),
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
	case gofire.Slice:
	case gofire.Map:
	default:
		return fmt.Errorf(
			"driver %s: argument type %s is not supported for an argument %s",
			d.Name(),
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
		return fmt.Errorf("driver %s: argument %w", d.Name(), err)
	}
	d.usageList = append(d.usageList, fmt.Sprintf("arg%d", a.Index))
	d.printList = append(d.printList, fmt.Sprintf("arg %d %s", a.Index, p.Type.Type()))
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
	full := p.Full
	if p.Ref != nil {
		full = fmt.Sprintf("%s.%s", p.Ref.Group(), full)
	}
	switch typ.Kind() {
	case gofire.Bool:
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
	case gofire.Float32, gofire.Float64:
	case gofire.Complex64, gofire.Complex128:
	case gofire.String:
	case gofire.Slice:
	case gofire.Map:
	default:
		return fmt.Errorf(
			"driver %s: flag type %s is not supported for a flag %s",
			d.Name(),
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
		full,
		typ,
		p.Name,
		typ.Format(f.Default),
		amp,
		typ.Type(),
		p.Name,
		p.Name,
		amp,
	); err != nil {
		return fmt.Errorf("driver %s: flag %w", d.Name(), err)
	}
	d.usageList = append(d.usageList, fmt.Sprintf("--%s=%s", full, typ.Format(f.Default)))
	d.printList = append(
		d.printList,
		fmt.Sprintf("--%s %s %s (default %s)", full, typ.Type(), p.Doc, typ.Format(f.Default)),
	)
	return nil
}
