package internal

import "github.com/1pkg/gofire/generators"

type annotation struct {
	generators.Driver
}

func Annotated(d generators.Driver) generators.Driver {
	return annotation{Driver: d}
}

func (d annotation) Imports() []string {
	return append(d.Driver.Imports(), `"log"`, `"os"`, `"os/signal"`)
}

func (d annotation) Template() string {
	return `
		// TODO annotation goes here
	` + d.Driver.Template() + `
		{{ if eq .Package "main" }}
			func main() {
				ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
				defer stop()
				func({{.Return}}){
					if err != nil {
						log.Fatal(err)
					}
				}({{.Function}}(ctx))
			}
		{{ end }}
	`
}
