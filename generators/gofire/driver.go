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
	params []generators.Parameter
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

func (d driver) Parameters() []generators.Parameter {
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
	// reset the buffer and append cli os.Args parse code.
	d.Buffer.Reset()
	d.appendf(
		`
			_x0tokens := make(map[string]string, len(os.Args))
			{
				largs := len(os.Args)
				var parg string
				var narg int
				for i := 0; i < largs; i++ {
					arg := os.Args[i]
					argb, pargb := strings.HasPrefix(arg, "-"), strings.HasPrefix(parg, "-")
					switch {
						case !argb && !pargb:
							_x0tokens["a"+strconv.Itoa(narg)] = arg
							narg++
						case !argb && pargb:
							fln := strings.ReplaceAll(parg, "-", "")
							_x0tokens["f"+fln] = arg
						case argb && pargb:
							fln := strings.ReplaceAll(parg, "-", "")
							_x0tokens["f"+fln] = "true"
							if i == largs-1 {
								fln := strings.ReplaceAll(arg, "-", "")
								_x0tokens["f"+fln] = "true"
							}
						case argb && strings.Contains(arg, "="):
							parts := strings.Split(arg, "=")
							fln := strings.ReplaceAll(parts[0], "-", "")
							_x0tokens["f"+fln] = parts[1]
						case argb && !pargb:
							continue 
						default:
							return fmt.Errorf("cli arguments %%v can't be tokenized near %%s %%s", os.Args, parg, arg)
					}
					parg = arg
				}
			}
			_x0alts := make(map[string]string, len(os.Args))
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
	d.params = append(d.params, generators.Parameter{
		Name: name,
		Type: typ,
	})
	key := fmt.Sprintf("%q", name)
	// in case alternative name is present add it to alt name lookups.
	if altName != "" {
		altKey := fmt.Sprintf("%q", altName)
		d.appendf(
			`
			_x0alts[%s] = %s
			`,
			key,
			altKey,
		)
	}
	d.typ(name, key, typ, defValue)
	return
}

func (d *driver) typ(name, key string, t gofire.Typ, defValue *string) *driver {
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
		return d.tprimitive(name, key, t.(gofire.TPrimitive), defValue)
	case gofire.Array:
		return d.tarray(name, key, t.(gofire.TArray), defValue)
	case gofire.Slice:
		return d.tslice(name, key, t.(gofire.TSlice), defValue)
	case gofire.Map:
		return d.tmap(name, key, t.(gofire.TMap), defValue)
	default:
		panic(fmt.Errorf("unknown or ambiguous type %q can't be parsed %s", t.Type(), name))
	}
}

func (d *driver) tarray(name, key string, t gofire.TArray, defValue *string) *driver {
	return d.tdefinition(name, t).appendf("{").tdefinition("t", gofire.TPrimitive{TKind: gofire.String}).tassignment(
		key,
		defValue,
		"t = %s",
	).appendf(
		`
			lp := len(t)
			if lp <= 2 || t[0] != '{' || t[lp-1] != '}' {
				return errors.New("value can't be parsed as an array")
			}
			t = t[1:lp-1]
			tp := strings.Split(t, ",")
			ltp := len(tp)
			if ltp != %d {
				return errors.New("different array size expected")
			}
			_x0tokens := make(map[string]string, ltp)
			for i := 0; i < ltp; i++ {
				_x0tokens[strconv.Itoa(i)] = tp[i]
		`,
		t.Size,
	).typ(
		name+"vi",
		"strconv.Itoa(i)",
		t.ETyp,
		nil,
	).appendf(
		`
				%s[i] = %s
			}
		`,
		name,
		name+"vi",
	).appendf("}")
}

func (d *driver) tslice(name, key string, t gofire.TSlice, defValue *string) *driver {
	return d.tdefinition(name, t).appendf("{").tdefinition("t", gofire.TPrimitive{TKind: gofire.String}).tassignment(
		key,
		defValue,
		"t = %s",
	).appendf(
		`
			lp := len(t)
			if lp <= 2 || t[0] != '{' || t[lp-1] != '}' {
				return errors.New("value can't be parsed as a slice")
			}
			t = t[1:lp-1]
			tp := strings.Split(t, ",")
			ltp := len(tp)
			_x0tokens := make(map[string]string, ltp)
			for i := 0; i < ltp; i++ {
				_x0tokens[strconv.Itoa(i)] = tp[i]
		`,
	).typ(
		name+"vi",
		"strconv.Itoa(i)",
		t.ETyp,
		nil,
	).appendf(
		`
				%s = append(%s, %s)
			}
		`,
		name,
		name,
		name+"vi",
	).appendf("}")
}

func (d *driver) tmap(name, key string, t gofire.TMap, defValue *string) *driver {
	return d.tdefinition(name, t).appendf("{").tdefinition("t", gofire.TPrimitive{TKind: gofire.String}).tassignment(
		key,
		defValue,
		"t = %s",
	).appendf(
		`
			lp := len(t)
			if lp <= 2 || t[0] != '{' || t[lp-1] != '}' {
				return errors.New("value can't be parsed as a map")
			}
			t = t[1:lp-1]
			tp := strings.Split(t, ",")
			ltp := len(tp)
			_x0tokens := make(map[string]string, ltp)
			for i := 0; i < ltp; i++ {
				pi := strings.Split(tp[i], ":")
				if len(pi) != 2 {
					return errors.New("value can't be parsed as a map entry")
				}
				_x0tokens["k"+strconv.Itoa(i)] = pi[0]
				_x0tokens["v"+strconv.Itoa(i)] = pi[1]
		`,
	).typ(
		name+"ki",
		`"k"+strconv.Itoa(i)`,
		t.KTyp,
		nil,
	).typ(
		name+"vi",
		`"v"+strconv.Itoa(i)`,
		t.VTyp,
		nil,
	).appendf(
		`
				%s[%s] = %s
			}
		`,
		name,
		name+"ki",
		name+"vi",
	).appendf("}")
}

func (d *driver) tprimitive(name, key string, t gofire.TPrimitive, defValue *string) *driver {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		return d.tdefinition(name, t).tassignment(
			key,
			defValue,
			fmt.Sprintf(
				`
					t, err := strconv.ParseBool(%%s)
					if err != nil {
						return err
					}
					%s = (t)
				`,
				name,
			),
		)
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		return d.tdefinition(name, t).tassignment(
			key,
			defValue,
			fmt.Sprintf(
				`
					t, err := strconv.ParseInt(%%s, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				`,
				k.Base(),
				name,
				k.Type(),
			),
		)
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		return d.tdefinition(name, t).tassignment(
			key,
			defValue,
			fmt.Sprintf(
				`
					t, err := strconv.ParseUint(%%s, 10, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				`,
				k.Base(),
				name,
				k.Type(),
			),
		)
	case gofire.Float32, gofire.Float64:
		return d.tdefinition(name, t).tassignment(
			key,
			defValue,
			fmt.Sprintf(
				`
					t, err := strconv.ParseFloat(%%s, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				`,
				k.Base(),
				name,
				k.Type(),
			),
		)
	case gofire.Complex64, gofire.Complex128:
		return d.tdefinition(name, t).tassignment(
			key,
			defValue,
			fmt.Sprintf(
				`
					t, err := strconv.ParseComplex(%%s, %d)
					if err != nil {
						return err
					}
					%s = %s(t)
				`,
				k.Base(),
				name,
				k.Type(),
			),
		)
	case gofire.String:
		return d.tdefinition(name, t).tassignment(
			key,
			defValue,
			fmt.Sprintf("%s = %%s", name),
		)
	default:
		panic(fmt.Errorf("type %q can't parsed as primitive type", t.Type()))
	}
}

func (d *driver) tdefinition(name string, typ gofire.Typ) *driver {
	// in case it's top level parameter do not add definitions.
	for _, p := range d.params {
		if p.Name == name && p.Type == typ {
			return d
		}
	}
	return d.appendf(
		`
			var %s %s
		`,
		name,
		typ.Type(),
	)
}

func (d *driver) tassignment(key string, defValue *string, assignmentStmt string) *driver {
	return d.appendf(
		`
			if p, ok := _x0tokens[%s]; ok {
				%s
			} else
		`,
		key,
		fmt.Sprintf(assignmentStmt, "p"),
	).appendf(
		`
			if p, ok := _x0tokens[_x0alts[%s]]; ok {
				%s
			} else
		`,
		key,
		fmt.Sprintf(assignmentStmt, "p"),
	).ifElseAppendf(
		defValue != nil,
		func(f fprintf) {
			f(
				`
					{
						%s
					}
				`,
				fmt.Sprintf(assignmentStmt, *defValue),
			)
		},
		func(f fprintf) {
			f(
				`
					{
						return errors.New("required cli argument hasn't been provided")
					}
				`,
			)
		},
	)
}
