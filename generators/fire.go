package generators

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"regexp"
	"strings"

	"github.com/xyz/playground/internal"
)

type Fire struct {
	bytes.Buffer
}

func (gen *Fire) Generate(c internal.Command, w io.Writer) error {
	if err := c.Accept(gen); err != nil {
		return err
	}
	strip := regexp.MustCompile(`\n(\s)+\n`)
	src := fmt.Sprintf(
		`
			package %s

			import(
				"context"
				"fmt"
				"strconv"
				"strings"
			)
			
			var params map[string]string
			func Command(ctx context.Context) error {
				%s
				return nil
			}
		`,
		c.Pckg,
		strings.Trim(strip.ReplaceAllString(gen.String(), "\n"), "\n\t "),
	)
	b, err := format.Source([]byte(src))
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func (gen *Fire) VisitArgument(a internal.Argument) (err error) {
	return gen.visit(fmt.Sprintf("a%d", a.Index), a.Type)
}

func (gen *Fire) VisitFlag(f internal.Flag, g *internal.Group) error {
	return gen.visit(fmt.Sprintf("f%s", f.Short), f.Type)
}

func (gen *Fire) visit(name string, typ internal.Typ) (err error) {
	defer func() {
		if v := recover(); v != nil {
			if verr, ok := v.(error); ok {
				err = verr
			}
		}
	}()
	gen.typ(name, name, typ)
	return
}

func (gen *Fire) typ(name, key string, t internal.Typ) *Fire {
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
		fallthrough
	case internal.String, internal.Interface:
		return gen.tprimitive(name, key, t.(internal.TPrimitive))
	case internal.Array:
		return gen.tarray(name, key, t.(internal.TArray))
	case internal.Slice:
		return gen.tslice(name, key, t.(internal.TSlice))
	case internal.Map:
		return gen.tmap(name, key, t.(internal.TMap))
	default:
		panic(fmt.Errorf("unknown type %q can't parsed", t.Type()))
	}
}

func (gen *Fire) tarray(name, key string, t internal.TArray) *Fire {
	return gen.appendf(
		`
			var %s %s
			for i := 0; i < %d; i++ {
		`,
		name,
		t.Type(),
		t.Size,
	).typ(
		fmt.Sprintf("i%s", name),
		fmt.Sprintf(`fmt.Sprintf("%s_%%d", i)`, name),
		t.ETyp,
	).appendf(
		`
				%s[i] = i%s
			}
		`,
		name,
		name,
	)
}

func (gen *Fire) tslice(name, key string, t internal.TSlice) *Fire {
	return gen.appendf(
		`
			var %s %s
			{
				var i int64
				for key := range params {
					if !strings.HasPrefix(key, %q) {
						continue
					}
					i++
		`,
		name,
		t.Type(),
		key,
	).typ(
		fmt.Sprintf("i%s", name),
		fmt.Sprintf(`fmt.Sprintf("%s_%%d", i)`, name),
		t.ETyp,
	).appendf(
		`
					%s[i] = i%s
				}
			}
		`,
		name,
		name,
	)
}

func (gen *Fire) tmap(name, key string, t internal.TMap) *Fire {
	return gen.appendf(
		`
			%s := make(%s)
			for key, val := range params {
				if !strings.HasPrefix(key, %q) {
					continue
				}
		`,
		name,
		t.Type(),
		key,
	).typ(
		fmt.Sprintf("k%s", name),
		fmt.Sprintf(`fmt.Sprintf("%s_%%v", key)`, name),
		t.KTyp,
	).typ(
		fmt.Sprintf("v%s", name),
		fmt.Sprintf(`fmt.Sprintf("%s_%%v", key)`, name),
		t.KTyp,
	).appendf(
		`
					%s[k%s] = v%s
				}
			}
		`,
		name,
		name,
		name,
	)
}

func (gen *Fire) tprimitive(name, key string, t internal.TPrimitive) *Fire {
	k := t.Kind()
	switch k {
	case internal.Bool:
		return gen.appendf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s)]; ok {
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
		return gen.appendf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s)]; ok {
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
		return gen.appendf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s)]; ok {
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
		return gen.appendf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s)]; ok {
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
		return gen.appendf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s)]; ok {
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
		return gen.appendf(
			`
				var %s %s
				if p, ok := params[fmt.Sprintf("%%q", %s)]; ok {
					%s = p
				}
			`,
			name,
			t.Type(),
			key,
			name,
		)
	default:
		panic(fmt.Errorf("type %q can't parsed as primitive type", t.Type()))
	}
}

func (gen *Fire) appendf(format string, a ...interface{}) *Fire {
	if _, err := fmt.Fprintf(gen, format, a...); err != nil {
		panic(err)
	}
	return gen
}
