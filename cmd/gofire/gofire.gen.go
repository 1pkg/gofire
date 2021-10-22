// THIS IS AUTOGENERATED FILE. DO NOT EDIT THIS FILE DIRECTLY.
// Generated using github.com/1pkg/gofire 🔥 2021-10-22T23:15:29+02:00.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"unicode"

	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/parsers"
	"github.com/mitchellh/mapstructure"
)

// Commandrun is autogenerated cli interface for run function.
func Commandrun(ctx context.Context) (o0 error, err error) {
	var a0 string
	var a1 string
	var a2 string
	var a3 string
	if err = func(ctx context.Context) (err error) {
		help := func() {
			doc, usage, list := "run wraps cmd run for cli generators.", "run arg0 arg1 arg2 arg3 [--help]", "func run(ctx context.Context, name, dir, pckg, function string) error, arg 0 string arg 1 string arg 2 string arg 3 string"
			if doc != "" {
				_, _ = fmt.Fprintln(flag.CommandLine.Output(), doc)
			}
			if usage != "" {
				_, _ = fmt.Fprintln(flag.CommandLine.Output(), usage)
			}
			if list != "" {
				_, _ = fmt.Fprintln(flag.CommandLine.Output(), list)
			}
		}
		defer func() {
			if err != nil {
				help()
			}
		}()
		var tokenize = func(tokens []string) (args []string, flags map[string]string, err error) {
			var flname = func(token string) (string, error) {
				fln := strings.Replace(token, "--", "", 1)
				for _, r := range fln {
					if r != '.' && !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
						return "", fmt.Errorf("flag name %s is not alphanumeric and can't be tokenized", fln)
					}
				}
				if _, ok := flags[fln]; ok {
					return "", fmt.Errorf("flag name %s used multiple times and can't be tokenized", fln)
				}
				return fln, nil
			}
			ltkns := len(tokens)
			args, flags = make([]string, 0, ltkns), make(map[string]string, ltkns)
			var fln, prevToken string
			for i := 0; i < ltkns; i++ {
				token := strings.TrimSpace(tokens[i])
				if token == "" {
					continue
				}
				iflag, iflagPrev := strings.HasPrefix(token, "--"), strings.HasPrefix(prevToken, "--")
				f, _ := flname(token)
				switch {
				case !iflag && !iflagPrev:
					if strings.HasPrefix(token, "-") {
						err = fmt.Errorf("short flag name %s can't be tokenized", token)
					} else {
						args = append(args, token)
						prevToken = ""
					}
				case iflag && strings.Contains(token, "="):
					parts := strings.SplitN(token, "=", 2)
					fln, err = flname(parts[0])
					flags[fln] = parts[1]
					prevToken = ""
				case !iflag && iflagPrev:
					fln, err = flname(prevToken)
					flags[fln] = token
					prevToken = ""
				case f == "help":
					args, flags = nil, map[string]string{"help": "true"}
					return
				case iflag && i != ltkns-1:
					prevToken = token
				default:
					err = fmt.Errorf("provided cli parameters %v can't be tokenized near token %s %s", tokens, prevToken, token)
				}
				if err != nil {
					args, flags = nil, nil
					return
				}
			}
			return
		}
		args, flags, err := tokenize(os.Args[1:])
		if err != nil {
			return err
		}
		if args == nil {
			args = nil
		}
		if flags == nil {
			flags = nil
		}
		if flags["help"] == "true" {
			return errors.New("help requested")
		}
		{
			i := 0
			if len(args) <= i {
				return fmt.Errorf("argument %d-th is required", i)
			}
			v, _, err := parsers.ParseTypeValue(gofire.TPrimitive{TKind: 0x10}, args[i])
			if err != nil {
				return fmt.Errorf("argument a0 value %v can't be parsed %v", args[i], err)
			}
			if err := mapstructure.Decode(v, &a0); err != nil {
				return fmt.Errorf("argument a0 value %v can't be decoded %v", v, err)
			}
		}
		{
			i := 1
			if len(args) <= i {
				return fmt.Errorf("argument %d-th is required", i)
			}
			v, _, err := parsers.ParseTypeValue(gofire.TPrimitive{TKind: 0x10}, args[i])
			if err != nil {
				return fmt.Errorf("argument a1 value %v can't be parsed %v", args[i], err)
			}
			if err := mapstructure.Decode(v, &a1); err != nil {
				return fmt.Errorf("argument a1 value %v can't be decoded %v", v, err)
			}
		}
		{
			i := 2
			if len(args) <= i {
				return fmt.Errorf("argument %d-th is required", i)
			}
			v, _, err := parsers.ParseTypeValue(gofire.TPrimitive{TKind: 0x10}, args[i])
			if err != nil {
				return fmt.Errorf("argument a2 value %v can't be parsed %v", args[i], err)
			}
			if err := mapstructure.Decode(v, &a2); err != nil {
				return fmt.Errorf("argument a2 value %v can't be decoded %v", v, err)
			}
		}
		{
			i := 3
			if len(args) <= i {
				return fmt.Errorf("argument %d-th is required", i)
			}
			v, _, err := parsers.ParseTypeValue(gofire.TPrimitive{TKind: 0x10}, args[i])
			if err != nil {
				return fmt.Errorf("argument a3 value %v can't be parsed %v", args[i], err)
			}
			if err := mapstructure.Decode(v, &a3); err != nil {
				return fmt.Errorf("argument a3 value %v can't be decoded %v", v, err)
			}
		}
		return
	}(ctx); err != nil {
		return
	}
	o0 = run(ctx, a0, a1, a2, a3)
	return
}

// auto generated main entrypoint.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	func(o0 error, err error) {
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}(Commandrun(ctx))
}
