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
	tp, ok := a.Type.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf("this driver doens't support non primitive type %s", a.Type.Type())
	}
	name := fmt.Sprintf("a%d", a.Index)
	d.params = append(d.params, generators.Parameter{
		Name: name,
		Type: a.Type,
	})
	return d.argument(name, a.Index, tp)
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	tp, ok := f.Type.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf("this driver doens't support non primitive type %s", f.Type.Type())
	}
	var gname string
	if g != nil {
		gname = g.Name
	}
	var defValue *string
	if f.Optional {
		value := fmt.Sprintf("%q", f.Default)
		defValue = &value
	}
	name := fmt.Sprintf("f%s%s", gname, f.Full)
	d.params = append(d.params, generators.Parameter{
		Name: name,
		Type: f.Type,
	})
	return d.flag(name, tp, defValue)
}

func (d *driver) argument(name string, index uint64, t gofire.TPrimitive) error {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		if _, err := fmt.Fprintf(&d.postParse,
			`
				{
					v, err := strconv.ParseBool(flag.Arg(%d))
					if err != nil {
						return err
					}
					%s = (v)
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
					v, err := strconv.ParseInt(flag.Arg(%d), 10, %d)
					if err != nil {
						return err
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
					v, err := strconv.ParseUint(flag.Arg(%d), 10, %d)
					if err != nil {
						return err
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
					v, err := strconv.ParseFloat(flag.Arg(%d), %d)
					if err != nil {
						return err
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
				%s = flag.Arg(%d)
			`,
			name,
			index,
		); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("type %q can't parsed as primitive type", t.Type())
	}
}

func (d *driver) flag(name string, t gofire.TPrimitive, val *string) error {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		v, err := d.parsev(t, val)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				flag.BoolVar(&%s_, %q, %t, "")
			`,
			name,
			name,
			v,
		); err != nil {
			return err
		}
		return nil
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		v, err := d.parsev(t, val)
		if err != nil {
			return err
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ int64
				flag.Int64Var(&%s_, %q, %d, "")
			`,
			name,
			name,
			name,
			v,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stafe.
		// TODO emit error on type overflow.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
			%s = %s(%s_)
		`,
			name,
			t.Type(),
			name,
		); err != nil {
			return err
		}
		return nil
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		v, err := d.parsev(t, val)
		if err != nil {
			return err
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ uint64
				flag.Uint64Var(&%s_, %q, %d, "")
			`,
			name,
			name,
			name,
			v,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stafe.
		// TODO emit error on type overflow.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
			%s = %s(%s_)
		`,
			name,
			t.Type(),
			name,
		); err != nil {
			return err
		}
		return nil
	case gofire.Float32, gofire.Float64:
		v, err := d.parsev(t, val)
		if err != nil {
			return err
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ float64
				flag.Float64Var(&%s_, %q, %f, "")
			`,
			name,
			name,
			name,
			v,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stafe.
		// TODO emit error on type overflow.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
			%s = %s(%s_)
		`,
			name,
			t.Type(),
			name,
		); err != nil {
			return err
		}
		return nil
	case gofire.String:
		v, err := d.parsev(t, val)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				flag.StringVar(&%s, %q, %q, "")
			`,
			name,
			name,
			v,
		); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("type %q can't parsed as primitive type", t.Type())
	}
}

func (d *driver) parsev(t gofire.TPrimitive, val *string) (interface{}, error) {
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
		return nil, fmt.Errorf("type %q can't parsed as primitive type", t.Type())
	}
}
