//go:build tcases

package main

import (
	"context"
	"log"
)

// echo documentation string.
func echo(_ context.Context, a *string, b *int, c *uint64, d *bool, e *float32, a1 string, b1 int, c1 uint64, d1 bool, e1 float32) int {
	log.Fatal("")
	return 0
}
