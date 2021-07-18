package generators

import (
	"bytes"
	"fmt"
	"go/format"
	"io"

	"github.com/xyz/playground/internal"
)

type fire struct {
	buffer
}

func NewFire() internal.Generator {
	gen := &fire{}
	gen.rw = &bytes.Buffer{}
	return &fire{}
}

func (gen *fire) Generate(c internal.Command, w io.Writer) error {
	if err := c.Accept(gen); err != nil {
		return nil
	}
	b, err := gen.bytes()
	if err != nil {
		return err
	}
	src := fmt.Sprintf(
		`
			package %s

			import(
				"context"
				"fmt"
				"strconv"
			)
			
			func Command(ctx context.Context) (err error) {
				%s
			}
		`,
		c.Pckg,
		string(b),
	)
	b, err = format.Source([]byte(src))
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func (gen *fire) VisitArgument(a internal.Argument) error {
	name := fmt.Sprintf("a%d", a.Index)
	return gen.typ(name, name, a.Type)
}

func (gen *fire) VisitFlag(f internal.Flag, g *internal.Group) error {
	name := fmt.Sprintf("f%s", f.Short)
	return gen.typ(name, name, f.Type)
}

func (gen *fire) typ(name, key string, t internal.Typ) error {
	k := t.Kind()
	switch k {
	case internal.Bool:
		fallthrough
	case internal.Int, internal.Int8, internal.Int16, internal.Int32, internal.Int64:
		fallthrough
	case internal.Uint, internal.Uint8, internal.Uint16, internal.Uint32, internal.Uint64:
		fallthrough
	case internal.Float32, internal.Float64:
		fallthrough
	case internal.Complex64, internal.Complex128:
		return gen.tprimitive(name, key, t.(internal.TPrimitive))
	case internal.Array:
		return gen.tarray(name, key, t.(internal.TArray))
	case internal.Slice:
		return gen.tslice(name, key, t.(internal.TSlice))
	case internal.Map:
		return gen.tmap(name, key, t.(internal.TMap))
	default:
		return fmt.Errorf("unknown type %q can't parsed", t.Type())
	}
}

func (gen *fire) tarray(name, key string, t internal.TArray) error {
	if err := gen.fprintf(
		`
			var %s %s
			for i := 0; i < %d; i++ {
		`,
		name,
		t.Type(),
		t.Size,
	); err != nil {
		return err
	}
	if err := gen.typ(
		fmt.Sprintf("i%s", name),
		fmt.Sprintf(`fmt.Sprintf("%s_%%d", i)`, name),
		t.ETyp,
	); err != nil {
		return err
	}
	return gen.fprintf(
		`
				%s[i] = i%s
			}
		`,
		name,
		name,
	)
}

func (gen *fire) tslice(name, key string, t internal.TSlice) error {
	if err := gen.fprintf(
		`
			var %s %s
			{
				var i int64
				for key := range params {
					if !strings.HasPrefix(key, %s) {
						continue
					}
					i++
		`,
		name,
		t.Type(),
		key,
	); err != nil {
		return err
	}
	if err := gen.typ(
		fmt.Sprintf("i%s", name),
		fmt.Sprintf(`fmt.Sprintf("%s_%%d", i)`, name),
		t.ETyp,
	); err != nil {
		return err
	}
	return gen.fprintf(
		`
					%s[i] = i%s
				}
			}
		`,
		name,
		name,
	)
}

func (gen *fire) tmap(name, key string, t internal.TMap) error {
	return nil
}

func (gen *fire) tprimitive(name, key string, t internal.TPrimitive) error {
	k := t.Kind()
	switch k {
	case internal.Bool:
		return gen.fprintf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s")]; ok {
					t%s, err := strconv.ParseBool(p)
					if err != nil {
						return err
					}
					%s = t%s
				}
			`,
			name,
			t.Type(),
			key,
			name,
			name,
			name,
		)
	case internal.Int, internal.Int8, internal.Int16, internal.Int32, internal.Int64:
		return gen.fprintf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s")]; ok {
					t%s, err := strconv.ParseInt(p, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			name,
			t.Type(),
			key,
			name,
			k.Base(),
			name,
			k.Type(),
			name,
		)
	case internal.Uint, internal.Uint8, internal.Uint16, internal.Uint32, internal.Uint64:
		return gen.fprintf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s")]; ok {
					t%s, err := strconv.ParseUint(p, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			name,
			t.Type(),
			key,
			name,
			k.Base(),
			name,
			k.Type(),
			name,
		)
	case internal.Float32, internal.Float64:
		return gen.fprintf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s")]; ok {
					t%s, err := strconv.ParseFloat(p, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			name,
			t.Type(),
			key,
			name,
			k.Base(),
			name,
			k.Type(),
			name,
		)
	case internal.Complex64, internal.Complex128:
		return gen.fprintf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s")]; ok {
					t%s, err := strconv.ParseComplex(p, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			name,
			t.Type(),
			key,
			name,
			k.Base(),
			name,
			k.Type(),
			name,
		)
	case internal.String, internal.Interface:
		return gen.fprintf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s")]; ok {
					%s = p
				}
			`,
			name,
			t.Type(),
			key,
			name,
		)
	default:
		return fmt.Errorf("type %q can't parsed as primitive type", t.Type())
	}
}
