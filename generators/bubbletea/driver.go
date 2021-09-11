package bubbletea

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

func init() {
	generators.Register(generators.DriverNameBubbleTea, new(driver))
}

type driver struct {
	generators.BaseDriver
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
					input.CharLimit = 256
					m.inputs = append(m.inputs, input)
				}
			`,
			input,
		); err != nil {
			return "", err
		}
	}
	if len(d.inputList) > 0 {
		if _, err := buf.WriteString("m.inputs[0].Focus()"); err != nil {
			return "", err
		}
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
	if _, err := buf.Write(d.postParse.Bytes()); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (d *driver) Reset() error {
	_ = d.BaseDriver.Reset()
	d.postParse.Reset()
	d.inputList = nil
	return nil
}

func (d driver) Imports() []string {
	return []string{
		`"fmt"`,
		`"strconv"`,
		`"github.com/charmbracelet/bubbles/textinput"`,
		`"github.com/charmbracelet/bubbletea"`,
	}
}

func (d driver) Template() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	prefix := fmt.Sprintf("%x", rnd.Uint64())
	return fmt.Sprintf(`
		package {{.Package}}

		import(
			{{.Import}}
		)

		type %s{{.Function}} struct {
			index	int
			inputs	[]textinput.Model
			err		error
		}

		func (%s{{.Function}}) Init() bubbletea.Cmd {
			return textinput.Blink
		}
		
		func (m *%s{{.Function}}) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
			kmsg, ok := msg.(bubbletea.KeyMsg)
			if !ok {
				m.err = fmt.Errorf("received unexpected message %%v", msg)
				return m, bubbletea.Quit
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
		
		func (m %s{{.Function}}) View() string {
			var b strings.Builder
			for i := range m.inputs {
				b.WriteString(m.inputs[i].View())
				if i < len(m.inputs)-1 {
					b.WriteRune('\n')
				}
			}
			_, _ = fmt.Fprintf(&b, "\n\n[Submit]\n\n")
			return b.String()
		}
		
		{{.Doc}}
		func {{.Function}}(ctx context.Context) ({{.Return}}) {
			{{.Vars}}
			m := new(%s{{.Function}})
			{{.Body}}
			{{.Groups}}
			{{.Call}}
			return
		}
	`, prefix, prefix, prefix, prefix, prefix)
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
		return fmt.Errorf("driver %s: argument %w", generators.DriverNameBubbleTea, err)
	}
	return nil
}

func (d *driver) VisitFlag(f gofire.Flag, g *gofire.Group) error {
	return fmt.Errorf("driver %s: doesn't support flags", generators.DriverNameBubbleTea)
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
	d.inputList = append(d.inputList, fmt.Sprintf("arg %d %s", index, t.Type()))
	return nil
}
