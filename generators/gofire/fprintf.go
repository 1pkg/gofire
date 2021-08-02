package gofire

import (
	"fmt"
)

type fprintf func(format string, a ...interface{}) *driver

func (d *driver) ifAppendf(condition bool, f func(fprintf)) *driver {
	if condition {
		f(d.appendf)
	}
	return d
}

func (d *driver) ifElseAppendf(condition bool, t func(fprintf), f func(fprintf)) *driver {
	if condition {
		t(d.appendf)
	} else {
		f(d.appendf)
	}
	return d
}

func (d *driver) appendf(format string, a ...interface{}) *driver {
	if _, err := fmt.Fprintf(d, format, a...); err != nil {
		panic(err)
	}
	return d
}
