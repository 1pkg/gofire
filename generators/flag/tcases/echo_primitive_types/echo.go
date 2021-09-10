package main

import (
	"context"
	"fmt"
)

func echo(_ context.Context, a *string, b *int, c float64) int {
	fmt.Printf("a:%s b:%d c:%f", *a, *b, c)
	return 0
}

func main() {
	Commandecho(context.TODO())
}
