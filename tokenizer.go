package miyo
import "regexp"

var splitTokenRegexp = regexp.MustCompile(`(?i)^(.+[^A-Za-z0-9])([SE]\d{1,2}|v\d|\d{1,4})$`)

func tokenize(input string) []*Token {
	input = normalizeInput(input)
	tokens := make([]*Token, 0)
	var buffer []rune
	bufferStart := 0
	bytePos := 0

	flushText := func() {
		if len(buffer) == 0 {
			return
		}
		tokens = append(tokens, newToken(string(buffer), bufferStart, TokenText))
		buffer = buffer[:0]
	}

	for _, r := range input {
		value := string(r)
		switch {
		case isOpeningBracket(value):
			flushText()
			tokens = append(tokens, newToken(value, bytePos, TokenOpenBracket))
		case isClosingBracket(value):
			flushText()
			tokens = append(tokens, newToken(value, bytePos, TokenCloseBracket))
		case isContextDelimiterRune(r):
			flushText()
			tokens = append(tokens, newToken(value, bytePos, TokenContextDelimiter))
		case isPlainDelimiter(r):
			flushText()
			tokens = append(tokens, newToken(value, bytePos, TokenDelimiter))
		default:
			if len(buffer) == 0 {
				bufferStart = bytePos
			}
			buffer = append(buffer, r)
		}
		bytePos += len(value)
	}
	flushText()

	splitTokens := make([]*Token, 0, len(tokens))
	for _, token := range tokens {
		splitTokens = append(splitTokens, splitTokenIfNeeded(token)...)
	}
	tokens = splitTokens

	for _, token := range tokens {
		switch token.Kind {
		case TokenDelimiter:
			token.addPossibility(TagDelimiter, 10, token.Value, "tokenizer")
			token.Category = TagDelimiter
		case TokenContextDelimiter:
			token.addPossibility(TagContextDelimiter, 10, token.Value, "tokenizer")
			token.Category = TagContextDelimiter
		case TokenOpenBracket, TokenCloseBracket:
			token.addPossibility(TagBracket, 10, token.Value, "tokenizer")
			token.Category = TagBracket
		}
	}

	return tokens
}

func splitTokenIfNeeded(t *Token) []*Token {
	if t.Kind != TokenText {
		return []*Token{t}
	}
	matches := splitTokenRegexp.FindStringSubmatch(t.Value)
	if matches == nil {
		return []*Token{t}
	}
	mainPart := matches[1]
	suffixPart := matches[2]

	mainToken := newToken(mainPart, t.Start, TokenText)
	suffixToken := newToken(suffixPart, t.Start+len(mainPart), TokenText)

	result := splitTokenIfNeeded(mainToken)
	result = append(result, suffixToken)
	return result
}
