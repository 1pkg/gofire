package parsers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1pkg/gofire"
)

// ParseTypeValue parses provided string value accordingly to the value type.
func ParseTypeValue(t gofire.Typ, val string) (interface{}, bool, error) {
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
			return k.Default(), false, nil
		}
		v, err := strconv.ParseBool(val)
		return v, err == nil, err
	case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
		if val == "" {
			return k.Default(), false, nil
		}
		v, err := strconv.ParseInt(val, 10, int(k.Base()))
		return v, err == nil, err
	case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
		if val == "" {
			return k.Default(), false, nil
		}
		v, err := strconv.ParseUint(val, 10, int(k.Base()))
		return v, err == nil, err
	case gofire.Float32, gofire.Float64:
		if val == "" {
			return k.Default(), false, nil
		}
		v, err := strconv.ParseFloat(val, int(k.Base()))
		return v, err == nil, err
	case gofire.Complex64, gofire.Complex128:
		if val == "" {
			return k.Default(), false, nil
		}
		v, err := strconv.ParseComplex(val, int(k.Base()))
		return v, err == nil, err
	case gofire.String:
		if sval, err := strconv.Unquote(val); err == nil {
			return sval, true, nil
		}
		return val, true, nil
	}
	return nil, false, nil
}

func parseTypeValueRange(t gofire.Typ, size int, val string) (interface{}, bool, error) {
	if val == "" || val == "{}" {
		return []interface{}{}, false, nil
	}
	if !strings.HasPrefix(val, "{") || !strings.HasSuffix(val, "}") {
		return nil, false, fmt.Errorf("invalid value %q can't be parsed as an array or a slice", val)
	}
	nval := val[:len(val)-1][1:]
	if len(strings.TrimSpace(nval)) == 0 {
		return []interface{}{}, false, nil
	}
	pvals := splitb(nval, ",", "{", "}")
	// Allow trailing coma in range definitions.
	if l := len(pvals); l > 0 && strings.TrimSpace(pvals[l-1]) == "" {
		pvals = pvals[:l-1]
	}
	// For arrays specifically we want to be sure that the sizes are matching.
	if size > -1 && len(pvals) != size {
		return nil, false, fmt.Errorf("invalid value %q can't be parsed as an array %d", val, size)
	}
	r := make([]interface{}, 0, len(pvals))
	for _, val := range pvals {
		v, _, err := ParseTypeValue(t, strings.TrimSpace(val))
		if err != nil {
			return nil, false, err
		}
		r = append(r, v)
	}
	return r, true, nil
}

func parseTypeValueMap(tk, tv gofire.Typ, val string) (interface{}, bool, error) {
	if val == "" || val == "{}" {
		return map[interface{}]interface{}{}, false, nil
	}
	if !strings.HasPrefix(val, "{") || !strings.HasSuffix(val, "}") {
		return nil, false, fmt.Errorf("invalid value %q can't be parsed as a map", val)
	}
	nval := val[:len(val)-1][1:]
	if len(strings.TrimSpace(nval)) == 0 {
		return map[interface{}]interface{}{}, false, nil
	}
	pairs := splitb(nval, ",", "{", "}")
	// Allow trailing coma in map definitions.
	if l := len(pairs); l > 0 && strings.TrimSpace(pairs[l-1]) == "" {
		pairs = pairs[:l-1]
	}
	pkeys := make([]string, 0, len(pairs))
	pvals := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		p := strings.SplitN(pair, ":", 2)
		if len(p) != 2 {
			return nil, false, fmt.Errorf("invalid value %q can't be parsed as a map", val)
		}
		pkeys = append(pkeys, p[0])
		pvals = append(pvals, p[1])
	}
	mp := make(map[interface{}]interface{}, len(pairs))
	for i := range pairs {
		k, _, err := ParseTypeValue(tk, strings.TrimSpace(pkeys[i]))
		if err != nil {
			return nil, false, err
		}
		v, _, err := ParseTypeValue(tv, strings.TrimSpace(pvals[i]))
		if err != nil {
			return nil, false, err
		}
		mp[k] = v
	}
	return mp, true, nil
}
