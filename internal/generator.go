package internal

import "io"

type Generator interface {
	Generate(Command, io.Writer) error
}
