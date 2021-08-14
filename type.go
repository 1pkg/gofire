package gofire

import (
	"fmt"
	"math"
)

type Kind uint8

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Float32
	Float64
	Complex64
	Complex128
	String
	Array
	Slice
	Map
	Ptr
	// Kinds bellow are not parsed and not processed by generators
	// but still defined here for visibility and potentially could be
	// processed in the future.
	UnsafePointer
	Chan
	Func
	Struct
	Interface
)

func (k Kind) Type() string {
	switch k {
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint:
		return "uint"
	case Uint8:
		return "uint8"
	case Uint16:
		return "uint16"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Complex64:
		return "complex64"
	case Complex128:
		return "complex128"
	case String:
		return "string"
	case Array:
		return "array"
	case Slice:
		return "slice"
	case Map:
		return "map"
	case Ptr:
		return "ptr"
	default:
		return "invalid"
	}
}

func (k Kind) Default() string {
	switch k {
	case Bool:
		return "false"
	case Int, Int8, Int16, Int32, Int64:
		return "0"
	case Uint, Uint8, Uint16, Uint32, Uint64:
		return "0"
	case Float32, Float64:
		return "0.0"
	case Complex64, Complex128:
		return "0"
	case String:
		return ""
	case Array, Slice, Map:
		return "{}"
	default:
		return "nil"
	}
}

func (k Kind) Base() int16 {
	switch k {
	case Int:
		return 64
	case Int8:
		return 8
	case Int16:
		return 16
	case Int32:
		return 32
	case Int64:
		return 64
	case Uint:
		return 64
	case Uint8:
		return 8
	case Uint16:
		return 16
	case Uint32:
		return 32
	case Uint64:
		return 64
	case Float32:
		return 32
	case Float64:
		return 64
	case Complex64:
		return 64
	case Complex128:
		return 128
	default:
		return 0
	}
}

func (k Kind) Max() interface{} {
	switch k {
	case Int:
		return int(^uint(0) >> 1)
	case Int8:
		return math.MaxInt8
	case Int16:
		return math.MaxInt16
	case Int32:
		return math.MaxInt32
	case Int64:
		return math.MaxInt64
	case Uint:
		return ^uint(0)
	case Uint8:
		return math.MaxUint8
	case Uint16:
		return math.MaxUint16
	case Uint32:
		return math.MaxUint32
	case Uint64:
		return uint64(math.MaxUint64)
	// for all the rest numeric type, probably,
	// it's unresonable to check for overflows anyway.
	default:
		return nil
	}
}

func (k Kind) Min() interface{} {
	switch k {
	case Int:
		return -int(^uint(0)>>1) - 1
	case Int8:
		return math.MinInt8
	case Int16:
		return math.MinInt16
	case Int32:
		return math.MinInt32
	case Int64:
		return math.MinInt64
	case Uint, Uint8, Uint16, Uint32, Uint64:
		return 0
	// For all the rest numeric type, probably,
	// it's unresonable to check for overflows anyway.
	default:
		return nil
	}
}

type Typ interface {
	Kind() Kind
	Type() string
}

type TPrimitive struct {
	TKind Kind
}

func (t TPrimitive) Kind() Kind {
	return t.TKind
}

func (t TPrimitive) Type() string {
	return t.TKind.Type()
}

type TArray struct {
	ETyp Typ
	Size int64
}

func (TArray) Kind() Kind {
	return Array
}

func (t TArray) Type() string {
	return fmt.Sprintf("[%d]%s", t.Size, t.ETyp.Type())
}

type TSlice struct {
	ETyp Typ
}

func (TSlice) Kind() Kind {
	return Slice
}

func (t TSlice) Type() string {
	return fmt.Sprintf("[]%s", t.ETyp.Type())
}

type TMap struct {
	KTyp, VTyp Typ
}

func (TMap) Kind() Kind {
	return Map
}

func (t TMap) Type() string {
	return fmt.Sprintf("map[%s]%s", t.KTyp.Type(), t.VTyp.Type())
}

type TPtr struct {
	ETyp Typ
}

func (TPtr) Kind() Kind {
	return Ptr
}

func (t TPtr) Type() string {
	return fmt.Sprintf("*%s", t.ETyp.Type())
}
