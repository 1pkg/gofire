package internal

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/1pkg/gofire"
)

func TestParseTypeValue(t *testing.T) {
	table := map[string]struct {
		typ gofire.Typ
		val string
		out interface{}
		err error
	}{
		"not supported type should not parse anything": {
			typ: gofire.TPrimitive{TKind: gofire.Chan},
		},
		"string type value should be always parsed as string": {
			typ: gofire.TPrimitive{TKind: gofire.String},
			val: "value",
			out: "value",
		},
		"string type value should be always parsed as string quoted": {
			typ: gofire.TPrimitive{TKind: gofire.String},
			val: `"value"`,
			out: "value",
		},
		"string type value should be always parsed as string double quoted": {
			typ: gofire.TPrimitive{TKind: gofire.String},
			val: `"\"value\""`,
			out: `"value"`,
		},
		"bool type true value should be parsed as a bool value": {
			typ: gofire.TPrimitive{TKind: gofire.Bool},
			val: "true",
			out: true,
		},
		"bool type false value should be parsed as a bool value": {
			typ: gofire.TPrimitive{TKind: gofire.Bool},
			val: "false",
			out: false,
		},
		"bool type empty value should be parsed as a bool value": {
			typ: gofire.TPrimitive{TKind: gofire.Bool},
			out: false,
		},
		"bool type string value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Bool},
			val: "value",
			out: false,
			err: errors.New(`strconv.ParseBool: parsing "value": invalid syntax`),
		},
		"int32 type int32 value should be parsed as an int64 value": {
			typ: gofire.TPrimitive{TKind: gofire.Int32},
			val: "42",
			out: int64(42),
		},
		"int32 type negative int32 value should be parsed as an int64 value": {
			typ: gofire.TPrimitive{TKind: gofire.Int32},
			val: "-42",
			out: int64(-42),
		},
		"int32 type empty value should be parsed as an int64 value": {
			typ: gofire.TPrimitive{TKind: gofire.Int32},
			out: int64(0),
		},
		"int32 type string value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Int32},
			val: "value",
			out: int64(0),
			err: errors.New(`strconv.ParseInt: parsing "value": invalid syntax`),
		},
		"int32 type int64 value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Int32},
			val: fmt.Sprint(math.MaxInt64),
			out: int64(math.MaxInt32),
			err: errors.New(`strconv.ParseInt: parsing "9223372036854775807": value out of range`),
		},
		"uint32 type uint32 value should be parsed as an uint64 value": {
			typ: gofire.TPrimitive{TKind: gofire.Uint32},
			val: "42",
			out: uint64(42),
		},
		"uint32 type negative int32 value should  fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Uint32},
			val: "-42",
			out: uint64(0),
			err: errors.New(`strconv.ParseUint: parsing "-42": invalid syntax`),
		},
		"uint32 type empty value should be parsed as an uint64 value": {
			typ: gofire.TPrimitive{TKind: gofire.Uint32},
			out: uint64(0),
		},
		"uint32 type string value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Uint32},
			val: "value",
			out: uint64(0),
			err: errors.New(`strconv.ParseUint: parsing "value": invalid syntax`),
		},
		"uint32 type int64 value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Uint32},
			val: fmt.Sprint(uint64(math.MaxUint64)),
			out: uint64(math.MaxUint32),
			err: errors.New(`strconv.ParseUint: parsing "18446744073709551615": value out of range`),
		},
		"float32 type float value should be parsed as a float64 value": {
			typ: gofire.TPrimitive{TKind: gofire.Float32},
			val: "-12.125",
			out: float64(-12.125),
		},
		"float32 type empty value should be parsed as a float64 value": {
			typ: gofire.TPrimitive{TKind: gofire.Float32},
			out: float64(0),
		},
		"float32 type float64 value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Float32},
			val: fmt.Sprint(float64(math.MaxFloat64)),
			out: float64(math.Inf(1)),
			err: errors.New(`strconv.ParseFloat: parsing "1.7976931348623157e+308": value out of range`),
		},
		"float32 type string value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Float32},
			val: "value",
			out: float64(0),
			err: errors.New(`strconv.ParseFloat: parsing "value": invalid syntax`),
		},
		"complex64 type complex value should be parsed as a complex128 value": {
			typ: gofire.TPrimitive{TKind: gofire.Complex64},
			val: "(10+10i)",
			out: complex(10, 10),
		},
		"complex64 type empty value should be parsed as a complex128 value": {
			typ: gofire.TPrimitive{TKind: gofire.Complex64},
			out: complex(0, 0),
		},
		"complex64 type complex128 value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Complex64},
			val: fmt.Sprint(complex(math.MaxFloat64, 0)),
			out: complex(math.Inf(1), 0),
			err: errors.New(`strconv.ParseComplex: parsing "(1.7976931348623157e+308+0i)": value out of range`),
		},
		"complex64 type string value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Complex64},
			val: "value",
			out: complex(0, 0),
			err: errors.New(`strconv.ParseComplex: parsing "value": invalid syntax`),
		},
		"string slice type slice value should be parsed as a slice value": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "{value_1, value_2 , value_3}",
			out: []interface{}{"value_1", "value_2", "value_3"},
		},
		"string slice empty value should be parsed as an empty slice value": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "",
			out: []interface{}{},
		},
		"string slice empty slice value should be parsed as an empty slice value": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "{}",
			out: []interface{}{},
		},
		"string slice empty slice with trailing spaces value should be parsed as an empty slice value": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "{   }",
			out: []interface{}{},
		},
		"string slice type escaped slice value should be parsed as a slice value": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: `{ "value 1", "value 2", " value_3 " }`,
			out: []interface{}{"value 1", "value 2", " value_3 "},
		},
		"string slice type mixed slice value should be parsed as a slice value": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: `{ "value 1", value 2, " value_3 " }`,
			out: []interface{}{"value 1", "value 2", " value_3 "},
		},
		"string slice type brackets value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "[value_1, value_2, value_3]",
			err: errors.New(`invalid value "[value_1, value_2, value_3]" can't be parsed as an array or a slice`),
		},
		"string slice type unformatted value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "value_1, value_2, value_3",
			err: errors.New(`invalid value "value_1, value_2, value_3" can't be parsed as an array or a slice`),
		},
		"float slice type float slice value should be parsed as a slice value": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Float32}},
			val: "{1.0, 10.5, -100.0}",
			out: []interface{}{1.0, 10.5, -100.0},
		},
		"float slice empty value should be parsed as an empty slice value": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Float32}},
			val: "",
			out: []interface{}{},
		},
		"float slice empty slice value should be parsed as empty slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Float32}},
			val: "{}",
			out: []interface{}{},
		},
		"float slice type mixed string slice value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Float32}},
			val: `{ 10.250, val }`,
			err: errors.New(`strconv.ParseFloat: parsing "val": invalid syntax`),
		},
		"int array type int array value should be parsed as a slice value": {
			typ: gofire.TArray{ETyp: gofire.TPrimitive{TKind: gofire.Int32}, Size: 3},
			val: "{10, 10, -10}",
			out: []interface{}{int64(10), int64(10), int64(-10)},
		},
		"int array empty value should be parsed as an empty slice value": {
			typ: gofire.TArray{ETyp: gofire.TPrimitive{TKind: gofire.Int32}, Size: 0},
			val: "",
			out: []interface{}{},
		},
		"int array empty slice value should be parsed as an empty slice value": {
			typ: gofire.TArray{ETyp: gofire.TPrimitive{TKind: gofire.Int32}, Size: 0},
			val: "{}",
			out: []interface{}{},
		},
		"int array type mixed slice value should fail on parse": {
			typ: gofire.TArray{ETyp: gofire.TPrimitive{TKind: gofire.Int32}, Size: 2},
			val: `{ 10, val }`,
			err: errors.New(`strconv.ParseInt: parsing "val": invalid syntax`),
		},
		"int array type bigger int array value should fail on parse": {
			typ: gofire.TArray{ETyp: gofire.TPrimitive{TKind: gofire.Int32}, Size: 3},
			val: "{10, 10, -10, -10}",
			err: errors.New(`invalid value "{10, 10, -10, -10}" can't be parsed as an array 3`),
		},
		"map string:uint type with valid value should be parsed as a map value": {
			typ: gofire.TMap{KTyp: gofire.TPrimitive{TKind: gofire.String}, VTyp: gofire.TPrimitive{TKind: gofire.Uint}},
			val: "{aaa:10, bbb:100, ccc:1000}",
			out: map[interface{}]interface{}{"aaa": uint64(10), "bbb": uint64(100), "ccc": uint64(1000)},
		},
		"map string:uint type empty value should be parsed as an empty map value": {
			typ: gofire.TMap{KTyp: gofire.TPrimitive{TKind: gofire.String}, VTyp: gofire.TPrimitive{TKind: gofire.Uint}},
			val: "",
			out: map[interface{}]interface{}{},
		},
		"map string:uint type empty map value should be parsed as an empty map value": {
			typ: gofire.TMap{KTyp: gofire.TPrimitive{TKind: gofire.String}, VTyp: gofire.TPrimitive{TKind: gofire.Uint}},
			val: "{   }",
			out: map[interface{}]interface{}{},
		},
		"map string:uint type not formated map value should fail on parse": {
			typ: gofire.TMap{KTyp: gofire.TPrimitive{TKind: gofire.String}, VTyp: gofire.TPrimitive{TKind: gofire.Uint}},
			val: `{ val:10, test:100`,
			err: errors.New(`invalid value "{ val:10, test:100" can't be parsed as a map`),
		},
		"map string:uint type not formated pair value should fail on parse": {
			typ: gofire.TMap{KTyp: gofire.TPrimitive{TKind: gofire.String}, VTyp: gofire.TPrimitive{TKind: gofire.Uint}},
			val: `{ val:10, test+100 }`,
			err: errors.New(`invalid value "{ val:10, test+100 }" can't be parsed as a map`),
		},
		"map string:uint type mixed map value should fail on parse": {
			typ: gofire.TMap{KTyp: gofire.TPrimitive{TKind: gofire.String}, VTyp: gofire.TPrimitive{TKind: gofire.Uint}},
			val: `{ val:10, test:val }`,
			err: errors.New(`strconv.ParseUint: parsing "val": invalid syntax`),
		},
		"map uint:string type mixed map value should fail on parse": {
			typ: gofire.TMap{KTyp: gofire.TPrimitive{TKind: gofire.Uint}, VTyp: gofire.TPrimitive{TKind: gofire.String}},
			val: `{ 100:10, test:aaa }`,
			err: errors.New(`strconv.ParseUint: parsing "test": invalid syntax`),
		},
		"complex nested type with valid value should be parsed properly": {
			typ: gofire.TMap{KTyp: gofire.TPrimitive{TKind: gofire.String}, VTyp: gofire.TSlice{ETyp: gofire.TArray{ETyp: gofire.TPrimitive{TKind: gofire.Int}, Size: 3}}},
			val: `{ a:{{1,2,3}}, "c d":{}, test:{{0,0,0}, {1,-1,1,} , }, }`,
			out: map[interface{}]interface{}{
				"a":    []interface{}{[]interface{}{int64(1), int64(2), int64(3)}},
				"c d":  []interface{}{},
				"test": []interface{}{[]interface{}{int64(0), int64(0), int(0)}, []interface{}{int64(1), int64(-1), int64(1)}},
			},
		},
	}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			out, err := ParseTypeValue(tcase.typ, tcase.val)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q", tcase.err, err)
			}
			if !reflect.DeepEqual(tcase.out, out) {
				t.Fatalf("expected out %#v but got %#v", tcase.out, out)
			}
		})
	}
}
