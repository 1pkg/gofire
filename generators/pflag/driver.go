package pflag

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/generators/internal"
)

func init() {
	generators.Register(generators.DriverNamePFlag, internal.Cached(internal.Annotated(new(driver))))
}

type driver struct {
	internal.Driver
	preParse   bytes.Buffer
	postParse  bytes.Buffer
	usageList  []string
	printList  []string
	shortNames map[string]bool
}

func (d driver) Output(cmd gofire.Command) (string, error) {
	var buf bytes.Buffer
	if _, err := buf.WriteString(
		`
			defer func() {
				if err != nil {
					pflag.Usage()
				}
			}()
		`,
	); err != nil {
		return "", err
	}
	if _, err := buf.Write(d.preParse.Bytes()); err != nil {
		return "", err
	}
	sort.Strings(d.usageList)
	sort.Strings(d.printList)
	u := strings.Join(d.usageList, " ")
	p := strings.Join(d.printList, " ")
	if _, err := fmt.Fprintf(
		&buf,
		`
			pflag.Usage = func() {
				doc, usage, list := %q, %q, %q
				if doc != "" {
					_, _ = fmt.Fprintln(flag.CommandLine.Output(), doc)
				}
				if usage != "" {
					_, _ = fmt.Fprintln(flag.CommandLine.Output(), usage)
				}
				if list != "" {
					_, _ = fmt.Fprintln(flag.CommandLine.Output(), list)
				}
			}
		`,
		cmd.Doc,
		cmd.Function+" "+u,
		p,
	); err != nil {
		return "", err
	}
	if _, err := buf.WriteString("pflag.Parse()"); err != nil {
		return "", err
	}
	if _, err := buf.Write(d.postParse.Bytes()); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (d *driver) Reset() error {
	_ = d.Driver.Reset()
	d.preParse.Reset()
	d.postParse.Reset()
	d.usageList = nil
	d.printList = nil
	d.shortNames = make(map[string]bool)
	return nil
}

func (d driver) Imports() []string {
	return []string{
		`"fmt"`,
		`"strconv"`,
		`"github.com/spf13/pflag"`,
	}
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	_ = d.Driver.VisitArgument(a)
	p := d.Last()
	typ := a.Type
	tp, ok := typ.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf(
			"driver %s: non primitive argument types are not supported, got an argument %s %s",
			generators.DriverNamePFlag,
			p.Name,
			typ.Type(),
		)
	}
	if err := d.argument(p.Name, a.Index, tp, p.Ellipsis); err != nil {
		return fmt.Errorf("driver %s: argument %w", generators.DriverNamePFlag, err)
	}
	return nil
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	_ = d.Driver.VisitFlag(f, g)
	p := d.Last()
	typ := f.Type
	tprt, ptr := f.Type.(gofire.TPtr)
	if ptr {
		typ = tprt.ETyp
	}
	full := p.Full
	if p.Ref != nil {
		full = fmt.Sprintf("%s.%s", p.Ref.Group(), full)
	}
	switch len(p.Short) {
	case 0:
	case 1:
		if d.shortNames[p.Short] {
			return fmt.Errorf("driver %s: short flag name %q has been already registred", generators.DriverNamePFlag, p.Short)
		}
		d.shortNames[p.Short] = true
	default:
		return fmt.Errorf("driver %s: short flag name %q is not supported", generators.DriverNamePFlag, p.Short)
	}
	if err := d.flag(p.Name, full, p.Short, typ, ptr, f.Default, f.Doc, f.Deprecated, f.Hidden); err != nil {
		return fmt.Errorf("driver %s: flag %w", generators.DriverNamePFlag, err)
	}
	return nil
}

