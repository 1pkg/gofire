package intermediate

type Parameter interface {
	Accept(Visitor) error
}

type Argument struct {
	Index uint64
	Doc   string
	Kind  Kind
}

func (a Argument) Accept(s Visitor) error {
	return s.VisitArgument(a)
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
	Kind       Kind
}

func (f Flag) Accept(s Visitor) error {
	return s.VisitFlag(f)
}

type Group struct {
	Prefix     string
	Doc        string
	Parameters map[string]Parameter
}

func (g Group) Accept(s Visitor) error {
	cur := s.VisitGroup(g)
	if err := cur.Start(); err != nil {
		return err
	}
	for _, p := range g.Parameters {
		if err := p.Accept(s); err != nil {
			return err
		}
	}
	if err := cur.End(); err != nil {
		return err
	}
	return nil
}
