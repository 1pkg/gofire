package flag

import (
	"context"
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
		"echo primitive types should produce expected output on valid params": {
			dir:      "echo_primitive_types",
			pckg:     "main",
			function: "echo",
			params:   []string{"-a", "test", "-b", "100", "10.10"},
			out:      "a:test b:100 c:10.100000",
		},
	}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			fs := os.DirFS(filepath.Join("tcases", tcase.dir))
			out, err := internal.Execute(context.TODO(), generators.DriverNameFlag, fs, tcase.pckg, tcase.function, tcase.params...)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q", tcase.err, err)
			}
			if tcase.out != out {
				t.Fatalf("expected cmd %q but got %q", tcase.out, out)
			}
		})
	}
}
