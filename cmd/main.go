package main

import (
	"log"
	"os"

	"github.com/1pkg/gofire"
	generators "github.com/1pkg/gofire/generators/gofire.
)

func main() {
	cmd := gofire.Command{
		Pckg: "main",
		Parameters: []gofire.Parameter{
			gofire.Flag{
				Full:  "names",
				Short: "nms",
				Type: gofire.TArray{
					ETyp: gofire.TPrimitive{
						TKind: gofire.String,
					},
					Size: 10,
				},
			},
			gofire.Flag{
				Full:  "mp",
				Short: "p",
				Type: gofire.TMap{
					KTyp: gofire.TPrimitive{TKind: gofire.Int8},
					VTyp: gofire.TMap{
						KTyp: gofire.TPrimitive{TKind: gofire.String},
						VTyp: gofire.TSlice{
							ETyp: gofire.TPrimitive{TKind: gofire.Interface},
						},
					},
				},
			},
			gofire.Flag{
				Full:  "deep",
				Short: "d",
				Type: gofire.TSlice{
					ETyp: gofire.TArray{
						Size: 1,
						ETyp: gofire.TSlice{
							ETyp: gofire.TPrimitive{
								TKind: gofire.Int64,
							},
						},
					},
				},
			},
		},
	}
	fire := &generators.Fire{}
	if err := fire.Generate(cmd, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
