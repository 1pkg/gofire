package generators

import (
	"fmt"
	"strings"

	"github.com/1pkg/gofire"
)

type DriverName string

const (
	DriverNameGofire     DriverName = "gofire"
	DriverNameFlag      DriverName = "flag"
	DriverNamePFlag     DriverName = "pflag"
	DriverNameCobra     DriverName = "cobra"
	DriverNameBubbleTea DriverName = "bubbletea"
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
	Name     string
	Alt      string
	Type     gofire.Typ
	Ellipsis bool
	Doc      string
	Ref      *Reference
}

type Driver interface {
	Producer
	Imports() []string
	Parameters() []Parameter
	Template() string
	gofire.Visitor
}

type Producer interface {
	Output(gofire.Command) (string, error)
	Reset() error
}
