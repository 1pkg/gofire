package internal

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
	return v.VisitFlag(f)
}

type Group struct {
	Prefix     string
	Doc        string
	Parameters map[string]Parameter
}

func (g Group) Accept(v Visitor) error {
	if err := v.VisitGroupStart(g); err != nil {
		return err
	}
	for _, p := range g.Parameters {
		if err := p.Accept(v); err != nil {
			return err
		}
	}
	if err := v.VisitGroupEnd(g); err != nil {
		return err
	}
	return nil
}

type Command struct {
	Name       string
	Doc        string
	Pckg       string
	Func       string
	Parameters []Parameter
}

func (c Command) Accept(v Visitor) error {
	if err := v.VisitCommandStart(c); err != nil {
		return err
	}
	for _, p := range c.Parameters {
		if err := p.Accept(v); err != nil {
			return err
		}
	}
	if err := v.VisitCommandEnd(c); err != nil {
		return err
	}
	return nil
}
