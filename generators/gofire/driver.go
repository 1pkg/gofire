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
		`"errors"`,
		`"fmt"`,
		`"strconv"`,
		`"strings"`,
		`"os"`,
	}
}

func (d driver) Parameters() []string {
	return d.params
}

func (d driver) Output() []byte {
	return d.Bytes()
}

func (d *driver) Reset() (err error) {
	defer func() {
		if v := recover(); v != nil {
			if verr, ok := v.(error); ok {
				err = verr
			}
		}
	}()
	d.Buffer.Reset()
	d.appendf(`
		largs := len(os.Args)
		tokens := make(map[string]string, len(os.Args))
		var parg string
		var narg int
		for i := 0; i < largs; i++ {
			arg := os.Args[i]
			argb, pargb := strings.HasPrefix(arg, "-"), strings.HasPrefix(parg, "-")
			switch {
				case !argb && !pargb:
					 ["a"+strconv.Itoa(narg)] = arg
					narg++
				case !argb && pargb:
					fln := strings.ReplaceAll(parg, "-", "")
					tokens["f"+fln] = arg
				case argb && pargb:
					fln := strings.ReplaceAll(parg, "-", "")
					tokens["f"+fln] = "true"
					if i == largs-1 {
						fln := strings.ReplaceAll(arg, "-", "")
						tokens["f"+fln] = "true"
					}
				case argb && strings.Contains(arg, "="):
					parts := strings.Split(arg, "-")
					fln := strings.ReplaceAll(parts[0], "-", "")
					tokens["f"+fln] = parts[1]
				case argb && !pargb:
					continue 
				default:
					return fmt.Errorf("cli arguments %%v can't be tokenized near %%s %%s", os.Args, parg, arg)
			}
			parg = arg
		}
	`)
	return
}

func (d *driver) VisitArgument(a gofire.Argument) (err error) {
	return d.visit(
		fmt.Sprintf("a%d", a.Index),
		"",
		a.Type,
		nil,
	)
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	var gname string
	if g != nil {
		gname = g.Name
	}
	var defValue *string
	if f.Optional {
		defValue = &f.Default
	}
	return d.visit(
		fmt.Sprintf("f%s%s", gname, f.Full),
		fmt.Sprintf("f%s%s", gname, f.Short),
		f.Type,
		defValue,
	)
}

func (d *driver) visit(name, altName string, typ gofire.Typ, defValue *string) (err error) {
	defer func() {
		if v := recover(); v != nil {
			if verr, ok := v.(error); ok {
				err = verr
			}
		}
	}()
	d.params = append(d.params, name)
	d.typ(
		name,
		fmt.Sprintf(`"%s"`, name),
		fmt.Sprintf(`"%s"`, altName),
		typ,
		defValue,
	)
	return
}

func (d *driver) typ(name, key, altKey string, t gofire.Typ, defValue *string) *driver {
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
	case gofire.String:
		return d.tprimitive(name, key, altKey, t.(gofire.TPrimitive), defValue)
	case gofire.Array:
		return d.tarray(name, key, altKey, t.(gofire.TArray), defValue)
	case gofire.Slice:
		return d.tslice(name, key, altKey, t.(gofire.TSlice), defValue)
	case gofire.Map:
		return d.tmap(name, key, altKey, t.(gofire.TMap), defValue)
	default:
		panic(fmt.Errorf("unknown or ambiguous type %q can't be parsed", t.Type()))
	}
}

