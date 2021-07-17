package generators

import (
	"errors"
	"fmt"

	"github.com/xyz/playground/intermediate"
)

type fire struct {
	buf buffer
	idx uint64
}

func (gen *fire) VisitArgument(a intermediate.Argument) error {
	name := fmt.Sprintf("a%d", a.Index)
	if err := gen.stmtDefinition(name, a.Kind); err != nil {
		return err
	}
	return nil
}

func (gen *fire) VisitFlag(f intermediate.Flag) error {
	name := fmt.Sprintf("f%s", f.Short)
	if err := gen.stmtDefinition(name, f.Kind); err != nil {
		return err
	}
	return nil
}

func (gen *fire) VisitGroup(g intermediate.Group) intermediate.Cursor {
	return nil
}

func (gen *fire) stmtDefinition(name string, k intermediate.Kind) error {
	return gen.buf.Append("var %s %s", name, k)
}

func (gen *fire) tmpVar(k intermediate.Kind) (string, error) {
	gen.idx++
	name := fmt.Sprintf("v%d", gen.idx)
	if err := gen.stmtDefinition(name, k); err != nil {
		return "", err
	}
	return name, nil
}

func (gen *fire) stmtAssignment(name string, k intermediate.Kind) error {
	switch k {
	case intermediate.Bool:
		tmpv, err := gen.tmpVar(k)
		if err != nil {
			return err
		}
		if err := gen.buf.Append(
			`
				%s, err = strconv.ParseBool(tokens[%s], 10, %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.String(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.Int, intermediate.Int8, intermediate.Int16, intermediate.Int32, intermediate.Int64:
		tmpv, err := gen.tmpVar(k)
		if err != nil {
			return err
		}
		if err := gen.buf.Append(
			`
				%s, err = strconv.ParseInt(tokens[%s], 10, %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.String(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.Uint, intermediate.Uint8, intermediate.Uint16, intermediate.Uint32, intermediate.Uint64:
		tmpv, err := gen.tmpVar(k)
		if err != nil {
			return err
		}
		if err := gen.buf.Append(
			`
				%s, err = strconv.ParseUint(tokens[%s], 10, %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.String(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.Float32, intermediate.Float64:
		tmpv, err := gen.tmpVar(k)
		if err != nil {
			return err
		}
		if err := gen.buf.Append(
			`
				%s, err = strconv.ParseFloat(tokens[%s], %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.String(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.Complex64, intermediate.Complex128:
		tmpv, err := gen.tmpVar(k)
		if err != nil {
			return err
		}
		if err := gen.buf.Append(
			`
				%s, err = strconv.ParseComplex(tokens[%s], %d)
				if err != nil {
					return err
				}
				%s = %s(%s)
			`,
			tmpv,
			name,
			k.Base(),
			name,
			k.String(),
			tmpv,
		); err != nil {
			return err
		}
	case intermediate.String:
		if err := gen.buf.Append("%s = tokens[%s]", name, name); err != nil {
			return err
		}
	default:
		return errors.New("")
	}
	return nil
}
