package generators

import (
	"bytes"
	"fmt"
	"strings"
)

type buffer bytes.Buffer

func (b *buffer) Append(format string, a ...interface{}) error {
	format = strings.Trim(format, " ")
	format = strings.Trim(format, "\t")
	format = strings.Trim(format, "\n")
	format += "\n"
	if _, err := fmt.Fprintf((*bytes.Buffer)(b), format, a...); err != nil {
		return err
	}
	return nil
}
