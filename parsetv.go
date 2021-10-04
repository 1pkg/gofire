package gofire

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseTypeValue(t Typ, val string) (interface{}, error) {
	k := t.Kind()
	switch k {
	case Array:
		return parseTypeValueRange(t.(TArray).ETyp, int(t.(TArray).Size), val)
	case Slice:
		return parseTypeValueRange(t.(TSlice).ETyp, -1, val)
	case Map:
		return parseTypeValueMap(t.(TMap).KTyp, t.(TMap).VTyp, val)
	case Bool:
		if val == "" {
			return false, nil
		}
		return strconv.ParseBool(val)
	case Int, Int8, Int16, Int32, Int64:
		if val == "" {
			return int64(0), nil
		}
		return strconv.ParseInt(val, 10, int(k.Base()))
	case Uint, Uint8, Uint16, Uint32, Uint64:
		if val == "" {
			return uint64(0), nil
		}
		return strconv.ParseUint(val, 10, int(k.Base()))
	case Float32, Float64:
		if val == "" {
			return float64(0.0), nil
		}
		return strconv.ParseFloat(val, int(k.Base()))
	case Complex64, Complex128:
		if val == "" {
			return complex128(0.0), nil
		}
		return strconv.ParseComplex(val, int(k.Base()))
	case String:
		if sval, err := strconv.Unquote(val); err == nil {
			return sval, nil
		}
		return val, nil
	}
	return nil, nil
}

// ! Note that some delimeters used in Splitb ('{', '}')
// ! inside default value might still break the parser.
// ! Current value parser may need to be revisited in future.
func parseTypeValueRange(t Typ, size int, val string) (interface{}, error) {
	if val == "" || val == "{}" {
		return []interface{}{}, nil
	}
	if !strings.HasPrefix(val, "{") || !strings.HasSuffix(val, "}") {
		return nil, fmt.Errorf("invalid value %q can't be parsed as an array or a slice", val)
	}
	nval := val[:len(val)-1][1:]
	if len(strings.TrimSpace(nval)) == 0 {
		return []interface{}{}, nil
	}
	pvals := Splitb(nval, ",", "{", "}")
	// Allow trailing coma in range definitions.
	if l := len(pvals); l > 0 && strings.TrimSpace(pvals[l-1]) == "" {
		pvals = pvals[:l-1]
	}
	// For arrays specifically we want to be sure that the sizes are matching.
	if size > -1 && len(pvals) != size {
		return nil, fmt.Errorf("invalid value %q can't be parsed as an array %d", val, size)
	}
	r := make([]interface{}, 0, len(pvals))
	for _, val := range pvals {
		v, err := ParseTypeValue(t, strings.TrimSpace(val))
		if err != nil {
			return nil, err
		}
		r = append(r, v)
	}
	return r, nil
}

// ! Note that some delimeters used in Splitb ('{', '}')
// ! inside default value might still break the parser.
// ! Current value parser may need to be revisited in future.
func parseTypeValueMap(tk, tv Typ, val string) (interface{}, error) {
	if val == "" || val == "{}" {
		return map[interface{}]interface{}{}, nil
	}
	if !strings.HasPrefix(val, "{") || !strings.HasSuffix(val, "}") {
		return nil, fmt.Errorf("invalid value %q can't be parsed as a map", val)
	}
	nval := val[:len(val)-1][1:]
	if len(strings.TrimSpace(nval)) == 0 {
		return map[interface{}]interface{}{}, nil
	}
	pairs := Splitb(nval, ",", "{", "}")
	// Allow trailing coma in map definitions.
	if l := len(pairs); l > 0 && strings.TrimSpace(pairs[l-1]) == "" {
		pairs = pairs[:l-1]
	}
	pkeys := make([]string, 0, len(pairs))
	pvals := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		p := strings.SplitN(pair, ":", 2)
		if len(p) != 2 {
			return nil, fmt.Errorf("invalid value %q can't be parsed as a map", val)
		}
		pkeys = append(pkeys, p[0])
		pvals = append(pvals, p[1])
	}
	mp := make(map[interface{}]interface{}, len(pairs))
	for i := range pairs {
		k, err := ParseTypeValue(tk, strings.TrimSpace(pkeys[i]))
		if err != nil {
			return nil, err
		}
		v, err := ParseTypeValue(tv, strings.TrimSpace(pvals[i]))
		if err != nil {
			return nil, err
		}
		mp[k] = v
	}
	return mp, nil
}
