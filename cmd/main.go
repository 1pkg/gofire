package main

import (
	"log"
	"os"

	"github.com/xyz/gofire"
	"github.com/xyz/gofire/generators"
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
			gofire.Argument{
				Index: 1,
				Type: gofire.TPrimitive{
					TKind: gofire.Int,
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
