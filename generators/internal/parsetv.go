package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1pkg/gofire"
)

func ParseTypeValue(t gofire.Typ, val string) (interface{}, error) {
	slice := func(val string) ([]string, error) {
		if val == "" || val == "{}" {
			return nil, nil
		}
		if !strings.HasPrefix(val, "{") || !strings.HasSuffix(val, "}") {
			return nil, fmt.Errorf("invalid value %q can't be parsed as a slice", val)
		}
		val = val[:len(val)-1][1:]
		if len(strings.TrimSpace(val)) == 0 {
			return nil, nil
		}
		return strings.Split(val, ","), nil
	}
	k := t.Kind()
	switch k {
	case gofire.Slice:
		ts := t.(gofire.TSlice)
		etyp := ts.ETyp
		ek := etyp.Kind()
		switch etyp.Kind() {
		case gofire.Bool:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]bool, 0, len(pvals))
			for _, val := range pvals {
				b, err := strconv.ParseBool(strings.TrimSpace(val))
				if err != nil {
					return nil, err
				}
				v = append(v, b)
			}
			return v, nil
		case gofire.Int, gofire.Int8, gofire.Int16, gofire.Int32, gofire.Int64:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]int64, 0, len(pvals))
			for _, val := range pvals {
				i, err := strconv.ParseInt(strings.TrimSpace(val), 10, int(ek.Base()))
				if err != nil {
					return nil, err
				}
				v = append(v, i)
			}
			return v, nil
		case gofire.Uint, gofire.Uint8, gofire.Uint16, gofire.Uint32, gofire.Uint64:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]uint64, 0, len(pvals))
			for _, val := range pvals {
				i, err := strconv.ParseUint(strings.TrimSpace(val), 10, int(ek.Base()))
				if err != nil {
					return nil, err
				}
				v = append(v, i)
			}
			return v, nil
		case gofire.Float32, gofire.Float64:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]float64, 0, len(pvals))
			for _, val := range pvals {
				f, err := strconv.ParseFloat(strings.TrimSpace(val), int(ek.Base()))
				if err != nil {
					return nil, err
				}
				v = append(v, f)
			}
			return v, nil
		case gofire.Complex64, gofire.Complex128:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]complex128, 0, len(pvals))
			for _, val := range pvals {
				c, err := strconv.ParseComplex(strings.TrimSpace(val), int(ek.Base()))
				if err != nil {
					return nil, err
				}
				v = append(v, c)
			}
			return v, nil
		case gofire.String:
			pvals, err := slice(val)
			if err != nil {
				return nil, err
			}
			v := make([]string, 0, len(pvals))
			for _, val := range pvals {
				s := strings.Replace(strings.TrimSpace(val), `"`, "", 2)
				v = append(v, s)
			}
			return v, nil
		}
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
		return val, nil
	}
	return nil, nil
}
