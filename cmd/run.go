package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/1pkg/gofire/generators"
	"github.com/1pkg/gofire/parsers"
)

// Run first parse provided package function, then
// generates relevant cli boilerplate and writes it to a file.
func Run(ctx context.Context, name generators.DriverName, dir, pckg, function string) error {
	cmd, err := parsers.Parse(ctx, os.DirFS(dir), pckg, function)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	if err := generators.Generate(ctx, name, *cmd, &b); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s.gen.go", name)))
	if err != nil {
		return err
	}
	if _, err := f.Write(b.Bytes()); err != nil {
		return err
	}
	return f.Close()
}
