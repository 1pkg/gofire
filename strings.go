package gofire

import "strings"

// Splitb splits the provided string tokens based on balanced char sequences.
func Splitb(s string, by string, bcs ...string) []string {
	tokens := strings.Split(s, by)
	var accum string
	result := make([]string, 0, len(tokens))
loop:
	for _, t := range tokens {
		accum += t
		var count, prev int = 0, -1
		for _, c := range bcs {
			cnt := strings.Count(accum, c)
			if prev != -1 && prev != cnt {
				accum += by
				continue loop
			}
			prev = cnt
			count += cnt
		}
		if count%2 != 0 {
			accum += by
			continue loop
		}
		result = append(result, strings.TrimSpace(accum))
		accum = ""
	}
	return result
}
