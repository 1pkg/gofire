package main

import (
	"github.com/xyz/playground/generators"
	"github.com/xyz/playground/intermediate"
)

func main() {
	cmd := intermediate.Command{
		Pckg: "main",
		Parameters: []intermediate.Parameter{
			intermediate.Flag{
				Full:  "names",
				Short: "nms",
				Type: intermediate.TArray{
					ETyp: intermediate.TPrimitive{
						TKind: intermediate.String,
					},
					Size: 10,
				},
			},
			intermediate.Argument{
				Index: 1,
				Type: intermediate.TPrimitive{
					TKind: intermediate.Int,
				},
			},
		},
	}
	v := generators.NewFire()
	_ = cmd.Accept(v)
	_ = v.Dump()
}
