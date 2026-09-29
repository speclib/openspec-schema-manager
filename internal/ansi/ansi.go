// Package ansi strips terminal escape sequences from drawn output.
//
// It exists so that a test can assert on what a user reads rather than on how
// it was styled. Matching styled output ties every assertion to the colour
// scheme, and a change of emphasis then breaks tests that care about nothing of
// the sort.
package ansi

import "strings"

func Strip(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		if s[i] != 0x1b {
			b.WriteByte(s[i])
			i++
			continue
		}

		i++
		if i >= len(s) {
			break
		}

		switch s[i] {
		case '[':
			i++
			for i < len(s) && (s[i] < 0x40 || s[i] > 0x7e) {
				i++
			}
			if i < len(s) {
				i++
			}
		case ']':
			i++
			for i < len(s) {
				if s[i] == 0x07 {
					i++
					break
				}
				if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '\\' {
					i += 2
					break
				}
				i++
			}
		default:
			for i < len(s) && s[i] >= 0x20 && s[i] <= 0x2f {
				i++
			}
			if i < len(s) {
				i++
			}
		}
	}

	return b.String()
}
