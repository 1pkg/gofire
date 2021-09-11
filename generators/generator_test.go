package generators

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/1pkg/gofire"
)

type driver struct {
	BaseDriver
	output   func(gofire.Command) (string, error)
	reset    func() error
	template func() string
}

func (d driver) Output(cmd gofire.Command) (string, error) {
	return d.output(cmd)
}

func (d driver) Reset() error {
	return d.reset()
}

func (d driver) Template() string {
	if d.template != nil {
		return d.template()
	}
	return d.BaseDriver.Template()
}

type writer func(p []byte) (int, error)

func (w writer) Write(p []byte) (int, error) {
	return w(p)
}

func TestGeneratorRegister(t *testing.T) {
	t.Run("should panic on nil driver", func(t *testing.T) {
		defer func(t *testing.T) {
			if err := recover(); fmt.Sprintf("%v", err) != "register driver is nil" {
				t.Fatalf("register should panic on nil driver with message %q", err)
			}
		}(t)
		Register(DriverName("test_register"), nil)
	})
	t.Run("should not panic on valid driver", func(t *testing.T) {
		d := &driver{}
		Register(DriverName("test_register"), d)
	})
	t.Run("should panic on duplicated driver", func(t *testing.T) {
		d := &driver{}
		defer func(t *testing.T) {
			if err := recover(); fmt.Sprintf("%v", err) != `register called twice for driver "test_register"` {
				t.Fatalf("register should panic on duplicated driver with message %q", err)
			}
		}(t)
		Register(DriverName("test_register"), d)
	})
}

func TestGeneratorGenerate(t *testing.T) {
	d := &driver{}
	Register(DriverName("test_generate"), d)
	t.Run("should fail on unregistered driver", func(t *testing.T) {
		err := Generate(context.TODO(), DriverName("test_generate_"), gofire.Command{}, nil)
		if fmt.Sprintf("%v", err) != `unknown driver "test_generate_" (forgotten import?)` {
			t.Fatalf("generate should fail on unregistered driver with message %q", err)
		}
	})
	t.Run("should fail on driver reset error", func(t *testing.T) {
		d.reset = func() error {
			return errors.New("test_reset")
		}
		err := Generate(context.TODO(), DriverName("test_generate"), gofire.Command{}, nil)
		if fmt.Sprintf("%v", err) != "test_reset" {
			t.Fatalf("generate should fail on driver reset error with message %q", err)
		}
	})
	t.Run("should fail on driver output error", func(t *testing.T) {
		d.reset = func() error {
			return nil
		}
		d.output = func(gofire.Command) (string, error) {
			return "", errors.New("test_output")
		}
		err := Generate(context.TODO(), DriverName("test_generate"), gofire.Command{}, nil)
		if fmt.Sprintf("%v", err) != "test_output" {
			t.Fatalf("generate should fail on driver output error with message %q", err)
		}
	})
	t.Run("should fail on driver broken template error", func(t *testing.T) {
		d.reset = func() error {
			return nil
		}
		d.output = func(gofire.Command) (string, error) {
			return "", nil
		}
		d.template = func() string {
			return "{{"
		}
		err := Generate(context.TODO(), DriverName("test_generate"), gofire.Command{}, nil)
		if fmt.Sprintf("%v", err) != `template: gen:4: unexpected "{" in command` {
			t.Fatalf("generate should fail on driver broken template error with message %q", err)
		}
	})
	t.Run("should fail on driver template expanding error", func(t *testing.T) {
		d.reset = func() error {
			return nil
		}
		d.output = func(gofire.Command) (string, error) {
			return "", nil
		}
		d.template = func() string {
			return "{{.Error}}"
		}
		err := Generate(context.TODO(), DriverName("test_generate"), gofire.Command{}, nil)
		if fmt.Sprintf("%v", err) != `template: gen:3:3: executing "gen" at <.Error>: can't evaluate field Error in type generators.proxy` {
			t.Fatalf("generate should fail on driver template expanding error with message %q", err)
		}
	})
	t.Run("should fail on driver code formating error", func(t *testing.T) {
		d.reset = func() error {
			return nil
		}
		d.output = func(gofire.Command) (string, error) {
			return "", nil
		}
		d.template = func() string {
			return "func {{.Import}}"
		}
		err := Generate(context.TODO(), DriverName("test_generate"), gofire.Command{}, nil)
		if fmt.Sprintf("%v", err) != "2:2: expected 'package', found 'func'" {
			t.Fatalf("generate should fail on driver code formating error with message %q", err)
		}
	})
	t.Run("should fail on internal writer error", func(t *testing.T) {
		d.reset = func() error {
			return nil
		}
		d.output = func(gofire.Command) (string, error) {
			return "", nil
		}
		d.template = nil
		cmd := gofire.Command{
			Package:  "test_package",
			Function: "test_function",
		}
		err := Generate(context.TODO(), DriverName("test_generate"), cmd, writer(func(p []byte) (int, error) {
			return 0, errors.New("test_write")
		}))
		if fmt.Sprintf("%v", err) != "test_write" {
			t.Fatalf("generate should fail on internal writer error with message %q", err)
		}
	})
	t.Run("should produce result into writer on valid preset", func(t *testing.T) {
		d.reset = func() error {
			return nil
		}
		d.output = func(gofire.Command) (string, error) {
			return "", nil
		}
		d.template = nil
		var buf bytes.Buffer
		cmd := gofire.Command{
			Package:  "main",
			Function: "test_function",
			Doc:      "test_doc",
			Context:  true,
			Results:  []string{"int", "string"},
			Parameters: []gofire.Parameter{
				gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Int}},
				gofire.Argument{Type: gofire.TPrimitive{TKind: gofire.Int}, Index: 0},
				gofire.Flag{Type: gofire.TPrimitive{TKind: gofire.Int}, Full: "flag", Short: "f"},
				gofire.Group{Type: gofire.TStruct{Typ: "test"}, Flags: []gofire.Flag{{Type: gofire.TPrimitive{TKind: gofire.Int}, Full: "f"}}, Name: "g"},
				gofire.Argument{Type: gofire.TPrimitive{TKind: gofire.Int}, Index: 1, Ellipsis: true},
			},
		}
		err := Generate(context.TODO(), DriverName("test_generate"), cmd, &buf)
		if fmt.Sprintf("%v", err) != "<nil>" {
			t.Fatalf("generate should not fail on valid preset %q", err)
		}
		if buf.String() == "" {
			t.Fatal("generate should produce non empty output")
		}
	})
}
