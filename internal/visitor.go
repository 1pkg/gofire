package internal

type Visitor interface {
	VisitArgument(Argument) error
	VisitFlag(Flag, *Group) error
}
