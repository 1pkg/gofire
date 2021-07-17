package intermediate

type Cursor interface {
	Start() error
	End() error
}

type Visitor interface {
	Dump() error
	VisitArgument(Argument) error
	VisitFlag(Flag) error
	VisitGroup(Group) Cursor
	VisitCommand(Command) Cursor
}
