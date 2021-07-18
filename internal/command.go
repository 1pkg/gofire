package internal

type Command struct {
	Name       string
	Doc        string
	Pckg       string
	Func       string
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
