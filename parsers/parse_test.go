package parsers

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/1pkg/gofire"
)

type mapfs struct {
	fstest.MapFS
	fileErr error
	dirErr  error
}

func (fsm mapfs) ReadFile(name string) ([]byte, error) {
	f, _ := fsm.MapFS.ReadFile(name)
	return f, fsm.fileErr
}

func (fsm mapfs) ReadDir(name string) ([]fs.DirEntry, error) {
	dir, _ := fsm.MapFS.ReadDir(name)
	return dir, fsm.dirErr
}

func TestParse(t *testing.T) {
	escape := func(body string) []byte {
		return []byte(strings.ReplaceAll(body, "#", "`"))
	}
	table := map[string]struct {
		ctx      context.Context
		dir      fs.FS
		pckg     string
		function string
		cmd      *gofire.Command
		err      error
	}{
		"empty dir should produce expected error message": {
			ctx:      context.TODO(),
			dir:      fstest.MapFS{},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("function bar can't be found in ast package foo"),
		},
		"dir without go files should produce expected error message": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file_x.cpp": {
					Data: []byte(`#include "file_x.hpp"`),
				},
				"file_x.hpp": {
					Data: []byte("#include <iostream>"),
				},
				"dir": {
					Mode: fs.ModeDir,
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("function bar can't be found in ast package foo"),
		},
		"not valid go package with valid function definition should produce expected error message": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo1

						import "context"

						// bar function doc.
						func bar(ctx context.Context, a int8, b *string, c *float64) int {
							return 0
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("function bar can't be found in ast package foo"),
		},
		"invalid ast in go package function definition should produce expected error": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						import "context" |
						func bar(a int8, b *string, c *float64 aaaaaaa)
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast file file.go in package foo can't be parsed, 4:24: expected ';', found '|'"),
		},
		"error in reading fs dir should produce expected error": {
			ctx: context.TODO(),
			dir: mapfs{MapFS: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						import "context"

						// bar function doc.
						func bar(ctx context.Context, a int8, _ uint, b *string, c *float32, d *float64) int {
							return 0
						}
					`),
				},
			}, dirErr: errors.New("test error")},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast package foo fs dir can't be read, test error"),
		},
		"error in reading fs file should produce expected error": {
			ctx: context.TODO(),
			dir: mapfs{MapFS: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						import "context"

						// bar function doc.
						func bar(ctx context.Context, a int8, _ uint, b *string, c *float32, d *float64) int {
							return 0
						}
					`),
				},
			}, fileErr: errors.New("test error")},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast file file.go in package foo fs file can't be read, test error"),
		},
		"valid go package with valid function definition should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						import "context"

						// bar function doc.
						func bar(ctx context.Context, a int8, _ uint, b *string, c *float32, d *float64) int {
							return 0
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &gofire.Command{
				Package:  "foo",
				Function: "bar",
				Doc:      "bar function doc.",
				Context:  true,
				Parameters: []gofire.Parameter{
					gofire.Argument{Index: 0, Type: gofire.TPrimitive{TKind: gofire.Int8}},
					gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Uint}},
					gofire.Flag{Full: "b", Default: "", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.String}}},
					gofire.Flag{Full: "c", Default: "0.0", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Float32}}},
					gofire.Flag{Full: "d", Default: "0.0", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Float64}}},
				},
				Results: []string{"int"},
			},
		},
		"valid go package with valid function definition with ellipsis param should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						import "context"

						// bar function doc.
						func bar(ctx context.Context, a int8, _ uint, b *string, c *float32, d *float64, f ...int) int {
							return 0
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &gofire.Command{
				Package:  "foo",
				Function: "bar",
				Doc:      "bar function doc.",
				Context:  true,
				Parameters: []gofire.Parameter{
					gofire.Argument{Index: 0, Type: gofire.TPrimitive{TKind: gofire.Int8}},
					gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Uint}},
					gofire.Flag{Full: "b", Default: "", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.String}}},
					gofire.Flag{Full: "c", Default: "0.0", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Float32}}},
					gofire.Flag{Full: "d", Default: "0.0", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Float64}}},
					gofire.Argument{Index: 1, Ellipsis: true, Type: gofire.TPrimitive{TKind: gofire.Int}},
				},
				Results: []string{"int"},
			},
		},
		"valid go package with valid function definition with placeholder ellipsis param should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						import "context"

						// bar function doc.
						func bar(ctx context.Context, a int8, _ uint, b *string, c *float32, d *float64, _ ...int) int {
							return 0
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &gofire.Command{
				Package:  "foo",
				Function: "bar",
				Doc:      "bar function doc.",
				Context:  true,
				Parameters: []gofire.Parameter{
					gofire.Argument{Index: 0, Type: gofire.TPrimitive{TKind: gofire.Int8}},
					gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Uint}},
					gofire.Flag{Full: "b", Default: "", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.String}}},
					gofire.Flag{Full: "c", Default: "0.0", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Float32}}},
					gofire.Flag{Full: "d", Default: "0.0", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Float64}}},
					gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Int}},
				},
				Results: []string{"int"},
			},
		},
		"valid go package with valid function definition with flag ellipsis param should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						import "context"

						// bar function doc.
						func bar(ctx context.Context, a int8, _ uint, b *string, c *float32, d *float64, f ...*int) int {
							return 0
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &gofire.Command{
				Package:  "foo",
				Function: "bar",
				Doc:      "bar function doc.",
				Context:  true,
				Parameters: []gofire.Parameter{
					gofire.Argument{Index: 0, Type: gofire.TPrimitive{TKind: gofire.Int8}},
					gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Uint}},
					gofire.Flag{Full: "b", Default: "", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.String}}},
					gofire.Flag{Full: "c", Default: "0.0", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Float32}}},
					gofire.Flag{Full: "d", Default: "0.0", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Float64}}},
					gofire.Flag{Full: "f", Default: "0", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Int}}},
				},
				Results: []string{"int"},
			},
		},
		"valid go package with empty valid function definition should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						import "context"

						// bar function doc.
						func bar() {
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &gofire.Command{
				Package:  "foo",
				Function: "bar",
				Doc:      "bar function doc.",
			},
		},
		"valid go package with unsupported types in function definition should produce expected error": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						import "regexp"

						// bar function doc.
						func bar(*regexp.Regexp) {
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast file file.go in package foo function bar ast parsing error, parameter *regexp.Regexp type can't be parsed, unsupported complex type"),
		},
		"valid go package with valid function definition and group reference should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						func bar(int, int32, int64, z, ...int32) {
						}
					`),
				},
				"struct.go": {
					Data: escape(`
						package foo

						// z a flag group
						type z struct {
							a uint16
							b uint32
							c uint64
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &gofire.Command{
				Package:  "foo",
				Function: "bar",
				Parameters: []gofire.Parameter{
					gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Int}},
					gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Int32}},
					gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Int64}},
					gofire.Placeholder{Type: gofire.TStruct{Typ: "z"}},
					gofire.Placeholder{Type: gofire.TPrimitive{TKind: gofire.Int32}},
				},
			},
		},
		"valid go package with valid function definition with placeholders and group reference should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						func bar(a int8, b []string, cz z, _ z) {
						}
					`),
				},
				"struct.go": {
					Data: escape(`
						package foo

						// z a flag group
						type z struct {
							string
							a *string
							// b flag boolean
							b bool
							_ [10]int16
							complex map[string]int
						}
					`),
				},
				"interface.go": {
					Data: escape(`
						package foo

						// zi a empty iface
						type zi interface {}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &gofire.Command{
				Package:  "foo",
				Function: "bar",
				Parameters: []gofire.Parameter{
					gofire.Argument{Index: 0, Type: gofire.TPrimitive{TKind: gofire.Int8}},
					gofire.Argument{Index: 1, Type: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}}},
					gofire.Group{
						Name: "cz",
						Doc:  "z a flag group",
						Flags: []gofire.Flag{
							{Full: "a", Default: "nil", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.String}}},
							{Full: "b", Doc: "b flag boolean", Default: "false", Type: gofire.TPrimitive{TKind: gofire.Bool}},
							{Full: "complex", Default: "{}", Type: gofire.TMap{KTyp: gofire.TPrimitive{TKind: gofire.String}, VTyp: gofire.TPrimitive{TKind: gofire.Int}}},
						},
						Type: gofire.TStruct{Typ: "z"},
					},
					gofire.Placeholder{Type: gofire.TStruct{Typ: "z"}},
				},
			},
		},
		"valid go package with valid function definition and group reference with unsupported types should produce expected error": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						func bar(cz z) {
						}
					`),
				},
				"struct.go": {
					Data: escape(`
						package foo

						// z a flag group
						type z struct {
							iface zi
						}
					`),
				},
				"interface.go": {
					Data: escape(`
						package foo

						// zi a empty iface
						type zi interface {}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast file struct.go in package foo group type z ast parsing error, field iface zi type can't be parsed, unsupported primitive type invalid"),
		},
		"valid go package with valid function definition and group reference with tags should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						func bar(a int8, b []string, cz z) (r1, r2, r3 int){
							return 0, 0, 0 
						}
					`),
				},
				"struct.go": {
					Data: escape(`
						package foo

						type z struct {
							a, b string #gofire:"deprecated,default=str"#
							long complex128 #json:"long" gofire:"hidden=true,short=l"#
							c complex64 #gofire:"-"#
							d *uint8 #json:"d"#
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &gofire.Command{
				Package:  "foo",
				Function: "bar",
				Parameters: []gofire.Parameter{
					gofire.Argument{Index: 0, Type: gofire.TPrimitive{TKind: gofire.Int8}},
					gofire.Argument{Index: 1, Type: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}}},
					gofire.Group{
						Name: "cz",
						Flags: []gofire.Flag{
							{Full: "a", Deprecated: true, Default: "str", Type: gofire.TPrimitive{TKind: gofire.String}},
							{Full: "b", Deprecated: true, Default: "str", Type: gofire.TPrimitive{TKind: gofire.String}},
							{Full: "long", Short: "l", Hidden: true, Type: gofire.TPrimitive{TKind: gofire.Complex128}},
							{Full: "c", Default: "0", Type: gofire.TPrimitive{TKind: gofire.Complex64}},
							{Full: "d", Default: "nil", Type: gofire.TPtr{ETyp: gofire.TPrimitive{TKind: gofire.Uint8}}},
						},
						Type: gofire.TStruct{Typ: "z"},
					},
				},
				Results: []string{"int", "int", "int"},
			},
		},
		"valid go package with valid function definition and group reference with invalid tags format should produce expected error": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						func bar(az z) {
						}
					`),
				},
				"struct.go": {
					Data: escape(`
						package foo

						type z struct {
							bad string #gofire:"short=b=a=d"#
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast file struct.go in package foo group type z ast parsing error, field bad string `gofire:\"short=b=a=d\"` tag can't be parsed, can't parse tag short=b=a=d as key=value pair in gofire:\"short=b=a=d\""),
		},
		"valid go package with valid function definition and group reference with invalid tags should produce expected error": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						func bar(az z) {
						}
					`),
				},
				"struct.go": {
					Data: escape(`
						package foo

						type z struct {
							a string #gofire:"tag=true"#
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast file struct.go in package foo group type z ast parsing error, field a string `gofire:\"tag=true\"` tag can't be parsed, can't parse tag tag=true unsupported \"tag\" key in gofire:\"tag=true\""),
		},
		"valid go package with valid function definition and group reference with ambiguous short tags should produce expected error": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						func bar(az z) {
						}
					`),
				},
				"struct.go": {
					Data: escape(`
						package foo

						type z struct {
							a, b string #gofire:"short=c"#
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast file struct.go in package foo group type z ast parsing error, ambiguous short flag name c for multiple fields a, b string `gofire:\"short=c\"`"),
		},
		"valid go package with valid function definition and group reference with invalid string tags should produce expected error": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						func bar(az z) {
						}
					`),
				},
				"struct.go": {
					Data: escape(`
						package foo

						type z struct {
							a string #gofire:"default"#
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast file struct.go in package foo group type z ast parsing error, field a string `gofire:\"default\"` tag can't be parsed, can't parse tag default missing \"default\" key value in gofire:\"default\""),
		},
		"valid go package with valid function definition and group reference with invalid bool tags should produce expected error": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escape(`
						package foo

						func bar(az z) {
						}
					`),
				},
				"struct.go": {
					Data: escape(`
						package foo

						type z struct {
							a string #gofire:"hidden=10"#
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			err:      errors.New("ast file struct.go in package foo group type z ast parsing error, field a string `gofire:\"hidden=10\"` tag can't be parsed, can't parse tag hidden=10 as boolean for \"hidden\" key and 10 value in gofire:\"hidden=10\""),
		},
	}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			cmd, err := Parse(tcase.ctx, tcase.dir, tcase.pckg, tcase.function)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q", tcase.err, err)
			}
			if !reflect.DeepEqual(tcase.cmd, cmd) {
				t.Fatalf("expected cmd %#v but got %#v", tcase.cmd, cmd)
			}
		})
	}
}
