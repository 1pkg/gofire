package main

import (
	"context"
	"log"
	"path/filepath"

	"github.com/1pkg/gofire/cmd"
	"github.com/1pkg/gofire/generators"
	_ "github.com/1pkg/gofire/generators/bubbletea"
	_ "github.com/1pkg/gofire/generators/cobra"
	_ "github.com/1pkg/gofire/generators/flag"
	_ "github.com/1pkg/gofire/generators/gofire"
	_ "github.com/1pkg/gofire/generators/pflag"
)

type prun struct {
	// driver name, one of [gofire, flag, pflag, cobra, bubbletea];
	driver string
	// directory path of source package;
	directory string
	// source function name;
	function string
}

// run wraps cmd run for cli generators.
func run(ctx context.Context, params prun) {
	dir := params.directory
	pckg := filepath.Base(dir)
	p, err := cmd.Run(ctx, generators.DriverName(params.driver), dir, pckg, params.function)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(p, "successfully generated")
}
