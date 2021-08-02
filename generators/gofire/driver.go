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
	d.appendf(
		`
			largs := len(os.Args)
			tokens := make(map[string]string, len(os.Args))
			var parg string
			var narg int
			for i := 0; i < largs; i++ {
				arg := os.Args[i]
				argb, pargb := strings.HasPrefix(arg, "-"), strings.HasPrefix(parg, "-")
				switch {
					case !argb && !pargb:
						tokens["a"+strconv.Itoa(narg)] = arg
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
		`,
	)
	return
}

func (d *driver) VisitArgument(a gofire.Argument) (err error) {
	name := fmt.Sprintf("a%d", a.Index)
	return d.visit(name, "", a.Type, nil)
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
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
	var altName string
	if f.Short != "" {
		altName = fmt.Sprintf("f%s%s", gname, f.Short)
	}
	return d.visit(name, altName, f.Type, defValue)
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
	key := fmt.Sprintf("%q", name)
	// escape alternative key only if it's not empty.
	var altKey string
	if altKey != "" {
		altKey = fmt.Sprintf("%q", altName)
	}
	d.typ(name, key, altKey, typ, defValue)
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
		panic(fmt.Errorf("unknown or ambiguous type %q can't be parsed %s", t.Type(), name))
	}
}

func (d *driver) tarray(name, key, altKey string, t gofire.TArray, defValue *string) *driver {
	return d.appendf(
		`
			var %s %s
			{
				var t string
				if p, ok := tokens[%s]; ok {
					t = p
				} else
		`,
		name,
		t.Type(),
		key,
	).ifAppendf(
		altKey != "",
		`
				if p, ok := tokens[%s]; ok {
					t = p
				} else
		`,
		altKey,
	).ifElseAppendf(
		defValue != nil,
		`
				{
					t = %s
				}
		`,
		sderef(defValue),
	)(
		`
				{
					return errors.New("required cli argument hasn't been provided")
				}
		`,
	).appendf(
		`
				lp := len(t)
				if lp <= 2 || t[0] != "{" || t[l-1] != "}" {
					return errors.New("value can't be parsed as an array")
				}
				t = t[1:lp-1]
				tp := strings.Split(t, ",")
				ltp := len(tp)
				if ltp != %d {
					return errors.New("different array size expected")
				}
				tokens := make(map[string]string, ltp)
				for i := 0; i < ltp; i++ {
					tokens[strconv.Itoa(i)] = tp[i]
		`,
		t.Size,
	).typ(
		"vi",
		"strconv.Itoa(i)",
		"",
		t.ETyp,
		nil,
	).appendf(
		`
					%s[i] = vi
				}
			}
		`,
		name,
	)
}

func (d *driver) tslice(name, key, altKey string, t gofire.TSlice, defValue *string) *driver {
	return d.appendf(
		`
			var %s %s
			{
				var t string
				if p, ok := tokens[%s]; ok {
					t = p
				} else
		`,
		name,
		t.Type(),
		key,
	).ifAppendf(
		altKey != "",
		`
				if p, ok := tokens[%s]; ok {
					t = p
				} else
		`,
		altKey,
	).ifElseAppendf(
		defValue != nil,
		`
				{
					t = %s
				}
		`,
		sderef(defValue),
	)(
		`
				{
					return errors.New("required cli argument hasn't been provided")
				}
		`,
	).appendf(
		`
				lp := len(t)
				if lp <= 2 || t[0] != "{" || t[l-1] != "}" {
					return errors.New("value can't be parsed as a slice")
				}
				t = t[1:lp-1]
				tp := strings.Split(t, ",")
				ltp := len(tp)
				tokens := make(map[string]string, ltp)
				for i := 0; i < ltp; i++ {
					tokens[strconv.Itoa(i)] = tp[i]
		`,
	).typ(
		"vi",
		"strconv.Itoa(i)",
		"",
		t.ETyp,
		nil,
	).appendf(
		`
					%s[i] = vi
				}
			}
		`,
		name,
	)
}

