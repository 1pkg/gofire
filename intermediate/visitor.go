package intermediate

type Cursor interface {
	Start() error
	End() error
}

type Visitor interface {
	VisitArgument(Argument) error
	VisitFlag(Flag) error
	VisitGroup(Group) Cursor
}
