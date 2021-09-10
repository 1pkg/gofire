package utils

import (
	"bytes"
	"context"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/parsers"
)

func Execute(ctx context.Context, name generators.DriverName, dir fs.FS, pckg, function string, params ...string) (string, error) {
	cmd, err := parsers.Parse(ctx, dir, pckg, function)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := generators.Generate(ctx, name, *cmd, &b); err != nil {
		return "", err
	}
	d, err := ioutil.TempDir("", "*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(d)
	fentries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return "", err
	}
	for _, fentry := range fentries {
		fname := fentry.Name()
		if fentry.IsDir() || !strings.HasSuffix(fname, ".go") || strings.HasSuffix(fname, "_test.go") {
			continue
		}
		b, err := fs.ReadFile(dir, fname)
		if err != nil {
			return "", err
		}
		f, err := os.Create(filepath.Join(d, fname))
		if err != nil {
			return "", err
		}
		if _, err := f.Write(b); err != nil {
			return "", err
		}
	}
	f, err := os.Create(filepath.Join(d, "f.go"))
	if err != nil {
		return "", err
	}
	if _, err := f.Write(b.Bytes()); err != nil {
		return "", err
	}
	exec := exec.CommandContext(ctx, "sh", "-c", "GO111MODULE=off go run . "+strings.Join(params, " "))
	exec.Dir = d
	exec.Env = os.Environ()
	out, err := exec.CombinedOutput()
	return string(out), err
}
