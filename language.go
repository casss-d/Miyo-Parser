package miyo

import (
	"regexp"
	"strings"
)

func (p *parser) scanEmbeddedLanguages() {
	for i, token := range p.tokens {
		if !token.isTextLike() {
			continue
		}
		value := strings.ToUpper(token.Value)
		for _, term := range embeddedLanguageTerms {
			if containsWord(value, term.Pattern) {
				p.addMatch(TagLanguage, term.Canonical, i, i, 2.7, "embedded-language")
			}
		}
	}
}

func containsWord(haystack string, needle string) bool {
	idx := strings.Index(haystack, needle)
	if idx == -1 {
		return false
	}
	end := idx + len(needle)
	if idx > 0 && isAlphaNumeric(haystack[idx-1]) {
		return false
	}
	if end < len(haystack) && isAlphaNumeric(haystack[end]) {
		return false
	}
	return true
}

func isAlphaNumeric(r byte) bool {
	return (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func (p *parser) scanEmbeddedAudioTerms() {
	p.scanEmbeddedTechnicalTerms(TagAudioTerm, embeddedAudioTerms, "embedded-audio")
}

func (p *parser) scanStandaloneDualAudioTerms() {
	if p.hasMatch(TagAudioTerm, "Dual Audio") {
		return
	}
	for i, token := range p.tokens {
		if !token.isTextLike() || normalizeKey(token.Value) != "DUAL" {
			continue
		}
		if token.Value != strings.ToUpper(token.Value) || p.dualTokenHasSubtitleContext(i) {
			continue
		}
		p.addMatch(TagAudioTerm, "Dual Audio", i, i, 3.1, "standalone-dual-audio")
	}
}

func (p *parser) hasMatch(tag Tag, value string) bool {
	for _, token := range p.tokens {
		if poss, ok := token.possibility(tag); ok && poss.Value == value {
			return true
		}
	}
	return false
}

func (p *parser) dualTokenHasSubtitleContext(index int) bool {
	next := p.nextMeaningful(index)
	if next == -1 {
		return false
	}
	if p.tokens[index].GroupID >= 0 && p.tokens[next].GroupID != p.tokens[index].GroupID {
		return false
	}
	switch normalizeKey(p.tokens[next].Value) {
	case "SUB", "SUBS", "SUBTITLE", "SUBTITLES":
		return true
	default:
		return tokenHasTag(p.tokens[next], TagSubtitles)
	}
}

func (p *parser) scanEmbeddedVideoTerms() {
	p.scanEmbeddedTechnicalTerms(TagVideoTerm, embeddedVideoTerms, "embedded-video")
}

func (p *parser) scanEmbeddedSubtitleTerms() {
	p.scanEmbeddedTechnicalTerms(TagSubtitles, embeddedSubtitleTerms, "embedded-subtitles")
}

func (p *parser) languageTokenHasMediaContext(index int) bool {
	if p.languageTokenLooksStandaloneMetadata(index) {
		return true
	}
	if p.languageTokenIsInLanguageCluster(index) {
		return true
	}
	if p.languageTokenHasSubtitleContext(index) || p.languageTokenHasAudioContext(index) {
		return true
	}
	return p.contextualSupportScore(index, index) > 0
}

func (p *parser) languageTokenLooksStandaloneMetadata(index int) bool {
	if index < 0 || index >= len(p.tokens) || !tokenHasTag(p.tokens[index], TagLanguage) {
		return false
	}
	switch normalizeKey(p.tokens[index].Value) {
	case "CHS", "CHT", "ENG", "ESP", "FR", "ITA", "JAP", "JP", "JPN", "PTBR", "RU", "RUS", "SUBFRENCH", "VF", "VOSTFR", "RAW":
		return true
	default:
		return false
	}
}

func (p *parser) languageTokenHasSubtitleContext(index int) bool {
	for _, neighbor := range p.languageContextNeighbors(index) {
		if tokenHasTag(p.tokens[neighbor], TagSubtitles) && !tokenIsDubSubtitleTerm(p.tokens[neighbor]) {
			return true
		}
		switch normalizeKey(p.tokens[neighbor].Value) {
		case "SUB", "SUBS", "SUBTITLE", "SUBTITLES":
			return true
		}
	}
	return false
}

func (p *parser) languageTokenHasAudioContext(index int) bool {
	for _, neighbor := range p.languageContextNeighbors(index) {
		if tokenHasTag(p.tokens[neighbor], TagAudioTerm) || tokenIsDubSubtitleTerm(p.tokens[neighbor]) {
			return true
		}
		switch normalizeKey(p.tokens[neighbor].Value) {
		case "AUDIO", "DUB", "DUBBED":
			return true
		}
	}
	return false
}

func (p *parser) languageContextNeighbors(index int) []int {
	neighbors := make([]int, 0, 2)
	for _, neighbor := range []int{p.prevMeaningful(index), p.nextMeaningful(index)} {
		if neighbor == -1 {
			continue
		}
		if p.tokens[index].GroupID >= 0 && p.tokens[neighbor].GroupID != p.tokens[index].GroupID {
			continue
		}
		neighbors = append(neighbors, neighbor)
	}
	return neighbors
}

func (p *parser) languageTokenIsInLanguageCluster(index int) bool {
	groupID := p.tokens[index].GroupID
	if groupID < 0 || groupID >= len(p.groups) {
		return false
	}
	languages := 0
	for i := p.groups[groupID].Open + 1; i < p.groups[groupID].Close; i++ {
		token := p.tokens[i]
		if token.isDelimiter() || token.isContextDelimiter() || token.isBracket() {
			continue
		}
		if !token.isTextLike() || !tokenHasTag(token, TagLanguage) {
			return false
		}
		languages++
	}
	return languages > 0
}

func tokenIsDubSubtitleTerm(token *Token) bool {
	if token == nil {
		return false
	}
	if !tokenHasTag(token, TagSubtitles) {
		return false
	}
	switch normalizeKey(token.Value) {
	case "DUB", "DUBBED":
		return true
	default:
		return false
	}
}

func (p *parser) scanEmbeddedTechnicalTerms(tag Tag, terms []embeddedTechnicalTerm, source string) {
	input := normalizeInput(p.baseName)
	for _, term := range terms {
		for _, loc := range term.Pattern.FindAllStringSubmatchIndex(input, -1) {
			if len(loc) < 4 || loc[2] < 0 {
				continue
			}
			from, to, ok := p.tokenRangeForByteSpan(loc[2], loc[3])
			if !ok {
				continue
			}
			p.addMatch(tag, term.Canonical, from, to, term.Score, source)
		}
	}
}

func (p *parser) tokenRangeForByteSpan(start int, end int) (int, int, bool) {
	from := -1
	to := -1
	for i, token := range p.tokens {
		if token.End <= start {
			continue
		}
		if token.Start >= end {
			break
		}
		if from == -1 {
			from = i
		}
		to = i
	}
	if from == -1 || to == -1 {
		return 0, 0, false
	}
	return from, to, true
}

type embeddedLanguageTerm struct {
	Pattern   string
	Canonical string
}

var embeddedLanguageTerms = []embeddedLanguageTerm{
	{Pattern: "SUBFRENCH", Canonical: "SUBFRENCH"},
	{Pattern: "VOSTFR", Canonical: "VOSTFR"},
	{Pattern: "ENGLISH", Canonical: "English"},
	{Pattern: "FRE", Canonical: "Fre"},
	{Pattern: "CHT", Canonical: "CHT"},
	{Pattern: "CHS", Canonical: "CHS"},
	{Pattern: "JAP", Canonical: "Jap"},
}

type embeddedTechnicalTerm struct {
	Pattern   *regexp.Regexp
	Canonical string
	Score     float64
}

func embeddedTermPattern(value string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)(?:^|[^A-Z0-9])(` + value + `)(?:$|[^A-Z0-9])`)
}

var embeddedAudioTerms = []embeddedTechnicalTerm{
	{Pattern: embeddedTermPattern(`DTS-HD(?:\s*MA)?`), Canonical: "DTS-HD", Score: 3.2},
	{Pattern: embeddedTermPattern(`E-?AC-?3`), Canonical: "E-AC-3", Score: 3.2},
	{Pattern: embeddedTermPattern(`TRUEHD`), Canonical: "TrueHD", Score: 3.2},
	{Pattern: embeddedTermPattern(`FLAC(?:\d\.\d)?`), Canonical: "FLAC", Score: 3.2},
	{Pattern: embeddedTermPattern(`AAC(?:\d\.\d)?`), Canonical: "AAC", Score: 3.2},
	{Pattern: embeddedTermPattern(`DDP(?:\d\.\d)?`), Canonical: "DDP", Score: 3.2},
	{Pattern: embeddedTermPattern(`AC3`), Canonical: "AC3", Score: 3.2},
	{Pattern: embeddedTermPattern(`OPUS(?:\d\.\d)?`), Canonical: "Opus", Score: 3.0},
	{Pattern: embeddedTermPattern(`PCM`), Canonical: "PCM", Score: 3.2},
	{Pattern: embeddedTermPattern(`ATMOS`), Canonical: "Atmos", Score: 3.0},
}

var embeddedVideoTerms = []embeddedTechnicalTerm{
	{Pattern: embeddedTermPattern(`HDR10\+`), Canonical: "HDR10+", Score: 3.2},
	{Pattern: embeddedTermPattern(`DV`), Canonical: "DV", Score: 3.2},
	{Pattern: embeddedTermPattern(`DOLBY\s*VISION`), Canonical: "Dolby Vision", Score: 3.2},
}

var embeddedSubtitleTerms = []embeddedTechnicalTerm{
	{Pattern: embeddedTermPattern(`MULTI-SUBTITLES`), Canonical: "Multi-Subtitles", Score: 3.2},
	{Pattern: embeddedTermPattern(`MULTI-SUBTITLE`), Canonical: "Multi-Subtitle", Score: 3.2},
	{Pattern: embeddedTermPattern(`MULTI-SUBS`), Canonical: "Multi-Subs", Score: 3.2},
	{Pattern: embeddedTermPattern(`MULTI-SUB`), Canonical: "Multi-Sub", Score: 3.2},
}
