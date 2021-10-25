package bubbletea

import (
	"bytes"
	"fmt"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/generators/internal"
)

func init() {
	generators.Register(generators.DriverNameBubbleTea, internal.Cached(internal.Annotated(new(driver))))
}

type driver struct {
	internal.Driver
	postParse bytes.Buffer
	inputList []string
}

func (d driver) Output(cmd gofire.Command) (string, error) {
	var buf bytes.Buffer
	for _, input := range d.inputList {
		if _, err := fmt.Fprintf(
			&buf,
			`
				{
					input := textinput.NewModel()
					input.Placeholder = %q
					input.CharLimit = 1024
					m.inputs = append(m.inputs, input)
				}
			`,
			input,
		); err != nil {
			return "", err
		}
	}
	if len(d.inputList) > 0 {
		if _, err := buf.WriteString("m.inputs[0].Focus();"); err != nil {
			return "", err
		}
	}
	if _, err := fmt.Fprintf(&buf, `m.doc = %q;`, cmd.Doc+"\n"+cmd.Definition); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(
		`
			if err = bubbletea.NewProgram(m).Start(); err != nil {
				return
			}
			if err = m.err; err != nil {
				return
			}
		`,
	); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(
		&buf,
		`
			if err = func() error {
				%s
				return nil
			}(); err != nil {
				return
			}
		`,
		d.postParse.String(),
	); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (d *driver) Reset() error {
	_ = d.Driver.Reset()
	d.postParse.Reset()
	d.inputList = nil
	return nil
}

func (driver) Name() generators.DriverName {
	return generators.DriverNameBubbleTea
}

func (d driver) Imports() []string {
	return []string{
		`"fmt"`,
		`"strconv"`,
		`"github.com/charmbracelet/bubbles/textinput"`,
		`bubbletea "github.com/charmbracelet/bubbletea"`,
	}
}

func (d driver) Template() string {
	return `
		package {{.Package}}

		import(
			{{.Import}}
		)

		type _bubbletea{{.Function}} struct {
			index	int
			inputs	[]textinput.Model
			doc 	string
			err		error
		}

		func (_bubbletea{{.Function}}) Init() bubbletea.Cmd {
			return textinput.Blink
		}
		
		func (m *_bubbletea{{.Function}}) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
			kmsg, ok := msg.(bubbletea.KeyMsg)
			if !ok {
				return m, nil
			}
			cmd := kmsg.String()
			switch cmd {
			case "ctrl+c", "esc":
				return m, bubbletea.Quit
			case "tab", "shift+tab", "enter", "up", "down":
				if cmd == "enter" && m.index == len(m.inputs) {
					return m, bubbletea.Quit
				}
				if cmd == "up" || cmd == "shift+tab" {
					m.index--
				} else {
					m.index++
				}
				if m.index > len(m.inputs) {
					m.index = 0
				} else if m.index < 0 {
					m.index = len(m.inputs)
				}
				cmds := make([]bubbletea.Cmd, len(m.inputs))
				for i := 0; i <= len(m.inputs)-1; i++ {
					if i == m.index {
						cmds[i] = m.inputs[i].Focus()
						continue
					}
					m.inputs[i].Blur()
				}
				return m, bubbletea.Batch(cmds...)
			default:
				var cmds = make([]bubbletea.Cmd, len(m.inputs))
				for i := range m.inputs {
					m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
				}
				return m, bubbletea.Batch(cmds...)
			}
		}
		
		func (m _bubbletea{{.Function}}) View() string {
			var b strings.Builder
			_, _ = fmt.Fprintf(&b, m.doc + "\n\n")
			for i := range m.inputs {
				b.WriteString(m.inputs[i].View())
				if i < len(m.inputs)-1 {
					b.WriteRune('\n')
				}
			}
			_, _ = fmt.Fprintf(&b, "\n\n[Execute]\n\n")
			return b.String()
		}
		
		{{.Doc}}
		func {{.Function}}(ctx context.Context) ({{.Return}}) {
			{{.Vars}}
			m := new(_bubbletea{{.Function}})
			{{.Body}}
			{{.Groups}}
			{{.Call}}
			return
		}
	`
}

func (d *driver) VisitArgument(a gofire.Argument) error {
	_ = d.Driver.VisitArgument(a)
	p := d.Last()
	tp, ok := p.Type.(gofire.TPrimitive)
	if !ok {
		return fmt.Errorf(
			"driver %s: non primitive argument types are not supported, got an argument %s %s",
			d.Name(),
			p.Name,
			a.Type.Type(),
		)
	}
	if err := d.argument(p.Name, a.Index, tp); err != nil {
		return fmt.Errorf("driver %s: argument %w", d.Name(), err)
	}
	return nil
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	return fmt.Errorf("driver %s: doesn't support flags", d.Name())
}

func (d *driver) argument(name string, index uint64, t gofire.TPrimitive) error {
	k := t.Kind()
	switch k {
	case gofire.Bool:
		if _, err := fmt.Fprintf(&d.postParse,
			`
				{	
					const i = %d
					if len(m.inputs) <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseBool(m.inputs[i].Value())
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
					if len(m.inputs) <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseInt(m.inputs[i].Value(), 10, %d)
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
					if len(m.inputs) <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseUint(m.inputs[i].Value(), 10, %d)
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
					if len(m.inputs) <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseFloat(m.inputs[i].Value(), %d)
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
	case gofire.Complex64, gofire.Complex128:
		if _, err := fmt.Fprintf(&d.postParse,
			`
				{
					const i = %d
					if len(m.inputs) <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					v, err := strconv.ParseComplex(m.inputs[i].Value(), %d)
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
					if len(m.inputs) <= i {
						return fmt.Errorf("argument %%d-th is required", i)
					}
					%s = m.inputs[i].Value()
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
	d.inputList = append(d.inputList, fmt.Sprintf("arg [%d] %s", index, t.Type()))
	return nil
}
