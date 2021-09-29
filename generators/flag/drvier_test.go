package flag

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/generators/internal"
)

func TestFlagDriver(t *testing.T) {
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
			params:   []string{"-a=test", "-b", "100", "-c", "10", "-d=true", "-e", "10.125", "test1", "101", "11", "false", "-10.5"},
			out:      "a:test b:100 c:10 d:true e:10.125 a1:test1 b1:101 c1:11 d1:false e1:-10.500\n",
		},
		"echo primitive params types should produce expected output on help flag": {
			dir:      "echo_primitive_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"-help"},
			out: `echo documentation string.
echo -a="" -b=0 -c=0 -d=false -e=0.0 arg0 arg1 arg2 arg3 arg4 [-help -h]
func echo(_ context.Context, a *string, b *int, c *uint64, d *bool, e *float32, a1 string, b1 int, c1 uint64, d1 bool, e1 float32) int, -a string (default "") -b int (default 0) -c uint64 (default 0) -d bool (default false) -e float32 (default 0.0) arg 0 string arg 1 int arg 2 uint64 arg 3 bool arg 4 float32
`,
		},
		"echo primitive params types should produce expected error on invalid flags": {
			dir:      "echo_primitive_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"-e", "test"},
			err:      errors.New("exit status 1"),
			out: `invalid value "test" for flag -e: parse error
echo documentation string.
echo -a="" -b=0 -c=0 -d=false -e=0.0 arg0 arg1 arg2 arg3 arg4 [-help -h]
func echo(_ context.Context, a *string, b *int, c *uint64, d *bool, e *float32, a1 string, b1 int, c1 uint64, d1 bool, e1 float32) int, -a string (default "") -b int (default 0) -c uint64 (default 0) -d bool (default false) -e float32 (default 0.0) arg 0 string arg 1 int arg 2 uint64 arg 3 bool arg 4 float32
exit status 2
`,
		},
		"echo primitive params types should produce expected error on invalid args": {
			dir:      "echo_primitive_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"-a=test", "-b", "100", "-c", "10", "-d=true", "-e", "10.125", "test1"},
			err:      errors.New("exit status 1"),
			out: `echo documentation string.
echo -a="" -b=0 -c=0 -d=false -e=0.0 arg0 arg1 arg2 arg3 arg4 [-help -h]
func echo(_ context.Context, a *string, b *int, c *uint64, d *bool, e *float32, a1 string, b1 int, c1 uint64, d1 bool, e1 float32) int, -a string (default "") -b int (default 0) -c uint64 (default 0) -d bool (default false) -e float32 (default 0.0) arg 0 string arg 1 int arg 2 uint64 arg 3 bool arg 4 float32
argument 1-th is required
exit status 2
`,
		},
		"echo non primitive args types should should fail on driver generation": {
			dir:      "echo_non_primitive_args",
			pckg:     "main",
			function: "echo",
			err:      errors.New("driver flag: non primitive argument types are not supported, got an argument a0 map[string]bool"),
		},
		"echo non primitive flags types should should fail on driver generation": {
			dir:      "echo_non_primitive_flags",
			pckg:     "main",
			function: "echo",
			err:      errors.New("driver flag: non pointer and non primitive flag types are not supported, got a flag a *map[string]bool"),
		},
		"echo group params types should produce expected output on valid params": {
			dir:      "echo_group_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"-g1.a=100"},
			out:      "1:100 2:10\n",
		},
		"echo group params types should produce expected error on invalid params": {
			dir:      "echo_group_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"-g1.a=group"},
			err:      errors.New("exit status 1"),
			out: `invalid value "group" for flag -g1.a: parse error
echo -g1.a=10 -g1.b=10 -g2.a=10 -g2.b=10 [-help -h]
func echo(g1 g, g2 g), -g1.a int some fields doc. (default 10) -g1.b int some fields doc. (default 10) -g2.a int some fields doc. (default 10) -g2.b int some fields doc. (default 10)
exit status 2
`,
		},
		"echo ellipsis params types should produce expected output on valid params": {
			dir:      "echo_ellipsis_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"1", "10", "100"},
			out:      "[1 10 100]\n",
		},
	}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			fs := os.DirFS(filepath.Join("tcases", tcase.dir))
			out, err := internal.GoExecute(context.TODO(), generators.DriverNameFlag, fs, tcase.pckg, tcase.function, "run -tags=tcases", tcase.params...)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q\n%v", tcase.err, err, out)
			}
			if tcase.out != out {
				t.Fatalf("expected output %q but got %q", tcase.out, out)
			}
		})
	}
}
