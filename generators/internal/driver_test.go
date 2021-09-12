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
		"string type value should be always parsed as string ": {
			typ: gofire.TPrimitive{TKind: gofire.String},
			val: "value",
			out: "value",
		},
		"bool type true value should be parsed as bool": {
			typ: gofire.TPrimitive{TKind: gofire.Bool},
			val: "true",
			out: true,
		},
		"bool type false value should be parsed as bool": {
			typ: gofire.TPrimitive{TKind: gofire.Bool},
			val: "false",
			out: false,
		},
		"bool type empty value should be parsed as bool": {
			typ: gofire.TPrimitive{TKind: gofire.Bool},
			out: false,
		},
		"bool type string value should fail on parse": {
			typ: gofire.TPrimitive{TKind: gofire.Bool},
			val: "value",
			out: false,
			err: errors.New(`strconv.ParseBool: parsing "value": invalid syntax`),
		},
		"int32 type int32 value should be parsed as int64 value": {
			typ: gofire.TPrimitive{TKind: gofire.Int32},
			val: "42",
			out: int64(42),
		},
		"int32 type negative int32 value should be parsed as int64 value": {
			typ: gofire.TPrimitive{TKind: gofire.Int32},
			val: "-42",
			out: int64(-42),
		},
		"int32 type empty value should be parsed as int64": {
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
		"uint32 type uint32 value should be parsed as uint64 value": {
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
		"uint32 type empty value should be parsed as uint64": {
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
		"float32 type float value should be parsed as float64": {
			typ: gofire.TPrimitive{TKind: gofire.Float32},
			val: "-12.125",
			out: float64(-12.125),
		},
		"float32 type empty value should be parsed as float64": {
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
		"complex64 type complex value should be parsed as complex128": {
			typ: gofire.TPrimitive{TKind: gofire.Complex64},
			val: "(10+10i)",
			out: complex(10, 10),
		},
		"complex64 type empty value should be parsed as complex128": {
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
		"string slice type slice value should be parsed as string slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "{value_1, value_2 , value_3}",
			out: []string{"value_1", "value_2", "value_3"},
		},
		"string slice empty value should be parsed as empty string slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "",
			out: []string{},
		},
		"string slice empty slice value should be parsed as empty string slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "{}",
			out: []string{},
		},
		"string slice empty slice with trailing spaces value should be parsed as empty string slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "{   }",
			out: []string{},
		},
		"string slice type escaped slice value should be parsed as string slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: `{ "value 1", "value 2", " value_3 " }`,
			out: []string{"value 1", "value 2", " value_3 "},
		},
		"string slice type mixed slice value should be parsed as string slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: `{ "value 1", value 2, " value_3 " }`,
			out: []string{"value 1", "value 2", " value_3 "},
		},
		"string slice type brackets value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "[value_1, value_2, value_3]",
			err: errors.New(`invalid value "[value_1, value_2, value_3]" can't be parsed as a slice`),
		},
		"string slice type unformatted value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.String}},
			val: "value_1, value_2, value_3",
			err: errors.New(`invalid value "value_1, value_2, value_3" can't be parsed as a slice`),
		},
		"bool slice type bool slice value should be parsed as bool slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Bool}},
			val: "{true, false, true, true}",
			out: []bool{true, false, true, true},
		},
		"bool slice empty value should be parsed as empty bool slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Bool}},
			val: "",
			out: []bool{},
		},
		"bool slice empty slice value should be parsed as empty bool slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Bool}},
			val: "{}",
			out: []bool{},
		},
		"bool slice type mixed string slice value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Bool}},
			val: `{ true, false, " true ", val }`,
			err: errors.New(`strconv.ParseBool: parsing "\" true \"": invalid syntax`),
		},
		"bool slice type brackets value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Bool}},
			val: "[]",
			err: errors.New(`invalid value "[]" can't be parsed as a slice`),
		},
		"int32 slice type int32 slice value should be parsed as int64 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Int32}},
			val: "{1, -1, 10}",
			out: []int64{1, -1, 10},
		},
		"int32 slice empty value should be parsed as empty int64 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Int32}},
			val: "",
			out: []int64{},
		},
		"int32 slice empty slice value should be parsed as empty int64 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Int32}},
			val: "{}",
			out: []int64{},
		},
		"int32 slice type mixed string slice value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Int32}},
			val: `{ 1, 10, val }`,
			err: errors.New(`strconv.ParseInt: parsing "val": invalid syntax`),
		},
		"int32 slice type brackets value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Int32}},
			val: "[]",
			err: errors.New(`invalid value "[]" can't be parsed as a slice`),
		},
		"uint32 slice type uint32 slice value should be parsed as uint64 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Uint32}},
			val: "{1, 10, 100}",
			out: []uint64{1, 10, 100},
		},
		"uint32 slice empty value should be parsed as empty uint64 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Uint32}},
			val: "",
			out: []uint64{},
		},
		"uint32 slice empty slice value should be parsed as empty uint64 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Uint32}},
			val: "{}",
			out: []uint64{},
		},
		"uint32 slice type mixed string slice value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Uint32}},
			val: `{ 1, 10, val }`,
			err: errors.New(`strconv.ParseUint: parsing "val": invalid syntax`),
		},
		"uint32 slice type brackets value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Uint32}},
			val: "[]",
			err: errors.New(`invalid value "[]" can't be parsed as a slice`),
		},
		"float32 slice type float32 slice value should be parsed as float64 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Float32}},
			val: "{1.0, 10.5, -100.0}",
			out: []float64{1.0, 10.5, -100.0},
		},
		"float32 slice empty value should be parsed as empty float64 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Float32}},
			val: "",
			out: []float64{},
		},
		"float32 slice empty slice value should be parsed as empty float64 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Float32}},
			val: "{}",
			out: []float64{},
		},
		"float32 slice type mixed string slice value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Float32}},
			val: `{ 10.250, val }`,
			err: errors.New(`strconv.ParseFloat: parsing "val": invalid syntax`),
		},
		"float32 slice type brackets value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Float32}},
			val: "[]",
			err: errors.New(`invalid value "[]" can't be parsed as a slice`),
		},
		"complex64 slice type uint32 slice value should be parsed as complex128 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Complex64}},
			val: "{(1.0+1.0i), (-10.0+10.0i)}",
			out: []complex128{(1.0 + 1.0i), (-10.0 + 10.0i)},
		},
		"complex64 slice empty value should be parsed as empty complex128 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Complex64}},
			val: "",
			out: []complex128{},
		},
		"complex64 slice empty slice value should be parsed as empty complex128 slice": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Complex64}},
			val: "{}",
			out: []complex128{},
		},
		"complex64 slice type mixed string slice value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Complex64}},
			val: `{ (-10.0+10.0i), val }`,
			err: errors.New(`strconv.ParseComplex: parsing "val": invalid syntax`),
		},
		"complex64 slice type brackets value should fail on parse": {
			typ: gofire.TSlice{ETyp: gofire.TPrimitive{TKind: gofire.Complex64}},
			val: "[]",
			err: errors.New(`invalid value "[]" can't be parsed as a slice`),
		},
	}
	d := BaseDriver{}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			out, err := d.ParseTypeValue(tcase.typ, tcase.val)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q", tcase.err, err)
			}
			if !reflect.DeepEqual(tcase.out, out) {
				t.Fatalf("expected out %#v but got %#v", tcase.out, out)
			}
		})
	}
}
