//go:build tcases

package main

import (
	"context"
	"fmt"
)

func echo(_ context.Context, a *[]int, b *[][]int, c map[string][]string) int {
	fmt.Println(*a, *b, c)
	return 0
}