func (d *driver) argument(name string, index uint64, t gofire.TPrimitive, ellipsis bool) error {
	k := t.Kind()
	if ellipsis {
		switch k {
		case gofire.Bool:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					for i := %d; i < pflag.NArg(); i++ {
						v, err := strconv.ParseBool(pflag.Arg(i))
						if err != nil {
							return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
						}
						%s = append(%s, v)
					}
				`,
				index,
				name,
				name,
			); err != nil {
				return err
			}
		case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					for i := %d; i < pflag.NArg(); i++ {
						v, err := strconv.ParseInt(pflag.Arg(i), 10, %d)
						if err != nil {
							return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
						}
						%s = append(%s, %s(v))
					}
				`,
				index,
				k.Base(),
				name,
				name,
				k.Type(),
			); err != nil {
				return err
			}
		case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					for i := %d; i < pflag.NArg(); i++ {
						v, err := strconv.ParseUint(pflag.Arg(i), 10, %d)
						if err != nil {
							return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
						}
						%s = append(%s, %s(v))
					}
				`,
				index,
				k.Base(),
				name,
				name,
				k.Type(),
			); err != nil {
				return err
			}
		case gofire.Float32, gofire.Float64:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					for i := %d; i < pflag.NArg(); i++ {
						v, err := strconv.ParseFloat(pflag.Arg(i), %d)
						if err != nil {
							return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
						}
						%s = append(%s, %s(v))
					}
				`,
				index,
				k.Base(),
				name,
				name,
				k.Type(),
			); err != nil {
				return err
			}
		case gofire.String:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					for i := %d; i < pflag.NArg(); i++ {
						%s = append(%s, pflag.Arg(i))
					}
				`,
				index,
				name,
				name,
			); err != nil {
				return err
			}
		default:
			return fmt.Errorf(
				"ellipsis type %s is not supported for an argument %s",
				t.Type(),
				name,
			)
		}
	} else {
		switch k {
		case gofire.Bool:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					{	
						const i = %d
						if pflag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						v, err := strconv.ParseBool(pflag.Arg(i))
						if err != nil {
							return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
						}
						%s = v
					}
				`,
				index,
				name,
			); err != nil {
				return err
			}
		case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					{
						const i = %d
						if pflag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						v, err := strconv.ParseInt(pflag.Arg(i), 10, %d)
						if err != nil {
							return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
						}
						%s = %s(v)
					}
				`,
				index,
				k.Base(),
				name,
				k.Type(),
			); err != nil {
				return err
			}
		case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					{
						const i = %d
						if pflag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						v, err := strconv.ParseUint(pflag.Arg(i), 10, %d)
						if err != nil {
							return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
						}
						%s = %s(v)
					}
				`,
				index,
				k.Base(),
				name,
				k.Type(),
			); err != nil {
				return err
			}
		case gofire.Float32, gofire.Float64:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					{
						const i = %d
						if pflag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						v, err := strconv.ParseFloat(pflag.Arg(i), %d)
						if err != nil {
							return fmt.Errorf("argument %%d-th parse error: %%v", i, err)
						}
						%s = %s(v)
					}
				`,
				index,
				k.Base(),
				name,
				k.Type(),
			); err != nil {
				return err
			}
		case gofire.String:
			if _, err := fmt.Fprintf(&d.postParse,
				`
					{
						const i = %d
						if pflag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						%s = pflag.Arg(i)
					}
				`,
				index,
				name,
			); err != nil {
				return err
			}
		default:
			return fmt.Errorf(
				"type %s is not supported for an argument %s",
				t.Type(),
				name,
			)
		}
	}
	symb := "arg"
	if ellipsis {
		symb += "..."
	}
	d.usageList = append(d.usageList, fmt.Sprintf("arg%d", index))
	d.printList = append(d.printList, fmt.Sprintf("%s %d %s", symb, index, t.Type()))
	return nil
}

func (d *driver) flag(name, full, short string, t gofire.Typ, ptr bool, val string, doc string, deprecated, hidden bool) error {
	var amp string
	if ptr {
		amp = "&"
	}
	switch t.Kind() {
	case gofire.Slice:
		ts := t.(gofire.TSlice)
		etyp := ts.ETyp
		switch etyp.Kind() {
		case gofire.Bool:
			v, err := internal.ParseTypeValue(t, val)
			if err != nil || v == nil {
				return fmt.Errorf(
					"can't parse default value for a flag %s type %s: %w",
					full,
					t.Type(),
					err,
				)
			}
			if _, err := fmt.Fprintf(&d.preParse,
				`
					var %s_ []bool
					pflag.BoolSliceVarP(&%s_, %q, %q, %#v, %q)
				`,
				name,
				name,
				full,
				short,
				v,
				doc,
			); err != nil {
				return err
			}
			// second emit type cast to real flag type into post parse stage.
			if _, err := fmt.Fprintf(
				&d.postParse,
				`
					%s = %s%s_
				`,
				name,
				amp,
				name,
			); err != nil {
				return err
			}
		case gofire.Int32, gofire.Int64:
			v, err := internal.ParseTypeValue(t, val)
			if err != nil || v == nil {
				return fmt.Errorf(
					"can't parse default value for a flag %s type %s: %w",
					full,
					t.Type(),
					err,
				)
			}
			if _, err := fmt.Fprintf(&d.preParse,
				`
					var %s_ %s
					pflag.%sSliceVarP(&%s_, %q, %q, %#v, %q)
				`,
				name,
				ts.Type(),
				strings.Title(etyp.Type()),
				name,
				full,
				short,
				v,
				doc,
			); err != nil {
				return err
			}
			// second emit type cast to real flag type into post parse stage.
			if _, err := fmt.Fprintf(
				&d.postParse,
				`
					%s = %s%s_
				`,
				name,
				amp,
				name,
			); err != nil {
				return err
			}
		case gofire.Float32, gofire.Float64:
			v, err := internal.ParseTypeValue(t, val)
			if err != nil || v == nil {
				return fmt.Errorf(
					"can't parse default value for a flag %s type %s: %w",
					full,
					t.Type(),
					err,
				)
			}
			if _, err := fmt.Fprintf(&d.preParse,
				`
					var %s_ %s
					pflag.%sSliceVarP(&%s_, %q, %q, %#v, %q)
				`,
				name,
				ts.Type(),
				strings.Title(etyp.Type()),
				name,
				full,
				short,
				v,
				doc,
			); err != nil {
				return err
			}
			// second emit type cast to real flag type into post parse stage.
			if _, err := fmt.Fprintf(
				&d.postParse,
				`
					%s = %s%s_
				`,
				name,
				amp,
				name,
			); err != nil {
				return err
			}
		default:
			return fmt.Errorf(
				"type %s is not supported for a flag %s",
				t.Type(),
				full,
			)
		}
	case gofire.Bool:
		v, err := internal.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				full,
				t.Type(),
				err,
			)
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ bool
				pflag.BoolVarP(&%s_, %q, %q, %t, %q)
			`,
			name,
			name,
			full,
			short,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				%s = %s%s_
			`,
			name,
			amp,
			name,
		); err != nil {
			return err
		}
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		v, err := internal.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				full,
				t.Type(),
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				pflag.%sVarP(&%s_, %q, %q, %d, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			full,
			short,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				%s = %s%s_
			`,
			name,
			amp,
			name,
		); err != nil {
			return err
		}
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		v, err := internal.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				full,
				t.Type(),
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				pflag.%sVarP(&%s_, %q, %q, %d, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			full,
			short,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				%s = %s%s_
			`,
			name,
			amp,
			name,
		); err != nil {
			return err
		}
	case gofire.Float32, gofire.Float64:
		v, err := internal.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				full,
				t.Type(),
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				pflag.%sVarP(&%s_, %q, %q, %f, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			full,
			short,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				%s = %s%s_
			`,
			name,
			amp,
			name,
		); err != nil {
			return err
		}
	case gofire.String:
		v, err := internal.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				full,
				t.Type(),
				err,
			)
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ string
				pflag.StringVarP(&%s_, %q, %q, %q, %q)
			`,
			name,
			name,
			full,
			short,
			v,
			doc,
		); err != nil {
			return err
		}
		// second emit type cast to real flag type into post parse stage.
		if _, err := fmt.Fprintf(
			&d.postParse,
			`
				%s = %s%s_
			`,
			name,
			amp,
			name,
		); err != nil {
			return err
		}
	default:
		return fmt.Errorf(
			"type %s is not supported for a flag %s",
			t.Type(),
			full,
		)
	}
	if deprecated {
		if _, err := fmt.Fprintf(&d.preParse,
			`
				pflag.CommandLine.MarkDeprecated(%q, "deprecated: %s")
			`,
			full,
			doc,
		); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				pflag.CommandLine.MarkShorthandDeprecated(%q, "deprecated: %s")
			`,
			short,
			doc,
		); err != nil {
			return err
		}
	}
	if hidden {
		if _, err := fmt.Fprintf(&d.preParse,
			`
				pflag.CommandLine.MarkHidden(%q)
			`,
			full,
		); err != nil {
			return err
		}
		return nil
	}
	if val == "" {
		val = `""`
	}
	var pdeprecated string
	if deprecated {
		pdeprecated = "(DEPRECATED)"
	}
	u := fmt.Sprintf(`--%s=%v`, full, val)
	var pshort string
	if short != "" {
		u += " " + fmt.Sprintf(`-%s=%v`, short, val)
		pshort = fmt.Sprintf("-%s", short)
	}
	d.usageList = append(d.usageList, u)
	d.printList = append(
		d.printList,
		fmt.Sprintf("--%s %s %s %s (default %v) %s", full, pshort, t.Type(), doc, val, pdeprecated),
	)
	return nil
}
