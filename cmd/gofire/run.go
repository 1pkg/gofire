package main

import (
	"context"

	"github.com/1pkg/gofire/cmd"
	"github.com/1pkg/gofire/generators"
)

// run wraps cmd run for cli generators.
func run(ctx context.Context, name, dir, pckg, function string) error {
	return cmd.Run(ctx, generators.DriverName(name), dir, pckg, function)
}
