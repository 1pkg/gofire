package main

type Kind uint8

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Float32
	Float64
	Complex64
	Complex128
	Array
	Interface
	Map
	Slice
	String
)

type Cursor interface {
	Start() error
	End() error
}

type Strategy interface {
	VisitArgument(Argument) error
	VisitFlag(Flag) error
	VisitGroup(Group) Cursor
}

type Parameter interface {
	Accept(Strategy) error
}

type Argument struct {
	Index uint64
	Doc   string
	Kind  Kind
}

func (a Argument) Accept(s Strategy) error {
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

func (f Flag) Accept(s Strategy) error {
	return s.VisitFlag(f)
}

type Group struct {
	Prefix     string
	Doc        string
	Parameters map[string]Parameter
}

func (g Group) Accept(s Strategy) error {
	cur := s.VisitGroup(g)
	if err := cur.Start(); err != nil {
		return err
	}
	// visit all group parameters.
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

type Command struct {
	Name       string
	Doc        string
	Func       string
	Parameters []Parameter
}

func main() {

}
