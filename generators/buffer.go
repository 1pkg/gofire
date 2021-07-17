package generators

import (
	"fmt"
	"io"
	"strings"
)

type buffer struct {
	w io.ReadWriter
}

func (b buffer) Append(format string, a ...interface{}) error {
	format = strings.Trim(format, " ")
	format = strings.Trim(format, "\t")
	format = strings.Trim(format, "\n")
	format += "\n"
	if _, err := fmt.Fprintf(b.w, format, a...); err != nil {
		return err
	}
	return nil
}

func (b buffer) Read(p []byte) (n int, err error) {
	return b.w.Read(p)
}
