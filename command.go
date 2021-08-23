package gofire

type Visitor interface {
	VisitPlaceholder(Placeholder) error
	VisitArgument(Argument) error
	VisitFlag(Flag, *Group) error
}

type Parameter interface {
	Accept(Visitor) error
}

type Placeholder struct {
	Type Typ
}

func (p Placeholder) Accept(v Visitor) error {
	return v.VisitPlaceholder(p)
}

type Argument struct {
	Index    uint64
	Ellipsis bool
	Type     Typ
}

func (a Argument) Accept(v Visitor) error {
	return v.VisitArgument(a)
}

type Flag struct {
	Full       string
	Short      string
	Doc        string
	Deprecated bool
	Hidden     bool
	Default    string
	Type       Typ
}

func (f Flag) Accept(v Visitor) error {
	return v.VisitFlag(f, nil)
}

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

type Command struct {
	Package    string
	Function   string
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
