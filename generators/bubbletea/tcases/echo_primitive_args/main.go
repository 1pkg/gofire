// +build tcases

package main

import (
	"context"
	"fmt"
)

// echo documentation string.
func echo(_ context.Context, a1 string, b1 int, c1 uint64, d1 bool, e1 float32) int {
	fmt.Printf("a1:%s b1:%d c1:%d d1:%t e1:%.3f\n", a1, b1, c1, d1, e1)
	return 0
}
