package generators

import (
	"github.com/1pkg/gofire"
)

type DriverName string

const (
	DriverNameGofire DriverName = "gofire"
	DriverNameFlag  DriverName = "flag"
	DriverNamePFlag DriverName = "pflag"
)

type Parameter struct {
	Name    string
	Alt     string
	Type    gofire.Typ
	Default string
	Doc     string
}

type Producer interface {
	Output(gofire.Command) (string, error)
}

type ProducerFunc func(gofire.Command) (string, error)

func (f ProducerFunc) Output(cmd gofire.Command) (string, error) {
	return f(cmd)
}

type Driver interface {
	gofire.Visitor
	Reset() error
	Imports() []string
	Parameters() []Parameter
	Template() string
	Producer
}
