package generators

import "github.com/1pkg/gofire"

type DriverName string

var (
	DriverNameGofire DriverName = "gofire"
)

type Driver interface {
	Imports() []string
	Parameters() []string
	Output() []byte
	gofire.Visitor
}
