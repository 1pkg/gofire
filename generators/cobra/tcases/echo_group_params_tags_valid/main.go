// +build tcases

package main

import "fmt"

type g struct {
	// flag 1 doc.
	flag1 int `gofire:"deprecated,default=10,short=a"`
	// flag 2 doc.
	flag2 string `gofire:"hidden,default=val,short=b"`
}

func echo(g1 g) {
	fmt.Printf("1:%d 2:%s\n", g1.flag1, g1.flag2)
}
