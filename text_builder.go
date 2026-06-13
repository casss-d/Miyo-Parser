package miyo
import (
	"strings"
	"unicode"
)

func buildText(tokens []*Token) string {
	if len(tokens) == 0 {
		return ""
	}

	var b strings.Builder
	for i, token := range tokens {
		if token.Value == "" {
			continue
		}
		prev := previousRealToken(tokens, i)
		next := nextRealToken(tokens, i)

		switch token.Kind {
		case TokenDelimiter:
			writeDelimiter(&b, token, prev, next)
		case TokenContextDelimiter:
			writeContextDelimiter(&b, token, prev, next)
		case TokenOpenBracket:
			trimTrailingSpace(&b)
			if b.Len() > 0 && prev != nil && prev.Kind == TokenDelimiter {
				b.WriteByte(' ')
			} else if b.Len() > 0 && prev != nil && prev.Kind != TokenOpenBracket && prev.End != token.Start {
				b.WriteByte(' ')
			}
			b.WriteString(token.Value)
		case TokenCloseBracket:
			trimTrailingSpace(&b)
			b.WriteString(token.Value)
			if next != nil && next.Kind == TokenText && token.End != next.Start {
				b.WriteByte(' ')
			}
		default:
			if b.Len() > 0 && prev != nil && shouldInsertSpace(prev, token) {
				b.WriteByte(' ')
			}
			b.WriteString(token.Value)
		}
	}
	return cleanupText(b.String())
}

func buildReleaseGroupText(tokens []*Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var b strings.Builder
	for i, token := range tokens {
		if token.Value == "" || token.isBracket() {
			continue
		}
		prev := previousRealToken(tokens, i)
		next := nextRealToken(tokens, i)
		switch token.Kind {
		case TokenDelimiter:
			switch token.Value {
			case "_", " ", "\t", "\n", "\r":
				writeSingleSpace(&b)
			case ".":
				trimTrailingSpace(&b)
				b.WriteByte('.')
			default:
				writeSingleSpace(&b)
			}
		case TokenContextDelimiter:
			if prev == nil || next == nil || (prev.Kind != TokenDelimiter && next.Kind != TokenDelimiter && prev.End == token.Start && token.End == next.Start) {
				trimTrailingSpace(&b)
				b.WriteString(token.Value)
			} else {
				writeSingleSpace(&b)
				trimTrailingSpace(&b)
				b.WriteByte(' ')
				b.WriteString(token.Value)
				b.WriteByte(' ')
			}
		default:
			if b.Len() > 0 && prev != nil && shouldInsertSpace(prev, token) {
				b.WriteByte(' ')
			}
			b.WriteString(token.Value)
		}
	}
	return cleanupText(b.String())
}

func previousRealToken(tokens []*Token, index int) *Token {
	for i := index - 1; i >= 0; i-- {
		if tokens[i].Value != "" {
			return tokens[i]
		}
	}
	return nil
}

func nextRealToken(tokens []*Token, index int) *Token {
	for i := index + 1; i < len(tokens); i++ {
		if tokens[i].Value != "" {
			return tokens[i]
		}
	}
	return nil
}

func shouldInsertSpace(prev *Token, current *Token) bool {
	if prev.Kind == TokenOpenBracket {
		return false
	}
	if current.Kind == TokenCloseBracket {
		return false
	}
	if prev.End == current.Start {
		return false
	}
	return true
}

func writeDelimiter(b *strings.Builder, token *Token, prev *Token, next *Token) {
	switch token.Value {
	case "_", " ", "\t", "\n", "\r":
		writeSingleSpace(b)
	case ".":
		if dotSeparatesWords(prev, next) {
			writeSingleSpace(b)
		} else if prev != nil && next != nil && prev.End == token.Start && token.End == next.Start {
			b.WriteByte('.')
		} else if next == nil {
			trimTrailingSpace(b)
			b.WriteByte('.')
		} else {
			writeSingleSpace(b)
		}
	case ",":
		trimTrailingSpace(b)
		b.WriteString(", ")
	case "。":
		trimTrailingSpace(b)
		b.WriteString("。")
	case "、":
		trimTrailingSpace(b)
		b.WriteString("、")
	default:
		writeSingleSpace(b)
	}
}

func dotSeparatesWords(prev *Token, next *Token) bool {
	if prev == nil || next == nil {
		return false
	}
	prevRunes := []rune(prev.Value)
	nextRunes := []rune(next.Value)
	if len(prevRunes) == 0 || len(nextRunes) == 0 {
		return false
	}
	if len(prevRunes) == 1 && !allDigits(prevRunes) {
		return unicode.IsLower(prevRunes[0]) && wordLike(nextRunes) && len(nextRunes) > 1
	}
	if normalizeWord(prev.Value) == "NO" && allDigits(nextRunes) {
		return false
	}
	if wordLike(prevRunes) && allDigits(nextRunes) {
		return true
	}
	if allDigits(prevRunes) && wordLike(nextRunes) {
		return true
	}
	if !wordLike(prevRunes) || !wordLike(nextRunes) {
		return false
	}
	return len(nextRunes) > 1 || unicode.IsLower(nextRunes[0])
}

func wordLike(value []rune) bool {
	hasLetter := false
	for _, r := range value {
		if unicode.IsLetter(r) {
			hasLetter = true
			continue
		}
		switch r {
		case '\'', '`', '\u2019':
			continue
		default:
			return false
		}
	}
	return hasLetter
}

func allDigits(value []rune) bool {
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(value) > 0
}

func writeContextDelimiter(b *strings.Builder, token *Token, prev *Token, next *Token) {
	if token.Value == "~" && next == nil && prev != nil && prev.End == token.Start {
		b.WriteString(token.Value)
		return
	}
	if prev != nil && next != nil && prev.End == token.Start && token.End == next.Start {
		b.WriteString(token.Value)
		return
	}
	writeSingleSpace(b)
	trimTrailingSpace(b)
	b.WriteByte(' ')
	b.WriteString(token.Value)
	b.WriteByte(' ')
}

func writeSingleSpace(b *strings.Builder) {
	if b.Len() == 0 {
		return
	}
	value := b.String()
	if strings.HasSuffix(value, " ") {
		return
	}
	b.WriteByte(' ')
}

func trimTrailingSpace(b *strings.Builder) {
	value := b.String()
	trimmed := strings.TrimRight(value, " ")
	if len(trimmed) == len(value) {
		return
	}
	b.Reset()
	b.WriteString(trimmed)
}

func cleanupText(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	value = strings.ReplaceAll(value, " .", ".")
	value = strings.ReplaceAll(value, " ,", ",")
	value = strings.ReplaceAll(value, "( ", "(")
	value = strings.ReplaceAll(value, " )", ")")
	value = strings.ReplaceAll(value, "[ ", "[")
	value = strings.ReplaceAll(value, " ]", "]")
	value = strings.ReplaceAll(value, "{ ", "{")
	value = strings.ReplaceAll(value, " }", "}")
	value = strings.ReplaceAll(value, " -  - ", " - ")
	return strings.TrimSpace(value)
}
