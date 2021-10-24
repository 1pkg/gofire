package main

import (
	"context"
	"log"

	"github.com/1pkg/gofire/cmd"
	"github.com/1pkg/gofire/generators"
	_ "github.com/1pkg/gofire/generators/gofire"
)

func main() {
	if _, err := cmd.Run(
		context.Background(),
		generators.DriverNameGofire,
		"../gofire",
		"main",
		"Gofire",
	); err != nil {
		log.Fatalln(err)
	}
}
