package miyo

import (
	"strings"
	"unicode"
)

func normalizeInput(input string) string {
	var b strings.Builder
	for _, r := range input {
		b.WriteRune(normalizeFullWidthRune(r))
	}
	return strings.TrimSpace(b.String())
}

func normalizeFullWidthRune(r rune) rune {
	if r >= 0xFF01 && r <= 0xFF5E {
		return r - 0xFEE0
	}
	return r
}

func normalizeKey(s string) string {
	var b strings.Builder
	for _, r := range s {
		r = normalizeFullWidthRune(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '@' {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

func normalizeWord(s string) string {
	var b strings.Builder
	for _, r := range s {
		r = normalizeFullWidthRune(r)
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}
