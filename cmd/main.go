package main

import (
	"context"
	"log"
	"os"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
	_ "github.com/1pkg/gofire/generators/flag"
	_ "github.com/1pkg/gofire/generators/gofire.
	_ "github.com/1pkg/gofire/generators/pflag"
)

func main() {
	cmd := gofire.Command{
		Package:  "main",
		Function: "SomeFunc",
		Doc: `
		Some comments here
		zzzzz
		// aaaaaaaaaaaaa
		`,
		Context: true,
		Results: []string{"int", "string", "error", "string", "error"},
		Parameters: []gofire.Parameter{
			gofire.Flag{
				Full:  "aiiii",
				Short: "z",
				Type: gofire.TPtr{
					ETyp: gofire.TPrimitive{
						TKind: gofire.String,
					},
				},
			},
			gofire.Flag{
				Full:  "aiiiiaa",
				Short: "ttt",
				Type: gofire.TPtr{
					ETyp: gofire.TPrimitive{
						TKind: gofire.Uint16,
					},
				},
				Default:    "100",
				Deprecated: true,
				Doc:        "test",
			},
			gofire.Argument{
				Index: 0,
				Type: gofire.TPrimitive{
					TKind: gofire.Float32,
				},
			},
			gofire.Argument{
				Index: 1,
				Type: gofire.TPrimitive{
					TKind: gofire.Bool,
				},
			},
			// gofire.Flag{
			// 	Full:     "names",
			// 	Optional: true,
			// 	Default:  "{1,2,3,4}",
			// 	Type: gofire.TArray{
			// 		ETyp: gofire.TPrimitive{
			// 			TKind: gofire.String,
			// 		},
			// 		Size: 10,
			// 	},
			// },
			// gofire.Flag{
			// 	Full:  "mp",
			// 	Short: "p",
			// 	Type: gofire.TMap{
			// 		KTyp: gofire.TPrimitive{TKind: gofire.Int8},
			// 		VTyp: gofire.TMap{
			// 			KTyp: gofire.TPrimitive{TKind: gofire.String},
			// 			VTyp: gofire.TSlice{
			// 				ETyp: gofire.TPrimitive{TKind: gofire.String},
			// 			},
			// 		},
			// 	},
			// },
			// gofire.Flag{
			// 	Full:  "deep",
			// 	Short: "d",
			// 	Type: gofire.TSlice{
			// 		ETyp: gofire.TArray{
			// 			Size: 1,
			// 			ETyp: gofire.TSlice{
			// 				ETyp: gofire.TPrimitive{
			// 					TKind: gofire.Int64,
			// 				},
			// 			},
			// 		},
			// 	},
			// },
		},
	}
	if err := generators.Generate(context.TODO(), generators.DriverNameFlag, cmd, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
