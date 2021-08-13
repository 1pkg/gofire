package flag

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

func init() {
	generators.Register(generators.DriverNameFlag, new(driver))
}

type driver struct {
	preParse  bytes.Buffer
	postParse bytes.Buffer
	params    []generators.Parameter
}

func (d driver) Imports() []string {
	return []string{
		`"flag"`,
		`"fmt"`,
		`"strconv"`,
	}
}

func (d driver) Parameters() []generators.Parameter {
	return d.params
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
	d.preParse.Reset()
	d.postParse.Reset()
	return nil
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	name := fmt.Sprintf("a%d", a.Index)
	tp, ok := a.Type.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf(
			"driver %s does not support non primitive type but got an argument %s %s",
			generators.DriverNameFlag,
			name,
			a.Type.Type(),
		)
	}
	d.params = append(d.params, generators.Parameter{
		Name: name,
		Type: a.Type,
	})
	return d.argument(name, a.Index, tp)
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	var gname string
	if g != nil {
		gname = g.Name
	}
	name := fmt.Sprintf("f%s%s", gname, f.Full)
	tp, ok := f.Type.(gofire.TPtr)
	if !ok {
		return fmt.Errorf(
			"driver %s does not support non pointer to primitive type but got a flag %s %s",
			generators.DriverNameFlag,
			name,
			f.Type.Type(),
		)
	}
	tpRef, ok := tp.ETyp.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf(
			"driver %s does not support non pointer to primitive type but got a flag %s %s",
			generators.DriverNameFlag,
			name,
			f.Type.Type(),
		)
	}
	d.params = append(d.params, generators.Parameter{
		Name: name,
		Type: f.Type,
	})
	var val *string
	if f.Optional {
		value := fmt.Sprintf("%q", f.Default)
		val = &value
	}
	return d.flag(name, tpRef, val, f.Doc)
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
			generators.DriverNameFlag,
			name,
			t.Type(),
		)
	}
}

func (d *driver) flag(name string, t gofire.TPrimitive, val *string, doc string) error {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNameFlag,
				name,
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
		return nil
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNameFlag,
				name,
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
					%s = &v
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
		); err != nil {
			return err
		}
		return nil
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNameFlag,
				name,
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
					%s = &v
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
		); err != nil {
			return err
		}
		return nil
	case gofire.Float32, gofire.Float64:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNameFlag,
				name,
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
					%s = &v
				}
			`,
			t.Type(),
			name,
			name,
		); err != nil {
			return err
		}
		return nil
	case gofire.String:
		v, err := d.parsev(name, t, val)
		if err != nil {
			return fmt.Errorf(
				"driver %s error happened during parsing default value for a flag %s: %v",
				generators.DriverNameFlag,
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
		return nil
	default:
		return fmt.Errorf(
			"driver %s does not support such type for a flag %s %s",
			generators.DriverNameFlag,
			name,
			t.Type(),
		)
	}
}

func (d *driver) parsev(name string, t gofire.TPrimitive, val *string) (interface{}, error) {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		if val == nil {
			return false, nil
		}
		return strconv.ParseBool(*val)
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		if val == nil {
			return int64(0), nil
		}
		return strconv.ParseInt(*val, 10, int(t.TKind.Base()))
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		if val == nil {
			return uint64(0), nil
		}
		return strconv.ParseUint(*val, 10, int(t.TKind.Base()))
	case gofire.Float32, gofire.Float64:
		if val == nil {
			return float64(0.0), nil
		}
		return strconv.ParseFloat(*val, int(t.TKind.Base()))
	case gofire.String:
		if val == nil {
			return "", nil
		}
		return *val, nil
	default:
		return nil, fmt.Errorf(
			"driver %s does not support such type for a flag %s %s",
			generators.DriverNameFlag,
			name,
			t.Type(),
		)
	}
}
