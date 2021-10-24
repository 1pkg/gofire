<p align="center">
    <img src="https://raw.githubusercontent.com/1pkg/gofire/master/gopher.png" alt="gofire"/>
</p>

# Gofire 🔥: Command Line Interface Generator tool for Go

[![lint](https://github.com/1pkg/gofire/workflows/lint/badge.svg)](https://github.com/1pkg/gofire/actions?query=workflow%3Alint+branch%3Amaster+)
[![build](https://github.com/1pkg/gofire/workflows/build/badge.svg)](https://github.com/1pkg/gofire/actions?query=workflow%3Abuild+branch%3Amaster+)
[![test](https://github.com/1pkg/gofire/workflows/test/badge.svg)](https://github.com/1pkg/gofire/actions?query=workflow%3Atest+branch%3Amaster+)
[![report](https://goreportcard.com/badge/github.com/1pkg/gofire?nocache=1)](https://goreportcard.com/report/github.com/1pkg/gofire)
[![version](https://img.shields.io/github/go-mod/go-version/1pkg/gofire)](https://github.com/1pkg/gofire/blob/master/go.mod)
[![license](https://img.shields.io/github/license/1pkg/gofire)](LICENSE)

## Introduction

Gofire automatically generates a command line interface (CLI) for your Go code, inspired by [python-fire](https://github.com/google/python-fire).

## Features

- Gofire uses code generation to generate predictable CLI tailored to your code. It takes care about all boilerplating around paramters parsing, types casting, setting the entrypoint, documentation, etc.
- Gofire provides multiple CLI backends with different capabilities and features, including: [flag](https://pkg.go.dev/flag), [cobra](https://github.com/spf13/cobra), [bubbletea](https://github.com/charmbracelet/bubbletea), and more.
- Gofire is opinionated about the defaults and always tries to generate simple and safe code. Yet it has optional go structs tag literals to flexibily configure CLI parameters if needed [see more](#gofire-and-tags).

## Requirements Installation and Usage

Gofire uses go1.17, but it's most likely gonna work with other versions too.  
Note that Gofire is heavily relying on [ast](https://golang.org/pkg/go/ast/) packages, and the package might be slightly different among major go releases.

To install and update Gofire CLI use:

```bash
go install github.com/1pkg/gofire@latest
```

```bash
gofire --helps
Gofire 🔥 is command line interface generator tool.
The first required argument dir represents directory path of source package.
The second required argument fun represents source function name.
Optional flag driver represents driver backend name, one of [gofire, flag, pflag, cobra, bubbletea], flag by default.
Optional flag pckg represents source package name, useful if package name and directory is different, last element of dir by default.
Gofire --driver="" --pckg="" arg0 arg1 [--help]
func Gofire(ctx context.Context, driver, pckg *string, dir, fun string), --driver string (default "") --pckg string (default "") arg 0 string arg 1 string
help requested
```

## Gofire and Tags

Gofire uses structure tag literals for flexible configuration of CLI parameters. Gofire uses next schema for tag literals `gofire:"short=name,deprecated,hidden,default=value"`. Where `short` represents optional flag short name, `deprecated` represents optional flag deprecation status, `hidden` represents optional flag hidden status, `default` represents optional flag default value for defined type.

## Licence

Gofire is licensed under the MIT License.  
See [LICENSE](LICENSE) for the full license text.
