package generators

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1pkg/gofire"
)

type DriverName string

const (
	DriverNameGofire     DriverName = "gofire"
	DriverNameFlag      DriverName = "flag"
	DriverNamePFlag     DriverName = "pflag"
	DriverNameCobra     DriverName = "cobra"
	DriverNameBubbleTea DriverName = "bubbletea"
)

type Reference string

func NewReference(g, f string) *Reference {
	ref := Reference(fmt.Sprintf("%s.%s", g, f))
	return &ref
}

func (g *Reference) Group() string {
	if g == nil {
		return ""
	}
	return strings.Split(string(*g), ".")[0]
}

func (g *Reference) Field() string {
	if g == nil {
		return ""
	}
	return strings.Split(string(*g), ".")[1]
}

type Parameter struct {
	Name     string
	Alt      string
	Type     gofire.Typ
	Ellipsis bool
	Doc      string
	Ref      *Reference
}

type Driver interface {
	Producer
	Imports() []string
	Parameters() []Parameter
	Template() string
	gofire.Visitor
}

type Producer interface {
	Output(gofire.Command) (string, error)
	Reset() error
}

type BaseDriver struct {
	params []Parameter
}

func (d *BaseDriver) Reset() error {
	d.params = nil
	return nil
}

func (d BaseDriver) Imports() []string {
	return []string{}
}

func (d BaseDriver) Parameters() []Parameter {
	return d.params
}

func (d BaseDriver) Last() *Parameter {
	l := len(d.params)
	if l == 0 {
		return nil
	}
	return &d.params[l-1]
}

func (d BaseDriver) Template() string {
	return `
		package {{.Package}}

		import(
			{{.Import}}
		)
		
		{{.Doc}}
		func {{.Function}}(ctx context.Context) ({{.Return}}) {
			{{.Vars}}
			if err = func(ctx context.Context) (err error) {
				{{.Body}}
				{{.Groups}}
				return
			}(ctx); err != nil {
				return
			}
			{{.Call}}
			return
		}
	`
}

func (d *BaseDriver) VisitPlaceholder(p gofire.Placeholder) error {
	d.params = append(d.params, Parameter{
		Name: fmt.Sprintf("p%d", len(d.params)),
		Type: p.Type,
	})
	return nil
}

func (d *BaseDriver) VisitArgument(a gofire.Argument) error {
	// For ellipsis argument we need to produce slice like parameter.
	typ := a.Type
	if a.Ellipsis {
		typ = gofire.TSlice{ETyp: typ}
	}
	d.params = append(d.params, Parameter{
		Name:     fmt.Sprintf("a%d", a.Index),
		Ellipsis: a.Ellipsis,
		Type:     typ,
	})
	return nil
}

func (d *BaseDriver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	var gname string
	var gdoc string
	if g != nil {
		gname = g.Name
		gdoc = g.Doc
	}
	name := fmt.Sprintf("%s%s", gname, f.Full)
	doc := fmt.Sprintf("%s %s", gdoc, f.Doc)
	var alt string
	if f.Short != "" {
		alt = fmt.Sprintf("%s%s", gname, f.Short)
	}
	var ref *Reference
	if g != nil {
		ref = NewReference(gname, f.Full)
	}
	d.params = append(d.params, Parameter{
		Name: name,
		Alt:  alt,
		Type: f.Type,
		Doc:  doc,
		Ref:  ref,
	})
	return nil
}

func (BaseDriver) ParseTypeValue(t gofire.Typ, val string) (interface{}, error) {
	slice := func(val string) ([]string, error) {
		if val == "" || val == "{}" {
			return nil, nil
		}
		if !strings.HasPrefix(val, "{") || !strings.HasSuffix(val, "}") {
			return nil, fmt.Errorf("invalid value %q can't be parsed as a slice", val)
		}
		val = val[:len(val)-1][1:]
		if len(strings.TrimSpace(val)) == 0 {
			return nil, nil
		}
		return strings.Split(val, ","), nil
	}
	k := t.Kind()
	switch k {
	case gofire.Slice:
		ts := t.(gofire.TSlice)
		etyp := ts.ETyp
		ek := etyp.Kind()
		switch etyp.Kind() {
		case gofire.Bool:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]bool, 0, len(pvals))
			for _, val := range pvals {
				b, err := strconv.ParseBool(strings.TrimSpace(val))
				if err != nil {
					return nil, err
				}
				v = append(v, b)
			}
			return v, nil
		case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]int64, 0, len(pvals))
			for _, val := range pvals {
				i, err := strconv.ParseInt(strings.TrimSpace(val), 10, int(ek.Base()))
				if err != nil {
					return nil, err
				}
				v = append(v, i)
			}
			return v, nil
		case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]uint64, 0, len(pvals))
			for _, val := range pvals {
				i, err := strconv.ParseUint(strings.TrimSpace(val), 10, int(ek.Base()))
				if err != nil {
					return nil, err
				}
				v = append(v, i)
			}
			return v, nil
		case gofire.Float32, gofire.Float64:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]float64, 0, len(pvals))
			for _, val := range pvals {
				f, err := strconv.ParseFloat(strings.TrimSpace(val), int(ek.Base()))
				if err != nil {
					return nil, err
				}
				v = append(v, f)
			}
			return v, nil
		case gofire.Complex64, gofire.Complex128:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]complex128, 0, len(pvals))
			for _, val := range pvals {
				c, err := strconv.ParseComplex(strings.TrimSpace(val), int(ek.Base()))
				if err != nil {
					return nil, err
				}
				v = append(v, c)
			}
			return v, nil
		case gofire.String:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]string, 0, len(pvals))
			for _, val := range pvals {
				s := strings.Replace(strings.TrimSpace(val), `"`, "", 2)
				v = append(v, s)
			}
			return v, nil
		}
	case gofire.Bool:
		if val == "" {
			return false, nil
		}
		return strconv.ParseBool(val)
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		if val == "" {
			return int64(0), nil
		}
		return strconv.ParseInt(val, 10, int(k.Base()))
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		if val == "" {
			return uint64(0), nil
		}
		return strconv.ParseUint(val, 10, int(k.Base()))
	case gofire.Float32, gofire.Float64:
		if val == "" {
			return float64(0.0), nil
		}
		return strconv.ParseFloat(val, int(k.Base()))
	case gofire.Complex64, gofire.Complex128:
		if val == "" {
			return complex128(0.0), nil
		}
		return strconv.ParseComplex(val, int(k.Base()))
	case gofire.String:
		return val, nil
	}
	return nil, nil
}

type annotation struct {
	Driver
}

func (d annotation) Imports() []string {
	return append(d.Driver.Imports(), `"log"`, `"os"`, `"os/signal"`)
}

func (d annotation) Template() string {
	return `
		// TODO annotation goes here
	` + d.Driver.Template() + `
		{{ if eq .Package "main" }}
			func main() {
				ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
				defer stop()
				{{ if eq .Result "" }}
					err := {{.Function}}(ctx)
				{{ else }}
					{{.Result}}, err := {{.Function}}(ctx)
				{{ end }}
				if err != nil {
					log.Fatal(err)
				}
				{{ if ne .Result "" }}
					log.Printf({{.Result}})
				{{ end }}
			}
		{{ end }}
	`
}
