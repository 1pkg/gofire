package generators

import (
	"fmt"
	"io"
	"io/ioutil"
)

type buffer struct {
	rw io.ReadWriter
}

type byter interface {
	Bytes() ([]byte, error)
}

func (b buffer) fprintf(format string, a ...interface{}) error {
	format += "\n"
	if _, err := fmt.Fprintf(b.rw, format, a...); err != nil {
		return err
	}
	if _, err := b.rw.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

func (b buffer) bytes() ([]byte, error) {
	if bt, ok := b.rw.(byter); ok {
		return bt.Bytes()
	}
	return ioutil.ReadAll(b.rw)
}
