package generators

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"regexp"
	"strings"

	"github.com/xyz/gofire"
)

type Fire struct {
	bytes.Buffer
}

func (gen *Fire) Generate(c gofire.Command, w io.Writer) error {
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

func (gen *Fire) VisitArgument(a gofire.Argument) (err error) {
	return gen.visit(fmt.Sprintf("a%d", a.Index), a.Type)
}

func (gen *Fire) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	return gen.visit(fmt.Sprintf("f%s", f.Short), f.Type)
}

func (gen *Fire) visit(name string, typ gofire.Typ) (err error) {
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

func (gen *Fire) typ(name, key string, t gofire.Typ) *Fire {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		fallthrough
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		fallthrough
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		fallthrough
	case gofire.Float32, gofire.Float64:
		fallthrough
	case gofire.Complex64, gofire.Complex128:
		fallthrough
	case gofire.String, gofire.Interface:
		return gen.tprimitive(name, key, t.(gofire.TPrimitive))
	case gofire.Array:
		return gen.tarray(name, key, t.(gofire.TArray))
	case gofire.Slice:
		return gen.tslice(name, key, t.(gofire.TSlice))
	case gofire.Map:
		return gen.tmap(name, key, t.(gofire.TMap))
	default:
		panic(fmt.Errorf("unknown type %q can't parsed", t.Type()))
	}
}

func (gen *Fire) tarray(name, key string, t gofire.TArray) *Fire {
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

func (gen *Fire) tslice(name, key string, t gofire.TSlice) *Fire {
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

func (gen *Fire) tmap(name, key string, t gofire.TMap) *Fire {
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

func (gen *Fire) tprimitive(name, key string, t gofire.TPrimitive) *Fire {
	k := t.Kind()
	switch k {
	case gofire.Bool:
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
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
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
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
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
	case gofire.Float32, gofire.Float64:
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
	case gofire.Complex64, gofire.Complex128:
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
	case gofire.String, gofire.Interface:
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
