package gofire

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	table := map[string]struct {
		tokens []string
		args   []string
		flags  map[string]string
		err    error
	}{
		"valid args only tokens should return expected result": {
			tokens: []string{"100", "aaaa", `"zzz"`, "20.20", "az-bd", "true"},
			args:   []string{"100", "aaaa", `"zzz"`, "20.20", "az-bd", "true"},
			flags:  map[string]string{},
		},
		"valid flags only tokens should return expected result": {
			tokens: []string{"--fl", "100", "--dd=aaaa", "--a", `"zzz"`, "--fff", "20.20", "--az", "az-bd", "--t=true"},
			args:   []string{},
			flags: map[string]string{
				"fl":  "100",
				"dd":  "aaaa",
				"a":   `"zzz"`,
				"fff": "20.20",
				"az":  "az-bd",
				"t":   "true",
			},
		},
		"mixture of valid params tokens should return expected result": {
			tokens: []string{"100", "--dd=aaaa", `"zzz"`, "--fff", "20.20", "--az", "az-bd", "true"},
			args:   []string{"100", `"zzz"`, "true"},
			flags: map[string]string{
				"dd":  "aaaa",
				"fff": "20.20",
				"az":  "az-bd",
			},
		},
		"mixture of valid params tokens at the beggining should return expected result": {
			tokens: []string{"100", `"zzz"`, "true", "--dd=aaaa", "--fff", "20.20", "--az", "az-bd"},
			args:   []string{"100", `"zzz"`, "true"},
			flags: map[string]string{
				"dd":  "aaaa",
				"fff": "20.20",
				"az":  "az-bd",
			},
		},
		"mixture of valid params tokens at the end should return expected result": {
			tokens: []string{"--dd=aaaa", "--fff", "20.20", "--az", "az-bd", "100", `"zzz"`, "true"},
			args:   []string{"100", `"zzz"`, "true"},
			flags: map[string]string{
				"dd":  "aaaa",
				"fff": "20.20",
				"az":  "az-bd",
			},
		},
		"mixture of valid complex params tokens should return expected result": {
			tokens: []string{"{100, 220, -10,  } ", `--dd={aaaa:"bbbb--"}`, `"zzz"`, "--fff", "    ", " {test:{}, t:{1,2} } ", " ", "true"},
			args:   []string{"{100, 220, -10,  }", `"zzz"`, "true"},
			flags: map[string]string{
				"dd":  `{aaaa:"bbbb--"}`,
				"fff": "{test:{}, t:{1,2} }",
			},
		},
		"implicit bool flags tokens should produce tokenizer error": {
			tokens: []string{"--fl", "100", "--dd=aaaa", "--az", "az-bd", "--t"},
			err:    errors.New("provided cli parameters [--fl 100 --dd=aaaa --az az-bd --t] can't be tokenized near token  --t"),
		},
		"non alphanumeric flags tokens should produce tokenizer error": {
			tokens: []string{"--fl", "100", "--dd++=aaaa", "--az", "az-bd", "--t=1"},
			err:    errors.New("flag name dd++ is not alphanumeric and can't be tokenized"),
		},
		"duplicated flags tokens should produce tokenizer error": {
			tokens: []string{"--fl", "100", "--dd=aaaa", "--dd", "az-bd", "--t=1"},
			err:    errors.New("flag name dd used multiple times and can't be tokenized"),
		},
	}
	for tname, tcase := range table {
		t.Run(tname, func(t *testing.T) {
			args, flags, err := tokenize(tcase.tokens)
			if fmt.Sprintf("%v", tcase.err) != fmt.Sprintf("%v", err) {
				t.Fatalf("expected error message %q but got %q", tcase.err, err)
			}
			if !reflect.DeepEqual(tcase.args, args) {
				t.Fatalf("expected args list %v but got %v", tcase.args, args)
			}
			if !reflect.DeepEqual(tcase.flags, flags) {
				t.Fatalf("expected args map %v but got %v", tcase.flags, flags)
			}
		})
	}
}
