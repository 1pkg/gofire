package intermediate

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
	Interface
)

func (k Kind) String() string {
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
	case Interface:
		return "interface"
	default:
		return "invalid"
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
