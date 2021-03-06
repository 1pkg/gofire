package cobra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/generators/internal"
)

func init() {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	_, _ = internal.GoExec("get github.com/spf13/cobra")(context.TODO(), dir)
}

func TestCobraDriver(t *testing.T) {
	unify := regexp.MustCompile(`\s|\n`)
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
		"echo primitive params types should produce expected output on help flag": {
			dir:      "echo_primitive_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"--help"},
			out: `echo: echo documentation string.
Usage: echo --a="" --b=0 --c=0 --d=false --e=0.000000 arg0 arg1 arg2 arg3 arg4
Flags: --a string --b int --c uint --d --e float32 -h, --help help for echo
`,
		},
		"echo primitive params types should produce expected error on invalid flags": {
			dir:      "echo_primitive_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"--e", "test"},
			err:      errors.New("exit status 1"),
			out: `Error: invalid argument "test" for "--e" flag: strconv.ParseFloat: parsing "test": invalid syntax
Usage: echo --a="" --b=0 --c=0 --d=false --e=0.000000 arg0 arg1 arg2 arg3 arg4
Flags: --a string --b int --c uint --d --e float32 -h, --help help for echo
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
			out: `Error: requires at least 5 arg(s), only received 1
Usage: echo --a="" --b=0 --c=0 --d=false --e=0.000000 arg0 arg1 arg2 arg3 arg4
Flags: --a string --b int --c uint --d --e float32 -h, --help help for echo
requires at least 5 arg(s), only received 1
exit status 2
`,
		},
		"echo non primitive args types should should fail on driver generation": {
			dir:      "echo_non_primitive_args",
			pckg:     "main",
			function: "echo",
			err:      errors.New("driver cobra: non primitive argument types are not supported, got an argument a0 map[string]bool"),
		},
		"echo non primitive flags map type types should should fail on driver generation": {
			dir:      "echo_non_primitive_flags_map",
			pckg:     "main",
			function: "echo",
			err:      errors.New("driver cobra: flag type map[string]bool is not supported for a flag a"),
		},
		"echo non primitive flags slice type should produce expected output on valid params": {
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
			params:   []string{"--g1.a=100"},
			out: `Flag --g1.a has been deprecated, deprecated: some fields doc.
1:100 2:10
`,
		},
		"echo group params types should produce expected error on invalid params": {
			dir:      "echo_group_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"--g1.a=group"},
			err:      errors.New("exit status 1"),
			out: `Error: invalid argument "group" for "--g1.a" flag: strconv.ParseInt: parsing "group": invalid syntax
Usage: echo --g1.a=10 --g1.b=10 --g2.a=10 --g2.b=10
Flags: -h, --help help for echo
invalid argument "group" for "--g1.a" flag: strconv.ParseInt: parsing "group": invalid syntax
exit status 2
`,
		},
		"echo group params with valid tags should produce expected output on invalid params": {
			dir:      "echo_group_params_tags_valid",
			pckg:     "main",
			function: "echo",
			params:   []string{"-b=test", "--g1.flag1=100"},
			out: `Flag --g1.flag1 has been deprecated, deprecated: flag 1 doc.
1:100 2:test 3:[10.5 10]
`,
		},
		"echo group params with invalid tags fail on driver generation": {
			dir:      "echo_group_params_tags_invalid",
			pckg:     "main",
			function: "echo",
			params:   []string{"-b=test", "--g1.flag1=100"},
			err:      errors.New(`driver cobra: short flag name "a" has been already registred`),
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
			exec := internal.GoExec("run -tags=tcases .", tcase.params...)
			out, err := exec.RunOnTest(context.TODO(), generators.DriverNameCobra, filepath.Join("tcases", tcase.dir), tcase.pckg, tcase.function)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q\n%v", tcase.err, err, out)
			}
			if unify.ReplaceAllString(tcase.out, "") != unify.ReplaceAllString(out, "") {
				t.Fatalf("expected output %q but got %q", tcase.out, out)
			}
		})
	}
}