func (d *driver) tarray(name, key, altKey string, t gofire.TArray, defValue *string) *driver {
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
		altKey,
		t.ETyp,
		defValue,
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

func (d *driver) tslice(name, key, altKey string, t gofire.TSlice, defValue *string) *driver {
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
		altKey,
		t.ETyp,
		defValue,
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

func (d *driver) tmap(name, key, altKey string, t gofire.TMap, defValue *string) *driver {
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
		altKey,
		t.KTyp,
		defValue,
	).typ(
		vpname,
		pkey,
		altKey,
		t.VTyp,
		defValue,
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

func (d *driver) tprimitive(name, key, altKey string, t gofire.TPrimitive, defValue *string) *driver {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
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
		).ifAppendf(
			altKey != "",
			`
				else if p, ok := tokens[%s]; ok {
					t%s, err := strconv.ParseBool(p)
					if err != nil {
						return err
					}
					%s = t%s
				}
			`,
			altKey,
			name,
			name,
			name,
		).ifElseAppendf(
			defValue != nil,
			`
				else {
					t%s, err := strconv.ParseBool(%q)
					if err != nil {
						return err
					}
					%s = t%s
				}
			`,
			name,
			sderef(defValue),
			name,
			name,
		)(
			`
				else {
					return errors.New("required cli argument %s hasn't been provided")
				}
			`,
			key,
		)
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
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
		).ifAppendf(
			altKey != "",
			`
				else if p, ok := tokens[%s]; ok {
					t%s, err := strconv.ParseInt(p, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			altKey,
			name,
			k.Base(),
			name,
			k.Type(),
			name,
		).ifElseAppendf(
			defValue != nil,
			`
				else {
					t%s, err := strconv.ParseInt(%q, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			name,
			sderef(defValue),
			k.Base(),
			name,
			k.Type(),
			name,
		)(
			`
				else {
					return errors.New("required cli argument %s hasn't been provided")
				}
			`,
			key,
		)
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
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
		).ifAppendf(
			altKey != "",
			`
				else if p, ok := tokens[%s]; ok {
					t%s, err := strconv.ParseUint(p, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			altKey,
			name,
			k.Base(),
			name,
			k.Type(),
			name,
		).ifElseAppendf(
			defValue != nil,
			`
				else {
					t%s, err := strconv.ParseUint(%q, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			name,
			sderef(defValue),
			k.Base(),
			name,
			k.Type(),
			name,
		)(
			`
				else {
					return errors.New("required cli argument %s hasn't been provided")
				}
			`,
			key,
		)
	case gofire.Float32, gofire.Float64:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
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
		).ifAppendf(
			altKey != "",
			`
				else if p, ok := tokens[%s]; ok {
					t%s, err := strconv.ParseFloat(p, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			altKey,
			name,
			k.Base(),
			name,
			k.Type(),
			name,
		).ifElseAppendf(
			defValue != nil,
			`
				else {
					t%s, err := strconv.ParseFloat(%q, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			name,
			sderef(defValue),
			k.Base(),
			name,
			k.Type(),
			name,
		)(
			`
				else {
					return errors.New("required cli argument %s hasn't been provided")
				}
			`,
			key,
		)
	case gofire.Complex64, gofire.Complex128:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
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
		).ifAppendf(
			altKey != "",
			`
				else if p, ok := tokens[%s]; ok {
					t%s, err := strconv.ParseComplex(p, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			altKey,
			name,
			k.Base(),
			name,
			k.Type(),
			name,
		).ifElseAppendf(
			defValue != nil,
			`
				else {
					t%s, err := strconv.ParseComplex(%q, %d)
					if err != nil {
						return err
					}
					%s = %s(t%s)
				}
			`,
			name,
			sderef(defValue),
			k.Base(),
			name,
			k.Type(),
			name,
		)(
			`
				else {
					return errors.New("required cli argument %s hasn't been provided")
				}
			`,
			key,
		)
	case gofire.String:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
					%s = p
				}
			`,
			name,
			t.Type(),
			key,
			name,
		).ifAppendf(
			altKey != "",
			`
				else if p, ok := tokens[%s]; ok {
					%s = p
				}
			`,
			altKey,
			name,
		).ifElseAppendf(
			defValue != nil,
			`
				else {
					%s = %q
				}
			`,
			name,
			sderef(defValue),
		)(
			`
				else {
					return errors.New("required cli argument %s hasn't been provided")
				}
			`,
			key,
		)
	default:
		panic(fmt.Errorf("type %q can't parsed as primitive type", t.Type()))
	}
}
