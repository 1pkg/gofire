package gofire

type Visitor interface {
	VisitPlaceholder(Placeholder) error
	VisitArgument(Argument) error
	VisitFlag(Flag, *Group) error
}

// Parameter defines an abstraction for cmd parameter.
type Parameter interface {
	Accept(Visitor) error
}

// Placeholder is a cmd parameter implementation
// that just hold parameter slot.
type Placeholder struct {
	Type Typ
}

func (p Placeholder) Accept(v Visitor) error {
	return v.VisitPlaceholder(p)
}

// Argument is a cmd parameter implementation
// that represents cmd positional argument.
type Argument struct {
	Index    uint64
	Ellipsis bool
	Type     Typ
}

func (a Argument) Accept(v Visitor) error {
	return v.VisitArgument(a)
}

// Flag is a cmd parameter implementation
// that represents cmd flag.
type Flag struct {
	Full       string
	Short      string
	Doc        string
	Deprecated bool
	Hidden     bool
	Default    interface{}
	Type       Typ
}

func (f Flag) Accept(v Visitor) error {
	return v.VisitFlag(f, nil)
}

// Group is a cmd parameter implementation
// that groups multiple cmd flags together.
type Group struct {
	Name  string
	Doc   string
	Flags []Flag
	Type  Typ
}

func (g Group) Accept(v Visitor) error {
	for _, f := range g.Flags {
		if err := v.VisitFlag(f, &g); err != nil {
			return err
		}
	}
	return nil
}

// Group is a cmd composite parameter implementation
// that represent function as a command.
type Command struct {
	Package    string
	Function   string
	Definition string
	Doc        string
	Context    bool
	Results    []string
	Parameters []Parameter
}

func (c Command) Accept(v Visitor) error {
	for _, p := range c.Parameters {
		if err := p.Accept(v); err != nil {
			return err
		}
	}
	return nil
}
