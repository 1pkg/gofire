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
		fmt.Sprintf("%s %s [--help -h]", cmd.Function, u),
		fmt.Sprintf("%s, %s", cmd.Definition, p),
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

func (driver) Name() generators.DriverName {
	return generators.DriverNamePFlag
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
			d.Name(),
			p.Name,
			typ.Type(),
		)
	}
	if err := d.argument(p.Name, a.Index, tp, p.Ellipsis); err != nil {
		return fmt.Errorf("driver %s: argument %w", d.Name(), err)
	}
	return nil
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	_ = d.Driver.VisitFlag(f, g)
	p := d.Last()
	typ := p.Type
	tprt, ptr := typ.(gofire.TPtr)
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
			return fmt.Errorf("driver %s: short flag name %q has been already registred", d.Name(), p.Short)
		}
		d.shortNames[p.Short] = true
	default:
		return fmt.Errorf("driver %s: short flag name %q is not supported", d.Name(), p.Short)
	}
	if err := d.flag(p.Name, full, p.Short, typ, ptr, f.Default, f.Doc, f.Deprecated, f.Hidden); err != nil {
		return fmt.Errorf("driver %s: flag %w", d.Name(), err)
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

func (d *driver) flag(name, full, short string, t gofire.Typ, ptr bool, val interface{}, doc string, deprecated, hidden bool) error {
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
			fallthrough
		case gofire.Int32, gofire.Int64:
			fallthrough
		case gofire.Float32, gofire.Float64:
			fallthrough
		case gofire.String:
			if _, err := fmt.Fprintf(&d.preParse,
				`
					var %s_ %s
					pflag.%sSliceVarP(&%s_, %q, %q, %s, %q)
				`,
				name,
				ts.Type(),
				strings.Title(etyp.Type()),
				name,
				full,
				short,
				t.Format(val),
				doc,
			); err != nil {
				return err
			}
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
		fallthrough
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		fallthrough
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		fallthrough
	case gofire.Float32, gofire.Float64:
		fallthrough
	case gofire.String:
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				pflag.%sVarP(&%s_, %q, %q, %s, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			full,
			short,
			t.Format(val),
			doc,
		); err != nil {
			return err
		}
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
		if short != "" {
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
	var pdeprecated string
	if deprecated {
		pdeprecated = "(DEPRECATED)"
	}
	u := fmt.Sprintf("--%s=%s", full, t.Format(val))
	var pshort string
	if short != "" {
		u += " " + fmt.Sprintf("-%s=%s", short, t.Format(val))
		pshort = fmt.Sprintf("-%s", short)
	}
	d.usageList = append(d.usageList, u)
	d.printList = append(
		d.printList,
		fmt.Sprintf("--%s %s %s %s (default %s) %s", full, pshort, t.Type(), doc, t.Format(val), pdeprecated),
	)
	return nil
}
