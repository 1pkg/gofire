package flag

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/generators/internal"
)

func init() {
	generators.Register(generators.DriverNameFlag, new(driver))
}

type driver struct {
	internal.Driver
	preParse  bytes.Buffer
	postParse bytes.Buffer
	usageList []string
	printList []string
}

func (d driver) Imports() []string {
	return []string{
		`"flag"`,
		`"fmt"`,
		`"strconv"`,
	}
}

func (d driver) Output(cmd gofire.Command) (string, error) {
	var buf bytes.Buffer
	if _, err := buf.WriteString(`
		defer func() {
			if err != nil {
				flag.PrintDefaults()
			}
		}()
	`); err != nil {
		return "", err
	}
	if _, err := buf.Write(d.preParse.Bytes()); err != nil {
		return "", err
	}
	sort.Strings(d.usageList)
	sort.Strings(d.printList)
	u := strings.Join(d.usageList, "\n")
	p := strings.Join(d.printList, "\n")
	if _, err := fmt.Fprintf(&buf, `
		flag.Usage = func() {
			_, _ = fmt.Fprintln(flag.CommandLine.Output(), %q)
			_, _ = fmt.Fprintln(flag.CommandLine.Output(), %q)
			_, _ = fmt.Fprintln(flag.CommandLine.Output(), %q)
		}
	`, cmd.Doc, cmd.Function+" "+u, p); err != nil {
		return "", err
	}
	if _, err := buf.WriteString("flag.Parse()"); err != nil {
		return "", err
	}
	if _, err := buf.Write(d.postParse.Bytes()); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (d *driver) Reset() error {
	_ = d.Driver.Reset()
	d.preParse.Reset()
	d.postParse.Reset()
	d.usageList = nil
	d.printList = nil
	return nil
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	_ = d.Driver.VisitArgument(a)
	p := d.Last()
	tp, ok := a.Type.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf(
			"driver %s: non primitive argument types are not supported, got an argument %s %s",
			generators.DriverNameFlag,
			p.Name,
			a.Type.Type(),
		)
	}
	if err := d.argument(p.Name, a.Index, tp); err != nil {
		return fmt.Errorf("driver %s: argument %w", generators.DriverNameFlag, err)
	}
	return nil
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	_ = d.Driver.VisitFlag(f, g)
	p := d.Last()
	typ := f.Type
	tptr, ptr := typ.(gofire.TPtr)
	if ptr {
		typ = tptr.ETyp
	}
	tprim, ok := typ.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf(
			"driver %s: non pointer and non primitive flag types are not supported, got a flag %s %s",
			generators.DriverNameFlag,
			p.Name,
			f.Type.Type(),
		)
	}
	if err := d.flag(p.Name, tprim, ptr, p.Default, p.Doc); err != nil {
		return fmt.Errorf("driver %s: flag %w", generators.DriverNameFlag, err)
	}
	return nil
}

func (d *driver) argument(name string, index uint64, t gofire.TPrimitive) error {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		if _, err := fmt.Fprintf(&d.postParse,
			`
				{	
					const i = %d
					if flag.NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseBool(flag.Arg(i))
					if err != nil {
						return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
					}
					%s = v
				}
			`,
			index,
			name,
		); err != nil {
			return err
		}
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		if _, err := fmt.Fprintf(&d.postParse,
			`
				{
					const i = %d
					if flag.NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseInt(flag.Arg(i), 10, %d)
					if err != nil {
						return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
					}
					%s = %s(v)
				}
			`,
			index,
			k.Base(),
			name,
			k.Type(),
		); err != nil {
			return err
		}
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		if _, err := fmt.Fprintf(&d.postParse,
			`
				{
					const i = %d
					if flag.NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseUint(flag.Arg(i), 10, %d)
					if err != nil {
						return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
					}
					%s = %s(v)
				}
			`,
			index,
			k.Base(),
			name,
			k.Type(),
		); err != nil {
			return err
		}
	case gofire.Float32, gofire.Float64:
		if _, err := fmt.Fprintf(&d.postParse,
			`
				{
					const i = %d
					if flag.NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseFloat(flag.Arg(i), %d)
					if err != nil {
						return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
					}
					%s = %s(v)
				}
			`,
			index,
			k.Base(),
			name,
			k.Type(),
		); err != nil {
			return err
		}
	case gofire.String:
		if _, err := fmt.Fprintf(&d.postParse,
			`
				{
					const i = %d
					if flag.NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					%s = flag.Arg(i)
				}
			`,
			index,
			name,
		); err != nil {
			return err
		}
	default:
		return fmt.Errorf(
			"type %s is not supported for an argument %s",
			t.Type(),
			name,
		)
	}
	d.usageList = append(d.usageList, fmt.Sprintf("arg%d", index))
	d.printList = append(d.printList, fmt.Sprintf("arg %d %s", index, t.Type()))
	return nil
}

