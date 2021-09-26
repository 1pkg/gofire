// +build tcases

package main

import "log"

type g struct {
	// flag 1 doc.
	flag1 int `gofire:"deprecated,default=10,short=a"`
	// flag 2 doc.
	flag2 string `gofire:"hidden,default=val,short=a"`
}

func echo(g1 g) {
	log.Fatal("")
}
