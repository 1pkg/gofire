package internal

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/1pkg/gofire/cmd"
	"github.com/1pkg/gofire/generators"
)

func GoExec(gocmd string, params ...string) Action {
	return func(ctx context.Context, dir string) (string, error) {
		exec := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("GO111MODULE=off go %s %s", gocmd, strings.Join(params, " ")))
		exec.Dir = dir
		exec.Env = os.Environ()
		out, err := exec.CombinedOutput()
		return string(out), err
	}
}

type Action func(context.Context, string) (string, error)

func (a Action) RunOnTest(ctx context.Context, name generators.DriverName, dir, pckg, function string) (string, error) {
	d, err := ioutil.TempDir("", "*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(d)
	if err := a.copy(dir, d); err != nil {
		return "", err
	}
	if _, err := cmd.Run(ctx, name, d, pckg, function); err != nil {
		return "", err
	}
	return a(ctx, d)
}

func (Action) copy(from, to string) error {
	fentries, err := fs.ReadDir(os.DirFS(from), ".")
	if err != nil {
		return err
	}
	for _, fentry := range fentries {
		fname := fentry.Name()
		if fentry.IsDir() || !strings.HasSuffix(fname, ".go") {
			continue
		}
		b, err := fs.ReadFile(os.DirFS(from), fname)
		if err != nil {
			return err
		}
		f, err := os.Create(filepath.Join(to, fname))
		if err != nil {
			return err
		}
		if _, err := f.Write(b); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}
