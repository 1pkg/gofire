package cobra

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

func init() {
	generators.Register(generators.DriverNameCobra, new(driver))
}

type driver struct {
	generators.BaseDriver
	preParse  bytes.Buffer
	postParse bytes.Buffer
	usageList []string
	printList []string
	nargs     uint
}

func (d driver) Output(cmd gofire.Command) (string, error) {
	var buf bytes.Buffer
	if _, err := fmt.Fprintf(&buf, `
		parse = func(ctx context.Context) (err error) {
			%s
			return
		}
	`, d.postParse.String()); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(&buf, "cmd.Args = cobra.MinimumNArgs(%d);", d.nargs); err != nil {
		return "", err
	}
	digest := cmd.Function
	if len(cmd.Doc) > len(digest) {
		digest = strings.Split(cmd.Doc, "\n")[0]
	}
	sort.Strings(d.usageList)
	sort.Strings(d.printList)
	u := strings.Join(d.usageList, "\n")
	p := strings.Join(d.printList, "\n")
	if _, err := fmt.Fprintf(&buf, "cmd.Short = %q;", digest); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(&buf, "cmd.Long = %q;", fmt.Sprintf("%s\n%s\n%s", cmd.Doc, cmd.Function+" "+u, p)); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(&buf, "cmd.Use = %q;", cmd.Function+" "+u); err != nil {
		return "", err
	}
	if _, err := buf.Write(d.preParse.Bytes()); err != nil {
		return "", err
	}
	if _, err := buf.WriteString("err = cli.ExecuteContext(ctx)"); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (d *driver) Reset() error {
	_ = d.BaseDriver.Reset()
	d.preParse.Reset()
	d.postParse.Reset()
	d.usageList = nil
	d.printList = nil
	return nil
}

func (d driver) Imports() []string {
	return []string{
		`"fmt"`,
		`"strconv"`,
		``,
		`"github.com/spf13/cobra"`,
	}
}

func (d driver) Template() string {
	return `
		package {{.Package}}

		import(
			{{.Import}}
		)
		
		func Command{{.Function}}(ctx context.Context) ({{.Return}}) {
			{{.Vars}}
			var cli *cobra.Command
			var parse func(context.Context) error
			cli = &cobra.Command{
				RunE: func(cmd *cobra.Command, _ []string) (err error) {
					ctx := cmd.Context()
					if err = parse(ctx); err != nil {
						retrun
					}
					{{.Groups}}
					{{.Call}}
					return
				},
			}
			{{.Body}}
			return
		}
	`
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	_ = d.BaseDriver.VisitArgument(a)
	p := d.Last()
	tp, ok := a.Type.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf(
			"driver %s: non primitive argument types are not supported, got an argument %s %s",
			generators.DriverNameCobra,
			p.Name,
			a.Type.Type(),
		)
	}
	if err := d.argument(p.Name, a.Index, tp); err != nil {
		return fmt.Errorf("driver %s: argument %w", generators.DriverNameCobra, err)
	}
	return nil
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	_ = d.BaseDriver.VisitFlag(f, g)
	p := d.Last()
	typ := f.Type
	tprt, ptr := f.Type.(gofire.TPtr)
	if ptr {
		typ = tprt.ETyp
	}
	if err := d.flag(p.Name, p.Alt, typ, ptr, f.Default, f.Doc, f.Deprecated, f.Hidden); err != nil {
		return fmt.Errorf("driver %s: flag %w", generators.DriverNameCobra, err)
	}
	return nil
}

