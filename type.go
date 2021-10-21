package gofire

import (
	"fmt"
	"strings"
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
	Struct
	// Kinds bellow are not parsed and not processed by generators
	// but still defined here for visibility and potentially could be
	// processed in the future.
	UnsafePointer
	Chan
	Func
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

func (k Kind) Default() interface{} {
	switch k {
	case Bool:
		return false
	case Int, Int8, Int16, Int32, Int64:
		return int64(0)
	case Uint, Uint8, Uint16, Uint32, Uint64:
		return uint64(0)
	case Float32, Float64:
		return float64(0.0)
	case Complex64, Complex128:
		return complex128(0.0)
	case String:
		return ""
	case Array, Slice:
		return []interface{}{}
	case Map:
		return map[interface{}]interface{}{}
	default:
		return nil
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

type Typ interface {
	Kind() Kind
	Type() string
	Format(interface{}) string
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

func (t TPrimitive) Format(v interface{}) string {
	switch t.TKind {
	case Bool:
		return fmt.Sprintf("%t", v)
	case Int, Int8, Int16, Int32, Int64:
		return fmt.Sprintf("%d", v)
	case Uint, Uint8, Uint16, Uint32, Uint64:
		return fmt.Sprintf("%d", v)
	case Float32, Float64:
		return fmt.Sprintf("%f", v)
	case String:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%v", v)
	}
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

func (t TArray) Format(v interface{}) string {
	vs := v.([]interface{})
	fmts := make([]string, 0, len(vs))
	for _, v := range v.([]interface{}) {
		fmts = append(fmts, t.ETyp.Format(v))
	}
	return fmt.Sprintf("%s{%s}", t.Type(), strings.Join(fmts, ","))
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

func (t TSlice) Format(v interface{}) string {
	vs := v.([]interface{})
	fmts := make([]string, 0, len(vs))
	for _, v := range v.([]interface{}) {
		fmts = append(fmts, t.ETyp.Format(v))
	}
	return fmt.Sprintf("%s{%s}", t.Type(), strings.Join(fmts, ","))
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

func (t TMap) Format(v interface{}) string {
	vs := v.(map[interface{}]interface{})
	fmts := make([]string, 0, len(vs))
	for k, v := range v.([]interface{}) {
		kv := fmt.Sprintf("%s:%s", t.KTyp.Format(k), t.VTyp.Format(v))
		fmts = append(fmts, kv)
	}
	return fmt.Sprintf("%s{%s}", t.Type(), strings.Join(fmts, ","))
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

func (t TPtr) Format(v interface{}) string {
	return "nil"
}

type TStruct struct {
	Typ string
}

func (TStruct) Kind() Kind {
	return Struct
}

func (t TStruct) Type() string {
	return t.Typ
}

func (t TStruct) Format(v interface{}) string {
	return fmt.Sprintf("%s{}", t.Typ)
}
