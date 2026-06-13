package miyo
type BracketGroup struct {
	ID        int
	Open      int
	Close     int
	OpenMark  string
	CloseMark string
	IsFirst   bool
	IsLast    bool
	Starts    bool
}

func analyzeBrackets(tokens []*Token) []BracketGroup {
	type stackEntry struct {
		index int
		value string
	}

	stack := make([]stackEntry, 0)
	groups := make([]BracketGroup, 0)

	for i, token := range tokens {
		if token.Kind == TokenOpenBracket {
			stack = append(stack, stackEntry{index: i, value: token.Value})
			continue
		}
		if token.Kind != TokenCloseBracket || len(stack) == 0 {
			continue
		}
		last := stack[len(stack)-1]
		if matchingClose(last.value) != token.Value {
			continue
		}
		stack = stack[:len(stack)-1]
		group := BracketGroup{
			ID:        len(groups),
			Open:      last.index,
			Close:     i,
			OpenMark:  last.value,
			CloseMark: token.Value,
			Starts:    last.index == 0,
		}
		for j := last.index + 1; j < i; j++ {
			tokens[j].GroupID = group.ID
		}
		groups = append(groups, group)
	}

	if len(groups) > 0 {
		groups[0].IsFirst = true
		groups[len(groups)-1].IsLast = true
	}

	return groups
}

func tokensInGroup(tokens []*Token, group BracketGroup) []*Token {
	if group.Open+1 >= group.Close {
		return nil
	}
	return tokens[group.Open+1 : group.Close]
}
