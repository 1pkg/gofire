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
	_ "github.com/1pkg/gofire/generators/pflag"
	_ "github.com/1pkg/gofire/generators/reftype"
)

// Gofire 🔥 is command line interface generator tool.
// The first required argument dir represents directory path of source package.
// The second required argument fun represents source function name.
// Optional flag driver represents driver backend name, one of [flag, pflag, cobra, reftype, bubbletea], flag by default.
// Optional flag pckg represents source package name, useful if package name and directory is different, last element of dir by default.
func Gofire(ctx context.Context, driver, pckg *string, dir, fun string) {
	var d = "flag"
	if *driver == "" {
		driver = &d
	}
	if *pckg == "" {
		p := filepath.Base(dir)
		pckg = &p
	}
	p, err := cmd.Run(ctx, generators.DriverName(*driver), dir, *pckg, fun)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(p, "successfully generated")
}
