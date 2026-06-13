package miyo
type TokenKind uint8

const (
	TokenText TokenKind = iota
	TokenDelimiter
	TokenContextDelimiter
	TokenOpenBracket
	TokenCloseBracket
)

type Token struct {
	Value         string
	Start         int
	End           int
	Kind          TokenKind
	GroupID       int
	Category      Tag
	Possibilities []Possibility
}

func newToken(value string, start int, kind TokenKind) *Token {
	return &Token{
		Value:    value,
		Start:    start,
		End:      start + len(value),
		Kind:     kind,
		GroupID:  -1,
		Category: TagUnknown,
	}
}

func (t *Token) addPossibility(tag Tag, score float64, value string, source string) {
	t.addPossibilityWithDescriptor(tag, tag, score, value, source)
}

func (t *Token) addPossibilityWithDescriptor(tag Tag, descriptor Tag, score float64, value string, source string) {
	possibility := Possibility{
		Tag:        tag,
		Descriptor: descriptor,
		Score:      score,
		Value:      value,
		Source:     source,
	}
	for i, existing := range t.Possibilities {
		if existing.Tag != tag {
			continue
		}
		if existing.Score > score {
			return
		}
		t.Possibilities[i] = possibility
		return
	}
	t.Possibilities = append(t.Possibilities, possibility)
}

func (t *Token) resolvedValue() string {
	if possibility, ok := t.possibility(t.Category); ok && possibility.Value != "" {
		return possibility.Value
	}
	return t.Value
}

func (t *Token) possibility(tag Tag) (Possibility, bool) {
	for _, possibility := range t.Possibilities {
		if possibility.Tag == tag {
			return possibility, true
		}
	}
	return Possibility{}, false
}

func (t *Token) isBracket() bool {
	return t.Kind == TokenOpenBracket || t.Kind == TokenCloseBracket
}

func (t *Token) isDelimiter() bool {
	return t.Kind == TokenDelimiter || t.Kind == TokenContextDelimiter
}

func (t *Token) isContextDelimiter() bool {
	return t.Kind == TokenContextDelimiter
}

func (t *Token) isTextLike() bool {
	return t.Kind == TokenText
}
