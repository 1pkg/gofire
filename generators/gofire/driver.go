package gofire

import (
	"bytes"
	"fmt"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

func init() {
	generators.Register(generators.DriverNameGofire, new(driver))
}

type driver struct {
	bytes.Buffer
	params []string
}

func (d driver) Imports() []string {
	return []string{
		`"context"`,
		`"strconv"`,
		`"strings"`,
	}
}

func (d driver) Parameters() []string {
	return d.params
}

func (d driver) Output() []byte {
	return d.Bytes()
}

func (d *driver) VisitArgument(a gofire.Argument) (err error) {
	return d.visit(fmt.Sprintf("a%d", a.Index), a.Type)
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	return d.visit(fmt.Sprintf("f%s", f.Short), f.Type)
}

func (d *driver) visit(name string, typ gofire.Typ) (err error) {
	defer func() {
		if v := recover(); v != nil {
			if verr, ok := v.(error); ok {
				err = verr
			}
		}
	}()
	d.typ(name, fmt.Sprintf(`"%s"`, name), typ)
	return
}

func (d *driver) typ(name, key string, t gofire.Typ) *driver {
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
		return d.tprimitive(name, key, t.(gofire.TPrimitive))
	case gofire.Array:
		return d.tarray(name, key, t.(gofire.TArray))
	case gofire.Slice:
		return d.tslice(name, key, t.(gofire.TSlice))
	case gofire.Map:
		return d.tmap(name, key, t.(gofire.TMap))
	default:
		panic(fmt.Errorf("unknown type %q can't parsed", t.Type()))
	}
}

func (d *driver) tarray(name, key string, t gofire.TArray) *driver {
	vname := fmt.Sprintf("%sv", name)
	iname := fmt.Sprintf("%si", name)
	ikey := fmt.Sprintf(`%s + "_" + strconv.Itoa(%s)`, key, iname)
	return d.appendf(
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

func (d *driver) tslice(name, key string, t gofire.TSlice) *driver {
	vname := fmt.Sprintf("%sv", name)
	iname := fmt.Sprintf("%si", name)
	ikey := fmt.Sprintf(`%s + "_" + strconv.Itoa(%s)`, key, iname)
	return d.appendf(
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

func (d *driver) tmap(name, key string, t gofire.TMap) *driver {
	vkname := fmt.Sprintf("%sx", name)
	vpname := fmt.Sprintf("%sz", name)
	kname := fmt.Sprintf("%sk", name)
	iname := fmt.Sprintf("%si", name)
	kkey := fmt.Sprintf(`%s + "_k_" + strconv.Itoa(%s)`, key, iname)
	pkey := fmt.Sprintf(`%s + "_v_" + strconv.Itoa(%s)`, key, iname)
	return d.appendf(
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

func (d *driver) tprimitive(name, key string, t gofire.TPrimitive) *driver {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		return d.appendf(
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
		return d.appendf(
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
		return d.appendf(
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
		return d.appendf(
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
		return d.appendf(
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
		return d.appendf(
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

func (d *driver) appendf(format string, a ...interface{}) *driver {
	if _, err := fmt.Fprintf(d, format, a...); err != nil {
		panic(err)
	}
	return d
}
