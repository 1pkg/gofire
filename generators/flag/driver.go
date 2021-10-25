package flag

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
	generators.Register(generators.DriverNameFlag, internal.Cached(internal.Annotated(new(driver))))
}

type driver struct {
	internal.Driver
	preParse  bytes.Buffer
	postParse bytes.Buffer
	usageList []string
	printList []string
}

func (d driver) Output(cmd gofire.Command) (string, error) {
	var buf bytes.Buffer
	if _, err := buf.WriteString(
		`
			defer func() {
				if err != nil {
					flag.Usage()
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
			flag.Usage = func() {
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
		fmt.Sprintf("%s %s [-help -h]", cmd.Function, u),
		fmt.Sprintf("%s, %s", cmd.Definition, p),
	); err != nil {
		return "", err
	}
	if _, err := buf.WriteString("flag.Parse()"); err != nil {
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
	return nil
}

func (driver) Name() generators.DriverName {
	return generators.DriverNameFlag
}

func (d driver) Imports() []string {
	return []string{
		`"flag"`,
		`"fmt"`,
		`"strconv"`,
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
	if err := d.argument(p.Name, a.Index, tp, a.Ellipsis); err != nil {
		return fmt.Errorf("driver %s: argument %w", d.Name(), err)
	}
	return nil
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	_ = d.Driver.VisitFlag(f, g)
	p := d.Last()
	typ := p.Type
	tptr, ptr := typ.(gofire.TPtr)
	if ptr {
		typ = tptr.ETyp
	}
	tprim, ok := typ.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf(
			"driver %s: non pointer and non primitive flag types are not supported, got a flag %s %s",
			d.Name(),
			p.Name,
			f.Type.Type(),
		)
	}
	flag := p.Full
	if p.Ref != nil {
		flag = fmt.Sprintf("%s.%s", p.Ref.Group(), flag)
	}
	if err := d.flag(p.Name, flag, tprim, ptr, f.Default, p.Doc); err != nil {
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
					for i := %d; i < flag.NArg(); i++ {
						v, err := strconv.ParseInt(flag.Arg(i), 10, %d)
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
					for i := %d; i < flag.NArg(); i++ {
						v, err := strconv.ParseUint(flag.Arg(i), 10, %d)
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
					for i := %d; i < flag.NArg(); i++ {
						v, err := strconv.ParseFloat(flag.Arg(i), %d)
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
					for i := %d; i < flag.NArg(); i++ {
						%s = append(%s, flag.Arg(i))
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
						if flag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						v, err := strconv.ParseBool(flag.Arg(i))
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
						if flag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						v, err := strconv.ParseInt(flag.Arg(i), 10, %d)
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
						if flag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						v, err := strconv.ParseUint(flag.Arg(i), 10, %d)
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
						if flag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						v, err := strconv.ParseFloat(flag.Arg(i), %d)
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
						if flag.NArg() <= i {
							return fmt.Errorf("argument %%d-th is required", i)
						}
						%s = flag.Arg(i)
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

func (d *driver) flag(name string, flag string, t gofire.TPrimitive, ptr bool, val interface{}, doc string) error {
	k := t.Kind()
	var amp string
	if ptr {
		amp = "&"
	}
	var tp string
	switch k {
	case gofire.Bool:
		tp = k.Type()
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		tp = gofire.Int64.Type()
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		tp = gofire.Uint64.Type()
	case gofire.Float32, gofire.Float64:
		tp = gofire.Float64.Type()
	case gofire.String:
		tp = k.Type()
	default:
		return fmt.Errorf(
			"type %s is not supported for a flag %s",
			t.Type(),
			flag,
		)
	}
	if _, err := fmt.Fprintf(&d.preParse,
		`
			var %s_ %s
			flag.%sVar(&%s_, %q, %s, %q)
		`,
		name,
		tp,
		strings.Title(tp),
		name,
		flag,
		t.Format(val),
		doc,
	); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(
		&d.postParse,
		`
			{
				v := %s(%s_)
				%s = %sv
			}
		`,
		t.Type(),
		name,
		name,
		amp,
	); err != nil {
		return err
	}
	d.usageList = append(d.usageList, fmt.Sprintf("-%s=%s", flag, t.Format(val)))
	d.printList = append(d.printList, fmt.Sprintf("-%s %s %s (default %s)", flag, t.Type(), doc, t.Format(val)))
	return nil
}
