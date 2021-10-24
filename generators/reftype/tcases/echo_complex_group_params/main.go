//go:build tcases

package main

import "fmt"

type g struct {
	a map[string]string `gofire:"deprecated,default={key:value}"`
	b map[string]string `gofire:"deprecated,default=nil"`
}

func echo(g g) {
	fmt.Println(g)
}
