package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1pkg/gofire"
)

func ParseTypeValue(t gofire.Typ, val string) (interface{}, error) {
	k := t.Kind()
	switch k {
	case gofire.Array:
		return parseTypeValueRange(t.(gofire.TArray).ETyp, int(t.(gofire.TArray).Size), val)
	case gofire.Slice:
		return parseTypeValueRange(t.(gofire.TSlice).ETyp, -1, val)
	case gofire.Map:
		return parseTypeValueMap(t.(gofire.TMap).KTyp, t.(gofire.TMap).VTyp, val)
	case gofire.Bool:
		if val == "" {
			return false, nil
		}
		return strconv.ParseBool(val)
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		if val == "" {
			return int64(0), nil
		}
		return strconv.ParseInt(val, 10, int(k.Base()))
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		if val == "" {
			return uint64(0), nil
		}
		return strconv.ParseUint(val, 10, int(k.Base()))
	case gofire.Float32, gofire.Float64:
		if val == "" {
			return float64(0.0), nil
		}
		return strconv.ParseFloat(val, int(k.Base()))
	case gofire.Complex64, gofire.Complex128:
		if val == "" {
			return complex128(0.0), nil
		}
		return strconv.ParseComplex(val, int(k.Base()))
	case gofire.String:
		if sval, err := strconv.Unquote(val); err == nil {
			return sval, nil
		}
		return val, nil
	}
	return nil, nil
}

func parseTypeValueRange(t gofire.Typ, size int, val string) (interface{}, error) {
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
	pvals := strings.Split(nval, ",")
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

func parseTypeValueMap(tk, tv gofire.Typ, val string) (interface{}, error) {
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
	pairs := strings.Split(nval, ",")
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
