package generators

import (
	"bytes"
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

func (b buffer) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(b.w); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
