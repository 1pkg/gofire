package bubbletea

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
	_, _ = internal.GoExec("go get github.com/charmbracelet/bubbletea")(context.TODO(), dir)
	_, _ = internal.GoExec("go get github.com/charmbracelet/bubbles")(context.TODO(), dir)
	_, _ = internal.GoExec("go get github.com/charmbracelet/lipgloss")(context.TODO(), dir)
	_, _ = internal.GoExec("go get github.com/atotto/clipboard")(context.TODO(), dir)
}

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
			exec := internal.GoExec("build -tags=tcases .")
			out, err := exec.RunOnTest(context.TODO(), generators.DriverNameBubbleTea, filepath.Join("tcases", tcase.dir), tcase.pckg, tcase.function)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q\n%v", tcase.err, err, out)
			}
		})
	}
}
