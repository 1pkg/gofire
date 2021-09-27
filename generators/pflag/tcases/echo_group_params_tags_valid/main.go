// +build tcases

package main

import "fmt"

type g struct {
	// flag 1 doc.
	flag1 int `gofire:"deprecated,default=10,short=a"`
	// flag 2 doc.
	flag2 string    `gofire:"hidden,default=val,short=b"`
	flag3 []float32 `json:"flag3" gofire:"default={10.5, 10},short=c"`
}

func echo(g1 g) {
	fmt.Printf("1:%d 2:%s 3:%v\n", g1.flag1, g1.flag2, g1.flag3)
}
