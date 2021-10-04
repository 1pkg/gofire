package gofire

import "github.com/1pkg/gofire"

// ---

var parsetv = func(t gofire.Typ, val string, def string) (interface{}, error) {
	v, err := gofire.ParseTypeValue(t, val)
	if err != nil {
		return gofire.ParseTypeValue(t, def)
	}
	return v, nil
}
