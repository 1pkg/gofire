package gofire

import (
	"fmt"
	"strings"
	"unicode"
)

// #include

var tokenize = func(tokens []string) (args []string, flags map[string]string, err error) {
	var flname = func(token string) (string, error) {
		fln := strings.Replace(token, "--", "", 1)
		for _, r := range fln {
			if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
				return "", fmt.Errorf("flag name %s is not alphanumeric and can't be tokenized", fln)
			}
		}
		if _, ok := flags[fln]; ok {
			return "", fmt.Errorf("flag name %s used multiple times and can't be tokenized", fln)
		}
		return fln, nil
	}
	ltkns := len(tokens)
	args, flags = make([]string, 0, ltkns), make(map[string]string, ltkns)
	var fln, prevToken string
	for i := 0; i < ltkns; i++ {
		token := strings.TrimSpace(tokens[i])
		if token == "" {
			continue
		}
		iflag, iflagPrev := strings.HasPrefix(token, "--"), strings.HasPrefix(prevToken, "--")
		switch {
		case !iflag && !iflagPrev:
			args = append(args, token)
			prevToken = ""
		case iflag && strings.Contains(token, "="):
			parts := strings.SplitN(token, "=", 2)
			fln, err = flname(parts[0])
			flags[fln] = parts[1]
			prevToken = ""
		case !iflag && iflagPrev:
			fln, err = flname(prevToken)
			flags[fln] = token
			prevToken = ""
		case iflag && i != ltkns-1:
			prevToken = token
		default:
			err = fmt.Errorf("provided cli parameters %v can't be tokenized near token %s %s", tokens, prevToken, token)
		}
		if err != nil {
			args, flags = nil, nil
			return
		}
	}
	return
}
