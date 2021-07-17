package intermediate

type Command struct {
	Name       string
	Doc        string
	Pckg       string
	Func       string
	Parameters []Parameter
}

func (c Command) Accept(v Visitor) error {
	_ = v.VisitCommand(c)
	// if err := cur.Start(); err != nil {
	// 	return err
	// }
	for _, p := range c.Parameters {
		if err := p.Accept(v); err != nil {
			return err
		}
	}
	// if err := cur.End(); err != nil {
	// 	return err
	// }
	return nil
}
