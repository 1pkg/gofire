// +build tcases

package main

import "fmt"

type g struct {
	// some fields doc.
	a, b int `gofire:"deprecated,default=10"`
}

func echo(g1 g, g2 g) {
	fmt.Printf("1:%d 2:%d\n", g1.a, g2.a)
}
