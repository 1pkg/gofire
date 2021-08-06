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
	bytes.Buffer
	params []generators.Parameter
}

func (d driver) Imports() []string {
	return []string{
		`"context"`,
		`"flag"`,
	}
}

func (d driver) Parameters() []generators.Parameter {
	return d.params
}

func (d driver) Output() ([]byte, error) {
	b := bytes.NewBuffer(d.Bytes()[:])
	if _, err := fmt.Fprintf(b, "flag.Parse()\n"); err != nil {
		return nil, err
	}
	for _, p := range d.params {
		if _, err := fmt.Fprintf(b, "%s = %s(%s_)\n", p.Name, p.Type.Type(), p.Name); err != nil {
			return nil, err
		}
	}
	return b.Bytes(), nil
}

func (d *driver) Reset() error {
	d.Buffer.Reset()
	return nil
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	// TODO
	return nil
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
	return d.fprimitive(name, tp, defValue)
}

func (d *driver) fprimitive(name string, t gofire.TPrimitive, defValue *string) error {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		v, err := d.parsev(t, defValue)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(d,
			`
				var %s_ bool
				flag.BoolVar(&%s_, %q, %t, "")
			`,
			name,
			name,
			name,
			v,
		); err != nil {
			return err
		}
		return nil
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		v, err := d.parsev(t, defValue)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(d,
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
		return nil
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		v, err := d.parsev(t, defValue)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(d,
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
		return nil
	case gofire.Float32, gofire.Float64:
		v, err := d.parsev(t, defValue)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(d,
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
		return nil
	case gofire.String:
		v, err := d.parsev(t, defValue)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(d,
			`
				var %s_ string
				flag.StringVar(&%s_, %q, %q, "")
			`,
			name,
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

func (d *driver) parsev(t gofire.TPrimitive, defValue *string) (interface{}, error) {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		if defValue == nil {
			return false, nil
		}
		return strconv.ParseBool(*defValue)
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		if defValue == nil {
			return int64(0), nil
		}
		return strconv.ParseInt(*defValue, 10, int(t.TKind.Base()))
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		if defValue == nil {
			return uint64(0), nil
		}
		return strconv.ParseUint(*defValue, 10, int(t.TKind.Base()))
	case gofire.Float32, gofire.Float64:
		if defValue == nil {
			return float64(0.0), nil
		}
		return strconv.ParseFloat(*defValue, int(t.TKind.Base()))
	case gofire.String:
		if defValue == nil {
			return "", nil
		}
		return *defValue, nil
	default:
		return nil, fmt.Errorf("type %q can't parsed as primitive type", t.Type())
	}
}
