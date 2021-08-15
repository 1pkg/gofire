package pflag

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

// TODO override flag.Usage function with command documentation.

func init() {
	generators.Register(generators.DriverNamePFlag, new(driver))
}

type driver struct {
	generators.Visitor
	preParse  bytes.Buffer
	postParse bytes.Buffer
}

func (d driver) Imports() []string {
	return []string{
		`flag "github.com/spf13/pflag"`,
		`"fmt"`,
		`"strconv"`,
	}
}

func (d driver) Output() ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.Write(d.preParse.Bytes()); err != nil {
		return nil, err
	}
	if _, err := buf.WriteString("flag.Parse()"); err != nil {
		return nil, err
	}
	if _, err := buf.Write(d.postParse.Bytes()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *driver) Reset() error {
	_ = d.Visitor.Reset()
	d.preParse.Reset()
	d.postParse.Reset()
	return nil
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	if err := d.Visitor.VisitArgument(a); err != nil {
		return err
	}
	p := d.Last()
	tp, ok := a.Type.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf(
			"driver %s does not support non primitive argument types but got an argument %s %s",
			generators.DriverNamePFlag,
			p.Name,
			a.Type.Type(),
		)
	}
	return d.argument(p.Name, a.Index, tp)
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	_ = d.Visitor.VisitFlag(f, g)
	p := d.Last()
	tp, ok := f.Type.(gofire.TPtr)
	if !ok {
		return fmt.Errorf(
			"driver %s does not support non pointer flag types but got a flag %s %s",
			generators.DriverNamePFlag,
			p.Name,
			f.Type.Type(),
		)
	}
	return d.flag(p.Name, p.Alt, tp.ETyp, f.Default, f.Doc, f.Deprecated, f.Hidden)
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
		return nil
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
		return nil
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
		return nil
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
		return nil
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
		return nil
	default:
		return fmt.Errorf(
			"driver %s does not support such type for an argument %s %s",
			generators.DriverNamePFlag,
			name,
			t.Type(),
		)
	}
}

func (d *driver) flag(name, short string, t gofire.Typ, val string, doc string, deprecated, hidden bool) error {
	k := t.Kind()
	switch k {
	case gofire.Slice:
		ts := t.(gofire.TSlice)
		tsk := ts.Kind()
		switch tsk {
		// TODO implement slice types here.
		}
	case gofire.Bool:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNamePFlag,
				name,
				err,
			)
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ bool
				flag.BoolVarP(&%s_, %q, %q, %t, %q)
			`,
			name,
			name,
			name,
			short,
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
					v := %s(%s_)
					%s = &v
				}
			`,
			t.Type(),
			name,
			name,
		); err != nil {
			return err
		}
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNamePFlag,
				name,
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				flag.%sVarP(&%s_, %q, %q, %d, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			name,
			short,
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
					v := %s(%s_)
					%s = &v
				}
			`,
			t.Type(),
			name,
			name,
		); err != nil {
			return err
		}
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNamePFlag,
				name,
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				flag.%sVarP(&%s_, %q, %q, %d, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			name,
			short,
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
					v := %s(%s_)
					%s = &v
				}
			`,
			t.Type(),
			name,
			name,
		); err != nil {
			return err
		}
	case gofire.Float32, gofire.Float64:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNamePFlag,
				name,
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				flag.%sVarP(&%s_, %q, %s, %f, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			name,
			short,
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
					v := %s(%s_)
					%s = &v
				}
			`,
			t.Type(),
			name,
			name,
		); err != nil {
			return err
		}
	case gofire.String:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNamePFlag,
				name,
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
				{
					v := %s(%s_)
					%s = &v
				}
			`,
			t.Type(),
			name,
			name,
		); err != nil {
			return err
		}
	default:
		return fmt.Errorf(
			"driver %s does not support such type for a flag %s %s",
			generators.DriverNamePFlag,
			name,
			t.Type(),
		)
	}
	if deprecated {
		if _, err := fmt.Fprintf(&d.preParse,
			`
				flag.MarkDeprecated(%q, "deprecated: %s")
			`,
			name,
			doc,
		); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				flag.MarkShorthandDeprecated(%q, "deprecated: %s")
			`,
			short,
			doc,
		); err != nil {
			return err
		}
	}
	if hidden {
		if _, err := fmt.Fprintf(&d.preParse,
			`
				flag.MarkHidden(%q)
			`,
			name,
		); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) parsev(name string, t gofire.Typ, val string) (interface{}, error) {
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
		return nil, fmt.Errorf(
			"driver %s does not support such type for a flag %s %s",
			generators.DriverNamePFlag,
			name,
			t.Type(),
		)
	}
}
