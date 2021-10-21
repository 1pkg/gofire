package gofire

import (
	"github.com/1pkg/gofire"
	"github.com/1pkg/gofire/parsers"
)

// ---

var parsetv = func(t gofire.Typ, val string, def interface{}) (interface{}, error) {
	v, set, err := parsers.ParseTypeValue(t, val)
	if err != nil {
		return nil, err
	}
	if set {
		return v, nil
	}
	return def, nil
}
