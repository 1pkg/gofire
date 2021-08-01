package gofire

type Visitor interface {
	VisitArgument(Argument) error
	VisitFlag(Flag, *Group) error
}

type Parameter interface {
	Accept(Visitor) error
}

type Argument struct {
	Index uint64
	Doc   string
	Type  Typ
}

func (a Argument) Accept(v Visitor) error {
	return v.VisitArgument(a)
}

type Flag struct {
	Full       string
	Short      string
	Prefix     string
	Doc        string
	Optional   bool
	Deprecated bool
	Hidden     bool
	Default    string
	Type       Typ
}

func (f Flag) Accept(v Visitor) error {
	return v.VisitFlag(f, nil)
}

type Group struct {
	Prefix string
	Doc    string
	Flags  []Flag
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
	Name       string
	Doc        string
	Pckg       string
	Context    bool
	Returns    []string
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
