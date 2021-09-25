package pflag

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

func TestPFlagDriver(t *testing.T) {
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
		"echo primitive params types should produce expected error on invalid flags": {
			dir:      "echo_primitive_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"--e", "test"},
			err:      errors.New("exit status 1"),
			out: `invalid argument "test" for "--e" flag: strconv.ParseFloat: parsing "test": invalid syntax
echo documentation string.
echo --a="" --b=0 --c=0 --d=false --e=0.0 arg0 arg1 arg2 arg3 arg4
--a string (default "") --b int (default 0) --c uint64 (default 0) --d bool (default false) --e float32 (default 0.0) arg 0 string arg 1 int arg 2 uint64 arg 3 bool arg 4 float32
invalid argument "test" for "--e" flag: strconv.ParseFloat: parsing "test": invalid syntax
exit status 2
`,
		},
		"echo primitive params types should produce expected error on invalid args": {
			dir:      "echo_primitive_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"--a=test", "--b", "100", "--c", "10", "--d=true", "--e", "10.125", "test1"},
			err:      errors.New("exit status 1"),
			out: `echo documentation string.
echo --a="" --b=0 --c=0 --d=false --e=0.0 arg0 arg1 arg2 arg3 arg4
--a string (default "") --b int (default 0) --c uint64 (default 0) --d bool (default false) --e float32 (default 0.0) arg 0 string arg 1 int arg 2 uint64 arg 3 bool arg 4 float32
argument 1-th is required
exit status 2
`,
		},
		"echo non primitive args types should should fail on driver generation": {
			dir:      "echo_non_primitive_args",
			pckg:     "main",
			function: "echo",
			err:      errors.New("driver pflag: non primitive argument types are not supported, got an argument a0 map[string]bool"),
		},
		"echo non primitive flags types should should fail on driver generation": {
			dir:      "echo_non_primitive_flags",
			pckg:     "main",
			function: "echo",
			err:      errors.New("driver pflag: flag type map[string]bool is not supported for a flag a"),
		},
		"echo non primitive flags slice produce expected output on valid params": {
			dir:      "echo_non_primitive_flags_slice",
			pckg:     "main",
			function: "echo",
			params:   []string{"--a=true", "--a=false", "--a=true"},
			out:      "&[true false true]\n",
		},
		"echo group params types should produce expected output on valid params": {
			dir:      "echo_group_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"--g1a=100"},
			out: `Flag --g1a has been deprecated, deprecated: some fields doc.
1:100 2:10
`,
		},
		"echo group params types should produce expected error on invalid params": {
			dir:      "echo_group_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"--g1a=group"},
			err:      errors.New("exit status 1"),
			out: `invalid argument "group" for "--g1a" flag: strconv.ParseInt: parsing "group": invalid syntax
echo --g1a=10 --g1b=10 --g2a=10 --g2b=10
--g1a int some fields doc. (default 10) (DEPRECATED) --g1b int some fields doc. (default 10) (DEPRECATED) --g2a int some fields doc. (default 10) (DEPRECATED) --g2b int some fields doc. (default 10) (DEPRECATED)
invalid argument "group" for "--g1a" flag: strconv.ParseInt: parsing "group": invalid syntax
exit status 2
`,
		},
		"echo short group params types should produce expected error on invalid params": {
			dir:      "echo_short_group_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"-b=test", "--g1flag1=100"},
			out: `Flag --g1flag1 has been deprecated, deprecated: flag 1 doc.
1:100 2:test
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
			out, err := internal.GoExecute(context.TODO(), generators.DriverNamePFlag, fs, tcase.pckg, tcase.function, "run -tags=tcases", tcase.params...)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q\n%v", tcase.err, err, out)
			}
			if tcase.out != out {
				t.Fatalf("expected output %q but got %q", tcase.out, out)
			}
		})
	}
}
