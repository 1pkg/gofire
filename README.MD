<p align="center">
    <img src="https://raw.githubusercontent.com/1pkg/gofire/master/gopher.png?cache=0" alt="gofire"/>
</p>

# Gofire 🔥: Command Line Interface Generator tool for Go

[![lint](https://github.com/1pkg/gofire/workflows/lint/badge.svg)](https://github.com/1pkg/gofire/actions?query=workflow%3Alint+branch%3Amaster+)
[![build](https://github.com/1pkg/gofire/workflows/build/badge.svg)](https://github.com/1pkg/gofire/actions?query=workflow%3Abuild+branch%3Amaster+)
[![test](https://github.com/1pkg/gofire/workflows/test/badge.svg)](https://github.com/1pkg/gofire/actions?query=workflow%3Atest+branch%3Amaster+)
[![report](https://goreportcard.com/badge/github.com/1pkg/gofire?cache=0)](https://goreportcard.com/report/github.com/1pkg/gofire)
[![version](https://img.shields.io/github/go-mod/go-version/1pkg/gofire?cache=0)](https://github.com/1pkg/gofire/blob/master/go.mod)
[![license](https://img.shields.io/github/license/1pkg/gofire?cache=0)](LICENSE)

## Introduction

Gofire is the tool for Go that automatically generates a command line interface (CLI) for your functions and does all required plumbing in between, inspired by [python-fire](https://github.com/google/python-fire).

## Features

- Gofire uses code generation to generate simple and predictable CLI tailored to your code. It takes care about all required plumbing around parameters parsing, types casting, setting the entrypoint, documentation, etc.
- Gofire provides multiple CLI backends with different capabilities and features, including: [flag](https://pkg.go.dev/flag), [cobra](https://github.com/spf13/cobra), [bubbletea](https://github.com/charmbracelet/bubbletea), [and more](#drivers-and-backends).
- Gofire is opinionated about the defaults and always tries to provide simple and safe CLI. Yet, if needed, it supports optional go structs tag literals to flexibly configure CLI parameters [see more](#flags-groups-and-tags).

## Requirements Installation and Usage

Gofire uses go1.17, but it's most likely gonna work with other versions too. Note that Gofire is heavily relying on [ast](https://golang.org/pkg/go/ast/) packages, and the package might be slightly different among major go releases.

To install Gofire generator CLI tool, use:

```bash
go install github.com/1pkg/gofire/cmd/gofire@latest
```

To call Gofire generator CLI tool help, use:

```bash
gofire --help
Gofire 🔥 is command line interface generator tool.
The first required argument dir represents directory path of source package.
The second required argument fun represents source function name.
Optional flag driver represents driver backend name, one of [gofire, flag, pflag, cobra, bubbletea], flag by default.
Optional flag pckg represents source package name, useful if package name and directory is different, last element of dir by default.
Gofire --driver="" --pckg="" arg0 arg1 [--help]
func Gofire(ctx context.Context, driver, pckg *string, dir, fun string), --driver string (default "") --pckg string (default "") arg 0 string arg 1 string
help requested
```

To run Gofire generator CLI tool with `pflag` driver on function `Gofire` in package `main` in path `cmd/gofire`, use:

```bash
gofire --driver=pflag --pckg=main cmd/gofire Gofire
cmd/gofire/pflag.gen.go successfully generated
```

Note that Gofire can be easily integrated into the build process on permanent basis by adding the comment to your Go code base and using `go generate` command.

```go
//go:generate gofire --driver=$DRIVER --pckg=$PACKAGE $DIRECTORY $FUNCTION
```

For more details refer to [generating code in Go](https://go.dev/blog/generate).

## Parsing and Generation Convention

Currently Gofire works only with standalone top level functions. Where the name of the function conveniently represents the CLI command name and parametrs of the function represent CLI flags and positional arguments. Gofire parser generally supports all built-in Go types for the functions parameters including strings, slices and maps. However different driver backends may not support all parsed types for the code generation, to find what is supported by what driver backend refer to [drivers and backends](#drivers-and-backends). Note also that some built-in Go types including channels and interfaces don't have an obvious CLI parameters mapping and currently are not supported by Gofire. Type aliases currently are not supported by Gofire as well.

Gofire uses next convention while parsing a function and genereting a bridge to CLI:

- function name become command name.
- function definition become a part of command documentation.
- function doc string become a part of command documentation.
- function non pointer parametrs become required command positional arguments.
- function pointer parametrs become optional command auto flags with default empty values.
- function ellipsis parametr `...` is a special case that become ellipsis positional argument.
- function placeholder parametrs `_` is a special case that become filled with empty value internally.
- entrypoint for main is generated only if source function is located in `main` package.
- entrypoint for command is always generated as exported function in case you need to use it outside.

As an concise example the definition below is converted to:

```go
//                          optional float64 cli flag 'c'
//                                          │ default 0.0
//                                          │
//                                          │
//                                          │
//            2 required int positional argu│ents
//                              │           │
//                              │           │
//                              │           │optional []float64 cli flag 'in'
//         command name         │           │          │  default []float64{}
//               │              │           │          │
//               │              │           │          │
//               │              │           │          │
//               │              │           │          │
func addThenMultiplyAndGuess(a, b int, c *float64, in *[]float64, _ string) bool {
  sum := float64(a + b)
  guess := sum * (*c)
  for _, f := range *in {
    if guess == f {
      fmt.Println("inside")
      return true
    }
  }
  fmt.Println("outside")
  return false
}
```

```bash
# help
gofire --driver=pflag --pckg=main . addThenMultiplyAndGuess
addThenMultiplyAndGuess --c=0.000000 --in=[]float64{} arg0 arg1 [--help -h]
func addThenMultiplyAndGuess(a, b int, c *float64, in *[]float64) bool, --c float64 (default 0.000000) --in []float64 (default []float64{}) arg 0 int arg 1 int
pflag: help requested
exit status 2
# exec
go run ./... 10 1 --c=2.0 --in=10.0
outside
go run ./... 10 1 --c=2.0 --in=10.0 --in=22.0
inside
```

## Flags Groups and Tags

Gofire provides a way to bypass some rules defined in [parsing and generation convention](#parsing-and-generation-convention). Mainly grouping; adding defaults, short names, docs to CLI flags; and marking them as deprecated or hidden. This can be achieved by using a struct type as a function parameter together with special structure tag literals which acts as a flags group.

Gofire uses next schema for tag literals `gofire:"short=name,default=value,deprecated,hidden"`. Where `short` represents optional flag short name, `default` represents optional flag default value accordingly the type, `deprecated` represents optional flag deprecation status, `hidden` represents optional flag hidden status. Note that the structure has to be defined in the same package with the source function and that type aliases currently are not supported by Gofire.

As an concise example the definition below is converted to:

```go
package main

import "fmt"

type person struct {
	fullName string `gofire:"short=n"`
	// bornY is deprecated use bornYear isntead.
	bornY    int `gofire:"deprecated"`
	bornYear int
	// please use full occupation.
	occupation string `gofire:"default='professional wrestler'"`
}

// wasBorn answers the question if person was born in the provided year and prints the person.
func wasBorn(p person, bornY *int) {
	if p.bornYear <= *bornY {
		fmt.Println("yes")
	} else {
		fmt.Println("no")
	}
	fmt.Println(p)
}
```

```bash
# help
go run ./... --help
wasBorn: wasBorn answers the question if person was born in the provided year and prints the person.

Usage:
  wasBorn --bornY=0 --p.bornY=0 --p.bornYear=0 --p.fullName="" -n="" --p.occupation="professional wrestler"

Flags:
      --bornY int
  -h, --help                  help for wasBorn
      --p.bornYear int
  -n, --p.fullName string
      --p.occupation string   please use full occupation. (default "professional wrestler")
# exec
go run ./... -n="the Rock" --p.bornYear=1972 --bornY=1975
yes
{the Rock 0 1972 professional wrestler}
```

You can specify default values for comlex data types using simplified Go syntax for slice and map literals e.g. `{1,2,3}`; `{10:aaa, 20:bbb}`; `{'foo bar':{1:test, 2:'not test'}}`. Currently, there are few known minor limitations in complex value parsing in Gofire [see more](parsers/doc.go).

## Drivers and Backends

#### Flag Backend

Flag Backend is used as the backend by default in Gofire. This backend strives for simplicity and doesn't support any complex types. It also doesn't support short flag names and flags deprecation and hidding. However it still supports flags default values and ellipsis positional argument. Flag Backend is based on https://pkg.go.dev/flag package.

#### PFlag Backend

PFlag Backend is similar to Flag Backend, but it also adds extra support for slice type flags and fully supports Gofire tag literals including grouping, short flag names and flags deprecation and hidding. PFlag Backend is based on https://github.com/spf13/pflag package.

#### Cobra Backend

Cobra Backend mirrors PFlag Backend support for data types and tag literals. Cobra Backend is based on https://github.com/spf13/cobra package.

#### RefType Backend

RefType Backend aims to support complex nested data types including slice and maps for both flags and positional arguments. It doesn't support any tag literals features other then default values. Note that ellipsis parameters are not supported as they generally makes no sense for this driver. Also note that it uses github.com/mitchellh/mapstructure and reflection underneath. This driver expects all CLI values be prodived in their raw form not preprocessed by a shell, meaning you should always escape comlex data types values with `"` when using this driver e.g. `--stringMapOfIntSlices="{first:{1,2,3}, second:{0}, third:{3,2,1}}"`.

#### Bubbletea Backend

Bubbletea Backend generates a simple interactive TUI bridge. This backend only supports positional arguments and naturally doesn't support any flags groups tag literals or ellipsis parameters. Bubbletea Backend is based on https://github.com/charmbracelet/bubbletea.

## Licence

Gofire is licensed under the MIT License.  
See [LICENSE](LICENSE) for the full license text.
