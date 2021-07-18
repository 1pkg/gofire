package generators

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
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
				"context"
				"fmt"
				"strconv"
			)
			
			func Command(ctx context.Context) (err error) {
				paramenters := make(map[string]string)
		`,
		c.Pckg,
	)
}

func (gen *fire) VisitCommandEnd(c internal.Command) error {
	if err := gen.Append("}"); err != nil {
		return err
	}
	b, err := gen.Bytes()
	if err != nil {
		return err
	}
	b, err = format.Source(b)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(b); err != nil {
		return err
	}
	return nil
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
		v, err := gen.vtmp(internal.TPrimitive{TKind: internal.Bool})
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseBool(parameters[%s], 10, %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			v,
			key,
			k.Base(),
			name,
			k.Type(),
			v,
		); err != nil {
			return err
		}
	case internal.Int, internal.Int8, internal.Int16, internal.Int32, internal.Int64:
		v, err := gen.vtmp(internal.TPrimitive{TKind: internal.Int64})
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseInt(parameters[%s], 10, %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			v,
			key,
			k.Base(),
			name,
			k.Type(),
			v,
		); err != nil {
			return err
		}
	case internal.Uint, internal.Uint8, internal.Uint16, internal.Uint32, internal.Uint64:
		v, err := gen.vtmp(internal.TPrimitive{TKind: internal.Uint64})
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseUint(parameters[%s], 10, %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			v,
			key,
			k.Base(),
			name,
			k.Type(),
			v,
		); err != nil {
			return err
		}
	case internal.Float32, internal.Float64:
		v, err := gen.vtmp(internal.TPrimitive{TKind: internal.Float64})
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseFloat(parameters[%s], %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			v,
			key,
			k.Base(),
			name,
			k.Type(),
			v,
		); err != nil {
			return err
		}
	case internal.Complex64, internal.Complex128:
		v, err := gen.vtmp(internal.TPrimitive{TKind: internal.Complex128})
		if err != nil {
			return err
		}
		if err := gen.Append(
			`
				%s, err = strconv.ParseComplex(parameters[%s], %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			v,
			key,
			k.Base(),
			name,
			k.Type(),
			v,
		); err != nil {
			return err
		}
	case internal.String, internal.Interface:
		if err := gen.Append(`%s = parameters[%s]`, name, key); err != nil {
			return err
		}
	case internal.Array:
		tarray := t.(internal.TArray)
		iname := fmt.Sprintf(`fmt.Sprintf("%s%%d", i)`, name)
		v, err := gen.vtmp(tarray.ETyp)
		if err != nil {
			return err
		}
		if err := gen.Append(`for i := 0; i < %d; i++ {`, tarray.Size); err != nil {
			return err
		}
		if err := gen.passign(v, iname, tarray.ETyp); err != nil {
			return err
		}
		if err := gen.Append(
			`
					%s[i] = %s
				}
			`,
			name,
			v,
		); err != nil {
			return err
		}
	default:
		return errors.New("")
	}
	return nil
}
