package generators

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/xyz/playground/intermediate"
)

type fire struct {
	buffer
	idx uint64
}

func NewFire() intermediate.Visitor {
	return &fire{
		buffer: buffer{w: &bytes.Buffer{}},
	}
}

func (gen *fire) Dump() error {
	_, err := os.Stdout.ReadFrom(gen)
	return err
}

func (gen *fire) VisitArgument(a intermediate.Argument) error {
	name := fmt.Sprintf("a%d", a.Index)
	return gen.passign(name, name, a.Type)
}

func (gen *fire) VisitFlag(f intermediate.Flag) error {
	name := fmt.Sprintf("f%s", f.Short)
	return gen.passign(name, name, f.Type)
}

func (gen *fire) VisitGroup(g intermediate.Group) intermediate.Cursor {
	return nil
}

func (gen *fire) VisitCommand(c intermediate.Command) intermediate.Cursor {
	_ = gen.Append(
		`
			package %s

			import(
				"fmt"
				"strconv"
			)
			
			func main() {}
		`,
		c.Pckg,
	)
	return nil
}

func (gen *fire) vdef(name string, t intermediate.Typ) error {
	return gen.Append("var %s %s", name, t.Type())
}

func (gen *fire) vtmp(t intermediate.Typ) (string, error) {
	gen.idx++
	name := fmt.Sprintf("v%d", gen.idx)
	if err := gen.vdef(name, t); err != nil {
		return "", err
	}
	return name, nil
}

func (gen *fire) passign(name, key string, t intermediate.Typ) error {
	if err := gen.vdef(name, t); err != nil {
		return err
	}
	k := t.Kind()
	switch k {
	case intermediate.Bool:
		tmpv, err := gen.vtmp(t)
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseBool(tokens[%s], 10, %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.Type(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.Int, intermediate.Int8, intermediate.Int16, intermediate.Int32, intermediate.Int64:
		tmpv, err := gen.vtmp(t)
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseInt(tokens[%s], 10, %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.Type(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.Uint, intermediate.Uint8, intermediate.Uint16, intermediate.Uint32, intermediate.Uint64:
		tmpv, err := gen.vtmp(t)
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseUint(tokens[%s], 10, %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.Type(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.Float32, intermediate.Float64:
		tmpv, err := gen.vtmp(t)
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseFloat(tokens[%s], %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.Type(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.Complex64, intermediate.Complex128:
		tmpv, err := gen.vtmp(t)
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseComplex(tokens[%s], %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.Type(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.String, intermediate.Interface:
		if err := gen.Append("%s = tokens[%s]", name, name); err != nil {
			return err
		}
	case intermediate.Array:
		tarray := t.(intermediate.TArray)
		iname := fmt.Sprintf(`fmt.Sprintf("%s%%d", i)`, name)
		tmpv, err := gen.vtmp(t)
		if err != nil {
			return err
		}
		if err := gen.Append(`for i := 0; i < %d; i++ {`, tarray.Size); err != nil {
			return err
		}
		if err := gen.passign(tmpv, iname, tarray.ETyp); err != nil {
			return err
		}
		if err := gen.Append(
			`
					%s = append(%s, %s)
				}
			`,
			name,
			name,
			iname,
		); err != nil {
			return err
		}
	default:
		return errors.New("")
	}
	return nil
}
