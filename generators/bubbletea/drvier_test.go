package bubbletea

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

func TestBubbleTeaDriver(t *testing.T) {
	table := map[string]struct {
		dir      string
		pckg     string
		function string
		err      error
	}{
		"echo no args types should produce expected output on valid params": {
			dir:      "echo_primitive_args",
			pckg:     "main",
			function: "echo",
		},
		"echo primitive args types should produce expected output on valid params": {
			dir:      "echo_primitive_args",
			pckg:     "main",
			function: "echo",
		},
		"echo flags with any types should fail on driver generation": {
			dir:      "echo_primitive_flags",
			pckg:     "main",
			function: "echo",
			err:      errors.New("driver bubbletea: doesn't support flags"),
		},
		"echo non primitive args types should fail on driver generation": {
			dir:      "echo_non_primitive_args",
			pckg:     "main",
			function: "echo",
			err:      errors.New("driver bubbletea: non primitive argument types are not supported, got an argument a0 map[string]bool"),
		},
		"echo ellipsis params types should produce expected output on valid params": {
			dir:      "echo_ellipsis_params",
			pckg:     "main",
			function: "echo",
			err:      errors.New("driver bubbletea: non primitive argument types are not supported, got an argument a0 int"),
		},
	}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			fs := os.DirFS(filepath.Join("tcases", tcase.dir))
			out, err := internal.GoExecute(context.TODO(), generators.DriverNameBubbleTea, fs, tcase.pckg, tcase.function, "build -tags=tcases")
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q\n%v", tcase.err, err, out)
			}
		})
	}
}
