//go:build tcases

package main

import (
	"context"
	"fmt"
)

func echo(_ context.Context, a *[]bool) {
	fmt.Println(a)
}
