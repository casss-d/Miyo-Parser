package miyo

import "strings"

type LexiconEntry struct {
	Key       string
	Tag       Tag
	Canonical string
	Score     float64
	Ambiguous bool
}

type termMatch struct {
	Tag       Tag
	Value     string
	TokenFrom int
	TokenTo   int
	Score     float64
	Source    string
}

var defaultLexicon = buildLexicon()
var defaultLexiconPrefixes = buildLexiconPrefixes(defaultLexicon)

func buildLexicon() map[string]LexiconEntry {
	entries := make(map[string]LexiconEntry)
	for _, entry := range lexiconEntries {
		key := normalizeKey(entry.Key)
		if key == "" {
			continue
		}
		if entry.Canonical == "" {
			entry.Canonical = entry.Key
		}
		if entry.Score == 0 {
			entry.Score = 3
		}
		if _, exists := entries[key]; exists {
			continue
		}
		entries[key] = entry
	}
	return entries
}

func (p *parser) scanLexicon() {
	for i := 0; i < len(p.tokens); i++ {
		if !p.tokens[i].isTextLike() {
			continue
		}
		best, ok := p.longestLexiconMatch(i)
		if !ok {
			continue
		}
		p.addMatch(best.Tag, best.Value, best.TokenFrom, best.TokenTo, best.Score, "lexicon")
		i = best.TokenTo
	}
}

func (p *parser) longestLexiconMatch(start int) (termMatch, bool) {
	var best termMatch
	found := false
	var key string
	words := 0
	last := start

	for i := start; i < len(p.tokens) && words < 5; i++ {
		token := p.tokens[i]
		if token.isBracket() {
			break
		}
		if token.isTextLike() {
			words++
		}
		key += token.Value
		normalized := normalizeKey(key)
		if normalized == "" {
			continue
		}
		if entry, ok := p.lexicon[normalized]; ok && token.isTextLike() {
			if !p.lexiconEntryAppliesToTokens(entry, start, i) {
				continue
			}
			best = termMatch{
				Tag:       entry.Tag,
				Value:     entry.Canonical,
				TokenFrom: start,
				TokenTo:   i,
				Score:     entry.Score,
				Source:    "lexicon",
			}
			found = true
			last = i
		}
		if words >= 1 && !hasLexiconPrefix(normalized) && i > last {
			break
		}
	}

	return best, found
}

func (p *parser) lexiconEntryAppliesToTokens(entry LexiconEntry, start int, end int) bool {
	if entryIsJoinedAnimeType(entry, start, end) {
		return false
	}
	if p.entryIsLowercaseUseOfShortAbbreviation(entry, start, end) {
		return false
	}
	if entry.Ambiguous {
		if p.contextualSupportScore(start, end) <= 0 {
			return false
		}
	}
	return true
}

func entryIsJoinedAnimeType(entry LexiconEntry, start int, end int) bool {
	if start == end || entry.Tag != TagAnimeType {
		return false
	}
	switch normalizeKey(entry.Key) {
	case "NCOP", "NCED", "OP", "ED", "OVA", "OVAS", "OAV", "ONA", "OAD", "SP":
		return true
	default:
		return false
	}
}

func (p *parser) entryIsLowercaseUseOfShortAbbreviation(entry LexiconEntry, start int, end int) bool {
	if !entry.Ambiguous || start != end {
		return false
	}
	key := normalizeKey(entry.Key)
	if len(key) == 0 || len(key) > 3 {
		return false
	}
	if strings.ToUpper(entry.Key) != entry.Key {
		return false
	}
	if p.tokens[start].GroupID >= 0 && p.groupLooksLikeShortAbbreviationCluster(p.tokens[start].GroupID) {
		return false
	}
	value := p.tokens[start].Value
	return strings.ToUpper(value) != value
}

func (p *parser) groupLooksLikeShortAbbreviationCluster(groupID int) bool {
	textTokens := 0
	for _, token := range p.tokens {
		if token.GroupID != groupID {
			continue
		}
		if token.isDelimiter() || token.isContextDelimiter() || token.isBracket() {
			continue
		}
		if !token.isTextLike() {
			return false
		}
		key := normalizeKey(token.Value)
		if key == "" {
			continue
		}
		if len(key) > 3 {
			return false
		}
		textTokens++
	}
	return textTokens > 0
}


func buildLexiconPrefixes(lexicon map[string]LexiconEntry) map[string]struct{} {
	prefixes := make(map[string]struct{})
	for key := range lexicon {
		for i := 1; i < len(key); i++ {
			prefixes[key[:i]] = struct{}{}
		}
	}
	return prefixes
}

func hasLexiconPrefix(prefix string) bool {
	if prefix == "" {
		return true
	}
	_, ok := defaultLexiconPrefixes[prefix]
	return ok
}

