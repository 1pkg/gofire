package gofire

import (
	"bytes"
	"fmt"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/generators/internal"
)

func init() {
	generators.Register(generators.DriverNameGofire, internal.Cached(internal.Annotated(new(driver))))
}

type driver struct {
	internal.Driver
	bytes.Buffer
	alternatives map[string]string
}

func (d driver) Output(gofire.Command) (string, error) {
	return d.String(), nil
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
	_ = d.Driver.Reset()
	d.Buffer.Reset()
	d.alternatives = make(map[string]string)
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
		`,
	)
	return
}

func (d driver) Imports() []string {
	return []string{
		`"errors"`,
		`"fmt"`,
		`"strconv"`,
		`"strings"`,
		`"os"`,
	}
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	_ = d.Driver.VisitArgument(a)
	return d.visit(*d.Last(), nil)
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	_ = d.Driver.VisitFlag(f, g)
	v := fmt.Sprintf("%q", f.Default)
	return d.visit(*d.Last(), &v)
}

func (d *driver) visit(p generators.Parameter, val *string) (err error) {
	defer func() {
		if v := recover(); v != nil {
			if verr, ok := v.(error); ok {
				err = verr
			}
		}
	}()
	key := fmt.Sprintf("%q", p.Name)
	// in case alternative name is present add it to alt name lookups.
	if p.Alt != "" {
		altKey := fmt.Sprintf("%q", p.Alt)
		d.alternatives[key] = altKey
	}
	d.typ(p.Name, key, p.Type, val)
	return
}

func (d *driver) typ(name, key string, t gofire.Typ, val *string) *driver {
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
		return d.tprimitive(name, key, t.(gofire.TPrimitive), val)
	case gofire.Array:
		return d.tarray(name, key, t.(gofire.TArray), val)
	case gofire.Slice:
		return d.tslice(name, key, t.(gofire.TSlice), val)
	case gofire.Map:
		return d.tmap(name, key, t.(gofire.TMap), val)
	case gofire.Ptr:
		return nil // TODO
	default:
		panic(fmt.Errorf("unknown or ambiguous type %q can't be parsed %s", t.Type(), name))
	}
}

func (d *driver) tarray(name, key string, t gofire.TArray, val *string) *driver {
	return d.tdefinition(name, t).appendf("{").tdefinition("t", gofire.TPrimitive{TKind: gofire.String}).tassignment(
		key,
		val,
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

func (d *driver) tslice(name, key string, t gofire.TSlice, val *string) *driver {
	return d.tdefinition(name, t).appendf("{").tdefinition("t", gofire.TPrimitive{TKind: gofire.String}).tassignment(
		key,
		val,
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

func (d *driver) tmap(name, key string, t gofire.TMap, val *string) *driver {
	return d.tdefinition(name, t).appendf("{").tdefinition("t", gofire.TPrimitive{TKind: gofire.String}).tassignment(
		key,
		val,
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

func (d *driver) tprimitive(name, key string, t gofire.TPrimitive, val *string) *driver {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		return d.tdefinition(name, t).tassignment(
			key,
			val,
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
			val,
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
			val,
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
			val,
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
			val,
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
			val,
			fmt.Sprintf("%s = %%s", name),
		)
	default:
		panic(fmt.Errorf("type %q can't parsed as primitive type", t.Type()))
	}
}

func (d *driver) tdefinition(name string, typ gofire.Typ) *driver {
	// in case it's top level parameter do not add definitions.
	for _, p := range d.Parameters() {
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

func (d *driver) tassignment(key string, val *string, assignmentStmt string) *driver {
	return d.appendf(
		`
			if p, ok := _x0tokens[%s]; ok {
				%s
			} else
		`,
		key,
		fmt.Sprintf(assignmentStmt, "p"),
	).ifElseAppendf(
		d.alternatives[key] != "",
		func(f fprintf) {
			f(
				`
					if p, ok := _x0tokens[%s]; ok {
						%s
					} else
				`,
				d.alternatives[key],
				fmt.Sprintf(assignmentStmt, "p"),
			)
		},
		func(f fprintf) {},
	).ifElseAppendf(
		val != nil,
		func(f fprintf) {
			f(
				`
					{
						%s
					}
				`,
				fmt.Sprintf(assignmentStmt, *val),
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
