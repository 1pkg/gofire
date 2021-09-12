package internal

import (
	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/generators"
)

type cached struct {
	generators.Driver
	cache string
}

func Cached(d generators.Driver) generators.Driver {
	return &cached{Driver: d}
}

func (d *cached) Output(cmd gofire.Command) (string, error) {
	if d.cache != "" {
		return d.cache, nil
	}
	out, err := d.Driver.Output(cmd)
	if err != nil {
		return "", err
	}
	d.cache = out
	return out, nil
}

func (d *cached) Reset() error {
	if err := d.Driver.Reset(); err != nil {
		return err
	}
	d.cache = ""
	return nil
}
