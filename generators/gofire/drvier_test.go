package gofire

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/generators/internal"
)

func init() {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	_, _ = internal.GoExec("get github.com/1pkg/gofire")(context.TODO(), dir)
	_, _ = internal.GoExec("get github.com/mitchellh/mapstructure")(context.TODO(), dir)
}

func TestGofireDriver(t *testing.T) {
	table := map[string]struct {
		dir      string
		pckg     string
		function string
		params   []string
		out      string
		err      error
	}{
		"echo primitive params types should produce expected output on valid params": {
			dir:      "echo_primitive_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"--a=test", "--b", "100", "--c", "10", "--d=true", "--e", "10.125", "test1", "101", "11", "false", "10.5"},
			out:      "a:test b:100 c:10 d:true e:10.125 a1:test1 b1:101 c1:11 d1:false e1:10.500\n",
		},
		// 		"echo primitive params types should produce expected output on help flag": {
		// 			dir:      "echo_primitive_params",
		// 			pckg:     "main",
		// 			function: "echo",
		// 			params:   []string{"--help"},
		// 			err:      errors.New("exit status 1"),
		// 			out: `echo documentation string.
		// echo --a="" --b=0 --c=0 --d=false --e=0.000000 arg0 arg1 arg2 arg3 arg4 [--help -h]
		// func echo(_ context.Context, a *string, b *int, c *uint64, d *bool, e *float32, a1 string, b1 int, c1 uint64, d1 bool, e1 float32) int, --a string (default "") --b int (default 0) --c uint64 (default 0) --d bool (default false) --e float32 (default 0.000000) arg 0 string arg 1 int arg 2 uint64 arg 3 bool arg 4 float32
		// pflag: help requested
		// exit status 2
		// `,
		// 		},
		// 		"echo primitive params types should produce expected error on invalid flags": {
		// 			dir:      "echo_primitive_params",
		// 			pckg:     "main",
		// 			function: "echo",
		// 			params:   []string{"--e", "test"},
		// 			err:      errors.New("exit status 1"),
		// 			out: `invalid argument "test" for "--e" flag: strconv.ParseFloat: parsing "test": invalid syntax
		// echo documentation string.
		// echo --a="" --b=0 --c=0 --d=false --e=0.000000 arg0 arg1 arg2 arg3 arg4 [--help -h]
		// func echo(_ context.Context, a *string, b *int, c *uint64, d *bool, e *float32, a1 string, b1 int, c1 uint64, d1 bool, e1 float32) int, --a string (default "") --b int (default 0) --c uint64 (default 0) --d bool (default false) --e float32 (default 0.000000) arg 0 string arg 1 int arg 2 uint64 arg 3 bool arg 4 float32
		// invalid argument "test" for "--e" flag: strconv.ParseFloat: parsing "test": invalid syntax
		// exit status 2
		// `,
		// 		},
		// 		"echo primitive params types should produce expected error on invalid args": {
		// 			dir:      "echo_primitive_params",
		// 			pckg:     "main",
		// 			function: "echo",
		// 			params:   []string{"--a=test", "--b", "100", "--c", "10", "--d=true", "--e", "10.125", "test1"},
		// 			err:      errors.New("exit status 1"),
		// 			out: `echo documentation string.
		// echo --a="" --b=0 --c=0 --d=false --e=0.000000 arg0 arg1 arg2 arg3 arg4 [--help -h]
		// func echo(_ context.Context, a *string, b *int, c *uint64, d *bool, e *float32, a1 string, b1 int, c1 uint64, d1 bool, e1 float32) int, --a string (default "") --b int (default 0) --c uint64 (default 0) --d bool (default false) --e float32 (default 0.000000) arg 0 string arg 1 int arg 2 uint64 arg 3 bool arg 4 float32
		// argument 1-th is required
		// exit status 2
		// `,
		// 		},
		// 		"echo group params types should produce expected output on valid params": {
		// 			dir:      "echo_group_params",
		// 			pckg:     "main",
		// 			function: "echo",
		// 			params:   []string{"--g1.a=100"},
		// 			out: `Flag --g1.a has been deprecated, deprecated: some fields doc.
		// 1:100 2:10
		// `,
		// 		},
		// 		"echo group params types should produce expected error on invalid params": {
		// 			dir:      "echo_group_params",
		// 			pckg:     "main",
		// 			function: "echo",
		// 			params:   []string{"--g1.a=group"},
		// 			err:      errors.New("exit status 1"),
		// 			out: `invalid argument "group" for "--g1.a" flag: strconv.ParseInt: parsing "group": invalid syntax
		// echo --g1.a=10 --g1.b=10 --g2.a=10 --g2.b=10 [--help -h]
		// func echo(g1 g, g2 g), --g1.a int some fields doc. (default 10) (DEPRECATED) --g1.b int some fields doc. (default 10) (DEPRECATED) --g2.a int some fields doc. (default 10) (DEPRECATED) --g2.b int some fields doc. (default 10) (DEPRECATED)
		// invalid argument "group" for "--g1.a" flag: strconv.ParseInt: parsing "group": invalid syntax
		// exit status 2
		// `,
		// 		},
		// 		"echo group params with valid tags should produce expected output on invalid params": {
		// 			dir:      "echo_group_params_tags_valid",
		// 			pckg:     "main",
		// 			function: "echo",
		// 			params:   []string{"-b=test", "--g1.flag1=100"},
		// 			out: `Flag --g1.flag1 has been deprecated, deprecated: flag 1 doc.
		// 1:100 2:test 3:[10.5 10]
		// `,
		// 		},
		// 		"echo group params with invalid tags fail on driver generation": {
		// 			dir:      "echo_group_params_tags_invalid",
		// 			pckg:     "main",
		// 			function: "echo",
		// 			params:   []string{"-b=test", "--g1.flag1=100"},
		// 			err:      errors.New(`driver pflag: short flag name "a" has been already registred`),
		// 		},
		"echo ellipsis params types should produce expected error": {
			dir:      "echo_ellipsis_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"1", "10", "100"},
			err:      errors.New(`driver gofire: ellipsis argument types are not supported, got an argument a0 int`),
		},
	}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			exec := internal.GoExec("run -tags=tcases .", tcase.params...)
			out, err := exec.RunOnTest(context.TODO(), generators.DriverNameGofire, filepath.Join("tcases", tcase.dir), tcase.pckg, tcase.function)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q\n%v", tcase.err, err, out)
			}
			if tcase.out != out {
				t.Fatalf("expected output %q but got %q", tcase.out, out)
			}
		})
	}
}
