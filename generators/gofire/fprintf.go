package gofire

import (
	"fmt"
)

type fprintf func(format string, a ...interface{}) *driver

func sderef(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (d *driver) ifAppendf(condition bool, format string, a ...interface{}) *driver {
	if condition {
		return d.appendf(format, a...)
	}
	return d
}

func (d *driver) ifElseAppendf(condition bool, format string, a ...interface{}) fprintf {
	if condition {
		d.appendf(format, a...)
		return d.emptyf
	} else {
		return d.appendf
	}
}

func (d *driver) appendf(format string, a ...interface{}) *driver {
	if _, err := fmt.Fprintf(d, format, a...); err != nil {
		panic(err)
	}
	return d
}

func (d *driver) emptyf(format string, a ...interface{}) *driver {
	return d
}
