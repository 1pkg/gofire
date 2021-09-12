package flag

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
			params:   []string{"-a", "test", "-b", "100", "-c", "10", "-d=true", "-e", "10.125", "test1", "101", "11", "false", "-10.5"},
			out:      "a:test b:100 c:10 d:true e:10.125 a1:test1 b1:101 c1:11 d1:false e1:-10.500",
		},
		"echo primitive params types should produce expected error on invalid params": {
			dir:      "echo_primitive_params",
			pckg:     "main",
			function: "echo",
			params:   []string{"-e", "test"},
			err:      errors.New("exit status 1"),
			out:      "invalid value \"test\" for flag -e: parse error\necho doc",
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
	}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			fs := os.DirFS(filepath.Join("tcases", tcase.dir))
			out, err := internal.GoExecute(context.TODO(), generators.DriverNameFlag, fs, tcase.pckg, tcase.function, "run -tags=tcases", tcase.params...)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q", tcase.err, err)
			}
			if !strings.Contains(out, tcase.out) {
				t.Fatalf("expected output %q should be inside %q", tcase.out, out)
			}
		})
	}
}
