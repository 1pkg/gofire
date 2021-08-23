package generators

import (
	"fmt"
	"strings"

	"github.com/1pkg/gofire"
)

type DriverName string

const (
	DriverNameGofire DriverName = "gofire"
	DriverNameFlag  DriverName = "flag"
	DriverNamePFlag DriverName = "pflag"
)

type Reference string

func NewReference(g, f string) *Reference {
	ref := Reference(fmt.Sprintf("%s.%s", g, f))
	return &ref
}

func (g *Reference) Group() string {
	if g == nil {
		return ""
	}
	return strings.Split(string(*g), ".")[0]
}

func (g *Reference) Field() string {
	if g == nil {
		return ""
	}
	return strings.Split(string(*g), ".")[1]
}

type Parameter struct {
	Name string
	Alt  string
	Type gofire.Typ
	Doc  string
	Ref  *Reference
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