func (d *driver) flag(name string, t gofire.TPrimitive, ptr bool, val string, doc string) error {
	k := t.Kind()
	var amp string
	if ptr {
		amp = "&"
	}
	switch k {
	case gofire.Bool:
		v, err := d.parsev(name, t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ bool
				flag.BoolVar(&%s_, %q, %t, %q)
			`,
			name,
			name,
			name,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				%s = %s%s_
			`,
			name,
			amp,
			name,
		); err != nil {
			return err
		}
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		v, err := d.parsev(name, t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ int64
				flag.Int64Var(&%s_, %q, %d, %q)
			`,
			name,
			name,
			name,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				{
					const min, max = %d, %d
					if %s_ < min || %s_ > max {
						return fmt.Errorf("flag %s overflow error: value %%d is out of the range [%%d,  %%d]", %s_, min, max)
					}
					v := %s(%s_)
					%s =  %sv
				}
			`,
			k.Min(),
			k.Max(),
			name,
			name,
			name,
			name,
			t.Type(),
			name,
			name,
			amp,
		); err != nil {
			return err
		}
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		v, err := d.parsev(name, t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ uint64
				flag.Uint64Var(&%s_, %q, %d, %q)
			`,
			name,
			name,
			name,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				{
					const min, max = %d, %d
					if %s_ < min || %s_ > max {
						return fmt.Errorf("flag %s overflow error: value %%d is out of the range [%%d,  %%d]", %s_, min, max)
					}
					v := %s(%s_)
					%s = %sv
				}
			`,
			k.Min(),
			k.Max(),
			name,
			name,
			name,
			name,
			t.Type(),
			name,
			name,
			amp,
		); err != nil {
			return err
		}
	case gofire.Float32, gofire.Float64:
		v, err := d.parsev(name, t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ float64
				flag.Float64Var(&%s_, %q, %f, %q)
			`,
			name,
			name,
			name,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		// for floats we don't need to emit overflow safeguards.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				{
					v := %s(%s_)
					%s = %sv
				}
			`,
			t.Type(),
			name,
			name,
			amp,
		); err != nil {
			return err
		}
	case gofire.String:
		v, err := d.parsev(name, t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ string
				flag.StringVar(&%s_, %q, %q, %q)
			`,
			name,
			name,
			name,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				%s = %s%s_
			`,
			name,
			amp,
			name,
		); err != nil {
			return err
		}
	default:
		return fmt.Errorf(
			"type %s is not supported for a flag %s",
			t.Type(),
			name,
		)
	}
	d.usageList = append(d.usageList, fmt.Sprintf("-%s", name))
	d.printList = append(d.printList, fmt.Sprintf("-%s %s %s (default %q)", name, t.Type(), doc, val))
	return nil
}

func (d *driver) parsev(name string, t gofire.TPrimitive, val string) (interface{}, error) {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		if val == "" {
			return false, nil
		}
		return strconv.ParseBool(val)
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		if val == "" {
			return int64(0), nil
		}
		return strconv.ParseInt(val, 10, int(k.Base()))
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		if val == "" {
			return uint64(0), nil
		}
		return strconv.ParseUint(val, 10, int(k.Base()))
	case gofire.Float32, gofire.Float64:
		if val == "" {
			return float64(0.0), nil
		}
		return strconv.ParseFloat(val, int(k.Base()))
	case gofire.String:
		return val, nil
	default:
		return nil, nil
	}
}
