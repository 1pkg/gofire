package gofire

import (
	"context"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParse(t *testing.T) {
	errMsg := func(err error) string {
		if err == nil {
			return "nil"
		}
		return err.Error()
	}
	escapeGO := func(body string) []byte {
		return []byte(strings.ReplaceAll(body, "#", "`"))
	}
	table := map[string]struct {
		ctx      context.Context
		dir      fs.FS
		pckg     string
		function string
		cmd      *Command
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
		"not valid go package with simple valid function definition should produce expected error message": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escapeGO(`
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
		"valid go package with simple valid function definition should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escapeGO(`
						package foo

						import "context"

						// bar function doc.
						func bar(ctx context.Context, a int8, _ uint, b *string, c *float64) int {
							return 0
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &Command{
				Package:  "foo",
				Function: "bar",
				Doc:      "bar function doc.",
				Context:  true,
				Parameters: []Parameter{
					Argument{Index: 0, Type: TPrimitive{TKind: Int8}},
					Placeholder{Type: TPrimitive{TKind: Uint}},
					Flag{Full: "b", Optional: true, Default: "", Type: TPtr{ETyp: TPrimitive{TKind: String}}},
					Flag{Full: "c", Optional: true, Default: "0.0", Type: TPtr{ETyp: TPrimitive{TKind: Float64}}},
				},
				Results: []string{"int"},
			},
		},
		"valid go package with valid function definition and group reference should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escapeGO(`
						package foo

						func bar(a int8, b []string, cz z) {
						}
					`),
				},
				"struct.go": {
					Data: escapeGO(`
						package foo

						// z a flag group
						type z struct {
							a *string
							// b flag boolean
							b bool
							_ [10]int16
						}
					`),
				},
				"interface.go": {
					Data: escapeGO(`
						package foo

						// zi a empty iface
						type zi interface {}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &Command{
				Package:  "foo",
				Function: "bar",
				Parameters: []Parameter{
					Argument{Index: 0, Type: TPrimitive{TKind: Int8}},
					Argument{Index: 1, Type: TSlice{ETyp: TPrimitive{TKind: String}}},
					Group{
						Name: "cz",
						Doc:  "z a flag group",
						Flags: []Flag{
							{Full: "a", Optional: true, Default: "nil", Type: TPtr{ETyp: TPrimitive{TKind: String}}},
							{Full: "b", Doc: "b flag boolean", Optional: true, Default: "false", Type: TPrimitive{TKind: Bool}},
						},
						Type: TStruct{Typ: "z"},
					},
				},
			},
		},
		"valid go package with valid function definition and group reference with tags should produce expected command": {
			ctx: context.TODO(),
			dir: fstest.MapFS{
				"file.go": {
					Data: escapeGO(`
						package foo

						func bar(a int8, b []string, cz z) (r1, r2, r3 int){
							return 0, 0, 0 
						}
					`),
				},
				"struct.go": {
					Data: escapeGO(`
						package foo

						type z struct {
							a, b string #gofire:"deprecated,optional,default=str"#
							long complex128 #json:"long" gofire:"hidden=true,short=l"#
							c complex64 #gofire:"-"#
							d *uint8 #json:"d"#
						}
					`),
				},
			},
			pckg:     "foo",
			function: "bar",
			cmd: &Command{
				Package:  "foo",
				Function: "bar",
				Parameters: []Parameter{
					Argument{Index: 0, Type: TPrimitive{TKind: Int8}},
					Argument{Index: 1, Type: TSlice{ETyp: TPrimitive{TKind: String}}},
					Group{
						Name: "cz",
						Flags: []Flag{
							{Full: "a", Optional: true, Deprecated: true, Default: "str", Type: TPrimitive{TKind: String}},
							{Full: "b", Optional: true, Deprecated: true, Default: "str", Type: TPrimitive{TKind: String}},
							{Full: "long", Short: "l", Hidden: true, Type: TPrimitive{TKind: Complex128}},
							{Full: "c", Optional: true, Default: "0", Type: TPrimitive{TKind: Complex64}},
							{Full: "d", Optional: true, Default: "nil", Type: TPtr{ETyp: TPrimitive{TKind: Uint8}}},
						},
						Type: TStruct{Typ: "z"},
					},
				},
				Results: []string{"int", "int", "int"},
			},
		},
	}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			cmd, err := Parse(tcase.ctx, tcase.dir, tcase.pckg, tcase.function)
			if errMsg(tcase.err) != errMsg(err) {
				t.Fatalf("expected error message %q but got %q", errMsg(tcase.err), errMsg(err))
			}
			if !reflect.DeepEqual(tcase.cmd, cmd) {
				t.Fatalf("expected cmd %#v but got %#v", tcase.cmd, cmd)
			}
		})
	}
}
