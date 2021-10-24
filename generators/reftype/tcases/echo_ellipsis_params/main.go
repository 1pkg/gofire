//go:build tcases

package main

import "fmt"

func echo(a *int, b *int, c ...int) {
	fmt.Println(c)
}
