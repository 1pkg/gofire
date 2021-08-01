package generators

import "github.com/1pkg/gofire"

type DriverName string

const (
	DriverNameGofire DriverName = "gofire"
)

type Driver interface {
	Imports() []string
	Parameters() []string
	Output() []byte
	Reset() error
	gofire.Visitor
}