func (d *driver) tmap(name, key, altKey string, t gofire.TMap, defValue *string) *driver {
	return d.appendf(
		`
			var %s %s
			{
				var t string
				if p, ok := tokens[%s]; ok {
					t = p
				} else
		`,
		name,
		t.Type(),
		key,
	).ifAppendf(
		altKey != "",
		`
				if p, ok := tokens[%s]; ok {
					t = p
				} else
		`,
		altKey,
	).ifElseAppendf(
		defValue != nil,
		`
				{
					t = %s
				}
		`,
		sderef(defValue),
	)(
		`
				{
					return errors.New("required cli argument hasn't been provided")
				}
		`,
	).appendf(
		`
				lp := len(t)
				if lp <= 2 || t[0] != "{" || t[l-1] != "}" {
					return errors.New("value can't be parsed as a map")
				}
				t = t[1:lp-1]
				tp := strings.Split(t, ",")
				ltp := len(tp)
				tokens := make(map[string]string, ltp)
				for i := 0; i < ltp; i++ {
					pi := strings.Split(tp[i], ":")
					if len(pi) != 2 {
						return errors.New("value can't be parsed as a map entry")
					}
					tokens["k"+strconv.Itoa(i)] = pi[0]
					tokens["v"+strconv.Itoa(i)] = tp[0]
		`,
	).typ(
		"ki",
		`"k"+strconv.Itoa(i)`,
		"",
		t.KTyp,
		nil,
	).typ(
		"vi",
		`"v"+strconv.Itoa(i)`,
		"",
		t.VTyp,
		nil,
	).appendf(
		`
					%s[ki] = vi
				}
			}
		`,
		name,
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
					t, err := strconv.ParseBool(p)
					if err != nil {
						return err
					}
					%s = t
				} else
			`,
			name,
			t.Type(),
			key,
			name,
		).ifAppendf(
			altKey != "",
			`
				if p, ok := tokens[%s]; ok {
					t, err := strconv.ParseBool(p)
					if err != nil {
						return err
					}
					%s = t
				} else
			`,
			altKey,
			name,
		).ifElseAppendf(
			defValue != nil,
			`
				{
					t, err := strconv.ParseBool(%s)
					if err != nil {
						return err
					}
					%s = t
				}
			`,
			sderef(defValue),
			name,
		)(
			`
				{
					return errors.New("required cli argument hasn't been provided")
				}
			`,
		)
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
					t, err := strconv.ParseInt(p, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				} else
			`,
			name,
			t.Type(),
			key,
			k.Base(),
			name,
			k.Type(),
		).ifAppendf(
			altKey != "",
			`
				if p, ok := tokens[%s]; ok {
					t, err := strconv.ParseInt(p, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				} else
			`,
			altKey,
			k.Base(),
			name,
			k.Type(),
		).ifElseAppendf(
			defValue != nil,
			`
				{
					t, err := strconv.ParseInt(%s, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				}
			`,
			sderef(defValue),
			k.Base(),
			name,
			k.Type(),
		)(
			`
				{
					return errors.New("required cli argument hasn't been provided")
				}
			`,
		)
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
					t, err := strconv.ParseUint(p, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				} else
			`,
			name,
			t.Type(),
			key,
			k.Base(),
			name,
			k.Type(),
		).ifAppendf(
			altKey != "",
			`
				if p, ok := tokens[%s]; ok {
					t, err := strconv.ParseUint(p, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				} else
			`,
			altKey,
			k.Base(),
			name,
			k.Type(),
		).ifElseAppendf(
			defValue != nil,
			`
				{
					t, err := strconv.ParseUint(%s, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				}
			`,
			sderef(defValue),
			k.Base(),
			name,
			k.Type(),
		)(
			`
				{
					return errors.New("required cli argument hasn't been provided")
				}
			`,
		)
	case gofire.Float32, gofire.Float64:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
					t, err := strconv.ParseFloat(p, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				} else
			`,
			name,
			t.Type(),
			key,
			k.Base(),
			name,
			k.Type(),
		).ifAppendf(
			altKey != "",
			`
				if p, ok := tokens[%s]; ok {
					t, err := strconv.ParseFloat(p, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				} else
			`,
			altKey,
			k.Base(),
			name,
			k.Type(),
		).ifElseAppendf(
			defValue != nil,
			`
				{
					t, err := strconv.ParseFloat(%s, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				}
			`,
			sderef(defValue),
			k.Base(),
			name,
			k.Type(),
		)(
			`
				{
					return errors.New("required cli argument hasn't been provided")
				}
			`,
		)
	case gofire.Complex64, gofire.Complex128:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
					t, err := strconv.ParseComplex(p, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				} else
			`,
			t.Type(),
			key,
			name,
			k.Base(),
			name,
			k.Type(),
		).ifAppendf(
			altKey != "",
			`
				if p, ok := tokens[%s]; ok {
					t, err := strconv.ParseComplex(p, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				} else
			`,
			altKey,
			k.Base(),
			name,
			k.Type(),
		).ifElseAppendf(
			defValue != nil,
			`
				{
					t, err := strconv.ParseComplex(%s, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				}
			`,
			sderef(defValue),
			k.Base(),
			name,
			k.Type(),
		)(
			`
				{
					return errors.New("required cli argument hasn't been provided")
				}
			`,
		)
	case gofire.String:
		return d.appendf(
			`
				var %s %s
				if p, ok := tokens[%s]; ok {
					%s = p
				} else
			`,
			name,
			t.Type(),
			key,
			name,
		).ifAppendf(
			altKey != "",
			`
				if p, ok := tokens[%s]; ok {
					%s = p
				} else
			`,
			altKey,
			name,
		).ifElseAppendf(
			defValue != nil,
			`
				{
					%s = %s
				}
			`,
			name,
			sderef(defValue),
		)(
			`
				{
					return errors.New("required cli argument hasn't been provided")
				}
			`,
		)
	default:
		panic(fmt.Errorf("type %q can't parsed as primitive type", t.Type()))
	}
}
