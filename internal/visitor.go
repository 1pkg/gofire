package internal

type Visitor interface {
	VisitArgument(Argument) error
	VisitFlag(Flag) error
	VisitGroupStart(Group) error
	VisitGroupEnd(Group) error
	VisitCommandStart(Command) error
	VisitCommandEnd(Command) error
}