func (d *driver) argument(name string, index uint64, t gofire.TPrimitive) error {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		if _, err := fmt.Fprintf(&d.postParse,
			`
				{	
					const i = %d
					if cli.Flags().NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseBool(cli.Flags().Arg(i))
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
					if cli.Flags().NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseInt(cli.Flags().Arg(i), 10, %d)
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
					if cli.Flags().NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseUint(cli.Flags().Arg(i), 10, %d)
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
					if cli.Flags().NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseFloat(cli.Flags().Arg(i), %d)
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
					if cli.Flags().NArg() <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					%s = cli.Flags().Arg(i)
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
	d.nargs++
	d.usageList = append(d.usageList, fmt.Sprintf("arg%d", index))
	d.printList = append(d.printList, fmt.Sprintf("arg %d %s", index, t.Type()))
	return nil
}

func (d *driver) flag(name, short string, t gofire.Typ, ptr bool, val string, doc string, deprecated, hidden bool) error {
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
			v, err := d.ParseTypeValue(t, val)
			if err != nil || v == nil {
				return fmt.Errorf(
					"can't parse default value for a flag %s type %s: %w",
					name,
					t.Type(),
					err,
				)
			}
			if _, err := fmt.Fprintf(&d.preParse,
				`
					var %s_ []bool
					cli.Flags().BoolSliceVarP(&%s_, %q, %q, %#v, %q)
				`,
				name,
				name,
				name,
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
			v, err := d.ParseTypeValue(t, val)
			if err != nil || v == nil {
				return fmt.Errorf(
					"can't parse default value for a flag %s type %s: %w",
					name,
					t.Type(),
					err,
				)
			}
			if _, err := fmt.Fprintf(&d.preParse,
				`
					var %s_ %s
					cli.Flags().%sSliceVarP(&%s_, %q, %q, %#v, %q)
				`,
				name,
				ts.Type(),
				strings.Title(etyp.Type()),
				name,
				name,
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
			v, err := d.ParseTypeValue(t, val)
			if err != nil || v == nil {
				return fmt.Errorf(
					"can't parse default value for a flag %s type %s: %w",
					name,
					t.Type(),
					err,
				)
			}
			if _, err := fmt.Fprintf(&d.preParse,
				`
					var %s_ %s
					cli.Flags().%sSliceVarP(&%s_, %q, %q, %#v, %q)
				`,
				name,
				ts.Type(),
				strings.Title(etyp.Type()),
				name,
				name,
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
				name,
			)
		}
	case gofire.Bool:
		v, err := d.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ bool
				cli.Flags().BoolVarP(&%s_, %q, %q, %t, %q)
			`,
			name,
			name,
			name,
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
		v, err := d.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				cli.Flags().%sVarP(&%s_, %q, %q, %d, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			name,
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
		v, err := d.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				cli.Flags().%sVarP(&%s_, %q, %q, %d, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			name,
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
		v, err := d.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		// first emit temp bigger var flag holder into pre parse stage.
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ %s
				cli.Flags().%sVarP(&%s_, %q, %q, %f, %q)
			`,
			name,
			t.Type(),
			strings.Title(t.Type()),
			name,
			name,
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
		v, err := d.ParseTypeValue(t, val)
		if err != nil || v == nil {
			return fmt.Errorf(
				"can't parse default value for a flag %s type %s: %w",
				name,
				t.Type(),
				err,
			)
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				var %s_ string
				cli.Flags().StringVar(&%s_, %q, %q, %q, %q)
			`,
			name,
			name,
			name,
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
			name,
		)
	}
	if deprecated {
		if _, err := fmt.Fprintf(&d.preParse,
			`
				cli.Flags().MarkDeprecated(%q, "deprecated: %s")
			`,
			name,
			doc,
		); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(&d.preParse,
			`
				cli.Flags().MarkShorthandDeprecated(%q, "deprecated: %s")
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
				cli.Flags().MarkHidden(%q)
			`,
			name,
		); err != nil {
			return err
		}
	}
	if !hidden {
		var pdeprecated string
		if deprecated {
			pdeprecated = "(DEPRECATED)"
		}
		u := fmt.Sprintf(`--%s=""`, name)
		var pshort string
		if short != "" {
			u += " " + fmt.Sprintf(`-%s=""`, short)
		}
		d.usageList = append(d.usageList, u)
		d.printList = append(
			d.printList,
			fmt.Sprintf("--%s %s %s %s (default %q) %s", name, pshort, t.Type(), doc, val, pdeprecated),
		)
	}
	return nil
}
