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
	gen.typ(name, fmt.Sprintf(`"%s"`, name), typ)
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
	vname := fmt.Sprintf("%sv", name)
	iname := fmt.Sprintf("%si", name)
	ikey := fmt.Sprintf(`%s + "_" + strconv.Itoa(%s)`, key, iname)
	return gen.appendf(
		`
			var %s %s
			for %s := 0; %s < %d; %s++ {
		`,
		name,
		t.Type(),
		iname,
		iname,
		t.Size,
		iname,
	).typ(
		vname,
		ikey,
		t.ETyp,
	).appendf(
		`
				%s[%s] = %s
			}
		`,
		name,
		iname,
		vname,
	)
}

func (gen *Fire) tslice(name, key string, t gofire.TSlice) *Fire {
	vname := fmt.Sprintf("%sv", name)
	iname := fmt.Sprintf("%si", name)
	ikey := fmt.Sprintf(`%s + "_" + strconv.Itoa(%s)`, key, iname)
	return gen.appendf(
		`
			var %s %s
			var %s int
			for k := range params {
				if !strings.HasPrefix(k, %s) {
					continue
				}
				%s++
		`,
		name,
		t.Type(),
		iname,
		key,
		iname,
	).typ(
		vname,
		ikey,
		t.ETyp,
	).appendf(
		`
				%s[%s] = %s
			}
		`,
		name,
		iname,
		vname,
	)
}

func (gen *Fire) tmap(name, key string, t gofire.TMap) *Fire {
	vkname := fmt.Sprintf("%sx", name)
	vpname := fmt.Sprintf("%sz", name)
	kname := fmt.Sprintf("%sk", name)
	iname := fmt.Sprintf("%si", name)
	kkey := fmt.Sprintf(`%s + "_k_" + strconv.Itoa(%s)`, key, iname)
	pkey := fmt.Sprintf(`%s + "_v_" + strconv.Itoa(%s)`, key, iname)
	return gen.appendf(
		`
			%s := make(%s)
			var %s int
			for %s := range params {
				if !strings.HasPrefix(%s, %s+"_k") {
					continue
				}
				%s++
		`,
		name,
		t.Type(),
		iname,
		kname,
		kname,
		key,
		iname,
	).typ(
		vkname,
		kkey,
		t.KTyp,
	).typ(
		vpname,
		pkey,
		t.VTyp,
	).appendf(
		`
				%s[%s] = %s
			}
		`,
		name,
		vkname,
		vpname,
	)
}

func (gen *Fire) tprimitive(name, key string, t gofire.TPrimitive) *Fire {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		return gen.appendf(
			`
				var %s %s
				if p, ok := params[%s]; ok {
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
				if p, ok := params[%s]; ok {
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
				if p, ok := params[%s]; ok {
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
				if p, ok := params[%s]; ok {
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
				if p, ok := params[%s]; ok {
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
				if p, ok := params[%s]; ok {
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
