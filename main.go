package main

import (
	"log"
	"os"

	"github.com/xyz/playground/generators"
	"github.com/xyz/playground/internal"
)

func main() {
	cmd := internal.Command{
		Pckg: "main",
		Parameters: []internal.Parameter{
			internal.Flag{
				Full:  "names",
				Short: "nms",
				Type: internal.TArray{
					ETyp: internal.TPrimitive{
						TKind: internal.String,
					},
					Size: 10,
				},
			},
			internal.Argument{
				Index: 1,
				Type: internal.TPrimitive{
					TKind: internal.Int,
				},
			},
		},
	}
	fire := generators.NewFire()
	if err := fire.Generate(cmd, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
