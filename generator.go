package gofire

import "io"

type Generator interface {
	Generate(Command, io.Writer) error
}

type Visitor interface {
	VisitArgument(Argument) error
	VisitFlag(Flag, *Group) error
}
