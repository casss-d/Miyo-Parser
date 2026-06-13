package miyo

type DebugResult struct {
	Tokens  []DebugToken `json:"tokens"`
	Matches []DebugMatch `json:"matches"`
}

type DebugToken struct {
	Value         string                   `json:"value"`
	Category      Tag                      `json:"category"`
	Start         int                      `json:"start"`
	End           int                      `json:"end"`
	GroupID       int                      `json:"group_id"`
	Possibilities map[Tag]DebugPossibility `json:"possibilities,omitempty"`
}

type DebugPossibility struct {
	Score      float64 `json:"score"`
	Descriptor Tag     `json:"descriptor,omitempty"`
	Value      string  `json:"value,omitempty"`
	Source     string  `json:"source,omitempty"`
}

type DebugMatch struct {
	Tag   Tag    `json:"tag"`
	Value string `json:"value"`
	From  int    `json:"from"`
	To    int    `json:"to"`
}

func newDebugResult(tokens []*Token, matches []termMatch) *DebugResult {
	result := &DebugResult{
		Tokens:  make([]DebugToken, 0, len(tokens)),
		Matches: make([]DebugMatch, 0, len(matches)),
	}
	for _, token := range tokens {
		possibilities := make(map[Tag]DebugPossibility)
		for _, possibility := range token.Possibilities {
			var descriptor Tag
			if possibility.Descriptor != possibility.Tag {
				descriptor = possibility.Descriptor
			}
			possibilities[possibility.Tag] = DebugPossibility{
				Score:      possibility.Score,
				Descriptor: descriptor,
				Value:      possibility.Value,
				Source:     possibility.Source,
			}
		}
		result.Tokens = append(result.Tokens, DebugToken{
			Value:         token.Value,
			Category:      token.Category,
			Start:         token.Start,
			End:           token.End,
			GroupID:       token.GroupID,
			Possibilities: possibilities,
		})
	}
	for _, match := range matches {
		result.Matches = append(result.Matches, DebugMatch{
			Tag:   match.Tag,
			Value: match.Value,
			From:  match.TokenFrom,
			To:    match.TokenTo,
		})
	}
	return result
}
