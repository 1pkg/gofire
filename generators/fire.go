package generators

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/xyz/playground/internal"
)

type fire struct {
	buffer
	idx uint64
}

func NewFire() internal.Visitor {
	return &fire{
		buffer: buffer{w: &bytes.Buffer{}},
	}
}

func (gen *fire) VisitArgument(a internal.Argument) error {
	name := fmt.Sprintf("a%d", a.Index)
	return gen.passign(name, name, a.Type)
}

func (gen *fire) VisitFlag(f internal.Flag) error {
	name := fmt.Sprintf("f%s", f.Short)
	return gen.passign(name, name, f.Type)
}

func (gen *fire) VisitGroupStart(g internal.Group) error {
	return nil
}

func (gen *fire) VisitGroupEnd(g internal.Group) error {
	return nil
}

func (gen *fire) VisitCommandStart(c internal.Command) error {
	return gen.Append(
		`
			package %s

			import(
				"fmt"
				"strconv"
			)
			
			func main() {
		`,
		c.Pckg,
	)
}

func (gen *fire) VisitCommandEnd(c internal.Command) error {
	if err := gen.Append("}"); err != nil {
		return err
	}
	_, err := os.Stdout.ReadFrom(gen.w)
	return err
}

func (gen *fire) vdef(name string, t internal.Typ) error {
	return gen.Append("var %s %s", name, t.Type())
}

func (gen *fire) vtmp(t internal.Typ) (string, error) {
	gen.idx++
	name := fmt.Sprintf("v%d", gen.idx)
	if err := gen.vdef(name, t); err != nil {
		return "", err
	}
	return name, nil
}

func (gen *fire) passign(name, key string, t internal.Typ) error {
	if err := gen.vdef(name, t); err != nil {
		return err
	}
	k := t.Kind()
	switch k {
	case internal.Bool:
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
	case internal.Int, internal.Int8, internal.Int16, internal.Int32, internal.Int64:
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
	case internal.Uint, internal.Uint8, internal.Uint16, internal.Uint32, internal.Uint64:
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
	case internal.Float32, internal.Float64:
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
	case internal.Complex64, internal.Complex128:
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
	case internal.String, internal.Interface:
		if err := gen.Append("%s = tokens[%s]", name, name); err != nil {
			return err
		}
	case internal.Array:
		tarray := t.(internal.TArray)
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
					%s[i] = %s
				}
			`,
			name,
			tmpv,
		); err != nil {
			return err
		}
	default:
		return errors.New("")
	}
	return nil
}
