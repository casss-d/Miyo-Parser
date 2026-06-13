package miyo
import "strings"

func (p *parser) detectTitle() {
	if p.leadingEpisodeHasEpisodeTitle() {
		return
	}
	start := 0
	if len(p.groups) > 0 && p.groups[0].Starts && p.releaseGroup != "" {
		start = p.leadingReleaseGroupTitleStart()
	}
	if prefixTitleStart := p.titleStartAfterLeadingEpisode(); prefixTitleStart != -1 {
		start = prefixTitleStart
	}

	p.titleTokens = p.collectTitleTokens(start)
	for _, token := range p.titleTokens {
		token.addPossibility(TagTitle, 3.2, token.Value, "title")
	}
}

func (p *parser) leadingReleaseGroupTitleStart() int {
	groups := p.leadingReleaseGroups()
	if len(groups) == 0 {
		return p.groups[0].Close + 1
	}
	return groups[len(groups)-1].Close + 1
}

func (p *parser) leadingEpisodeHasEpisodeTitle() bool {
	for _, match := range p.matches {
		if match.Tag != TagEpisode {
			continue
		}
		prev := p.prevMeaningful(match.TokenFrom)
		if prev == -1 {
			next := p.nextMeaningful(match.TokenTo)
			return next != -1 && p.tokens[next].isContextDelimiter()
		}
		if prev > 2 {
			return false
		}
		if possibility, ok := p.tokens[prev].possibility(TagSequencePrefix); !ok || possibility.effectiveDescriptor() != TagEpisode {
			return false
		}
		next := p.nextMeaningful(match.TokenTo)
		return next != -1 && p.tokens[next].isContextDelimiter()
	}
	return false
}

func (p *parser) titleStartAfterLeadingEpisode() int {
	for _, match := range p.matches {
		if match.Tag != TagEpisode {
			continue
		}
		prev := p.prevMeaningful(match.TokenFrom)
		if prev == -1 {
			continue
		}
		possibility, ok := p.tokens[prev].possibility(TagSequencePrefix)
		if !ok || possibility.effectiveDescriptor() != TagEpisode {
			continue
		}
		if prev <= 2 {
			return match.TokenTo + 1
		}
	}
	return -1
}

func (p *parser) collectTitleTokens(start int) []*Token {
	candidates := make([]*Token, 0)
	groupByOpen := p.groupsByOpen()

	for i := start; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if token.Kind == TokenOpenBracket {
			group, ok := groupByOpen[i]
			if ok {
				if len(candidates) == 0 {
					if titlePrefix := p.groupTitleTokensBeforeMetadata(group); len(titlePrefix) > 0 {
						candidates = append(candidates, titlePrefix...)
						i = group.Close
						continue
					}
				}
				if p.shouldIncludeBracketGroupInTitle(group, len(candidates) > 0) {
					if len(candidates) == 0 && (group.Starts || p.groupLooksLikeBracketTitleCandidate(group)) {
						candidates = append(candidates, p.tokens[group.Open+1:group.Close]...)
					} else {
						candidates = append(candidates, p.tokens[group.Open:group.Close+1]...)
					}
					i = group.Close
					continue
				}
				if len(candidates) > 0 {
					break
				}
				i = group.Close
				continue
			}
		}
		if token.Kind == TokenDelimiter {
			if len(candidates) > 0 {
				next := p.nextMeaningful(i)
				if token.Value == "." && next != -1 && p.tokenIsTitleBoundaryAt(next) && !p.dotIsTitleInternal(i) {
					break
				}
				candidates = append(candidates, token)
			}
			continue
		}
		if token.isContextDelimiter() {
			if len(candidates) == 0 {
				continue
			}
			if token.Value == "|" {
				break
			}
			next := p.nextMeaningful(i)
			if next != -1 && p.tokenIsTitleBoundaryAt(next) {
				break
			}
			candidates = append(candidates, token)
			continue
		}
		if p.tokenIsTitleBoundaryAt(i) {
			if nextTitleIndex, addColon, ok := p.titleContinuationAfterOrdinalSeason(i, len(candidates) > 0); ok {
				if addColon {
					candidates = trimTrailingTitleDelimiters(candidates)
					if len(candidates) > 0 {
						candidates = append(candidates, newToken(":", candidates[len(candidates)-1].End, TokenText))
					}
				}
				i = nextTitleIndex - 1
				continue
			}
			if len(candidates) > 0 {
				if p.tokenLooksLikeSeasonMarker(i) && p.hasSlashBeforeBoundary(i) {
					candidates = append(candidates, p.tokens[i])
					continue
				}
				break
			}
			continue
		}
		candidates = append(candidates, token)
	}

	return p.trimTitleTokens(candidates)
}

func (p *parser) shouldIncludeBracketGroupInTitle(group BracketGroup, hasTitleBefore bool) bool {
	if !hasTitleBefore && p.groupLooksLikeLeadingBracketTitle(group) {
		return true
	}
	if p.groupHasMetadata(group) {
		return false
	}
	if hasTitleBefore && groupLooksLikeQuotedTitleSuffix(group) {
		return true
	}
	if !hasTitleBefore {
		return p.groupLooksLikeBracketTitleCandidate(group)
	}
	next := p.nextMeaningful(group.Close)
	if next == -1 {
		return false
	}
	return !p.tokenIsTitleBoundaryAt(next)
}

func groupLooksLikeQuotedTitleSuffix(group BracketGroup) bool {
	return group.OpenMark == "「" || group.OpenMark == "『"
}

func (p *parser) groupLooksLikeBracketTitleCandidate(group BracketGroup) bool {
	hasTitleWord := false
	for i := group.Open + 1; i < group.Close; i++ {
		token := p.tokens[i]
		if token.isDelimiter() || token.isContextDelimiter() || token.isBracket() {
			continue
		}
		if !token.isTextLike() || tokenHasHardMetadata(token) {
			return false
		}
		key := normalizeKey(token.Value)
		if len(key) > 4 || containsNonASCII(token.Value) {
			hasTitleWord = true
		}
	}
	return hasTitleWord
}

func (p *parser) groupTitleTokensBeforeMetadata(group BracketGroup) []*Token {
	candidates := make([]*Token, 0)
	sawMetadata := false
	for i := group.Open + 1; i < group.Close; i++ {
		token := p.tokens[i]
		if token.isBracket() {
			continue
		}
		if tokenHasHardMetadata(token) {
			sawMetadata = true
			break
		}
		candidates = append(candidates, token)
	}
	if !sawMetadata {
		return nil
	}
	candidates = p.trimTitleTokens(candidates)
	if !tokensLookLikeTitle(candidates) {
		return nil
	}
	return candidates
}

func tokensLookLikeTitle(tokens []*Token) bool {
	for _, token := range tokens {
		if token.isDelimiter() || token.isContextDelimiter() {
			continue
		}
		if token.isTextLike() && (len(normalizeKey(token.Value)) > 1 || containsNonASCII(token.Value)) {
			return true
		}
	}
	return false
}

func (p *parser) groupHasMetadata(group BracketGroup) bool {
	for i := group.Open + 1; i < group.Close; i++ {
		token := p.tokens[i]
		if token.isDelimiter() || token.isContextDelimiter() {
			continue
		}
		if tokenHasHardMetadata(token) {
			return true
		}
	}
	return false
}

func (p *parser) tokenIsTitleBoundary(token *Token) bool {
	for i, candidate := range p.tokens {
		if candidate == token {
			return p.tokenIsTitleBoundaryAt(i)
		}
	}
	return p.tokenIsTitleBoundaryValue(token, false)
}

func (p *parser) tokenIsTitleBoundaryAt(index int) bool {
	if index < 0 || index >= len(p.tokens) {
		return false
	}
	token := p.tokens[index]
	return p.tokenIsTitleBoundaryValue(token, p.languageTokenHasMediaContext(index))
}

func (p *parser) hasSlashBeforeBoundary(index int) bool {
	// Look ahead from index+1 to find a slash before a structural
	// delimiter (dash followed by episode number). If found, the
	// current marker is embedded in a bilingual title.
	for j := index + 1; j < len(p.tokens); j++ {
		tok := p.tokens[j]
		if tok.Kind == TokenDelimiter && tok.Value == "-" {
			return false
		}
		if tok.isContextDelimiter() && tok.Value == "|" {
			return false
		}
		if tok.isContextDelimiter() && tok.Value == "/" {
			return true
		}
	}
	return false
}

func (p *parser) tokenLooksLikeSeasonMarker(index int) bool {
	if index < 0 || index >= len(p.tokens) || !p.tokens[index].isTextLike() {
		return false
	}
	return seasonRegexp.MatchString(p.tokens[index].Value)
}

func (p *parser) tokenIsTitleBoundaryValue(token *Token, languageHasContext bool) bool {
	if token == nil {
		return false
	}
	if isStandaloneTechnicalModeMarker(token) {
		return true
	}
	if token.Category.isHardMetadata() {
		if token.Category == TagLanguage && !languageHasContext {
			return false
		}
		return true
	}
	for _, possibility := range token.Possibilities {
		if possibility.Tag == TagSequencePrefix {
			if possibility.effectiveDescriptor().isHardMetadata() {
				return true
			}
			continue
		}
		if possibility.Tag == TagLanguage && !languageHasContext {
			continue
		}
		if possibility.Tag.isHardMetadata() {
			return true
		}
	}
	return false
}

func isStandaloneTechnicalModeMarker(token *Token) bool {
	if token == nil || !token.isTextLike() {
		return false
	}
	switch normalizeWord(token.Value) {
	case "MULTI", "DUAL":
		return true
	default:
		return false
	}
}

func (p *parser) titleContinuationAfterOrdinalSeason(index int, hasTitleBefore bool) (int, bool, bool) {
	if !hasTitleBefore {
		return -1, false, false
	}
	if _, ok := ordinalNumber(p.tokens[index].Value); !ok {
		return -1, false, false
	}
	seasonIndex := p.nextMeaningful(index)
	if seasonIndex == -1 || !isSeasonPrefixKey(normalizeKey(p.tokens[seasonIndex].Value)) {
		return -1, false, false
	}
	addColon := strings.Contains(p.tokens[seasonIndex].Value, ":")
	for i := seasonIndex + 1; i < len(p.tokens); i++ {
		if p.tokens[i].Kind == TokenDelimiter {
			continue
		}
		if p.tokenIsTitleBoundaryAt(i) {
			return -1, false, false
		}
		return i, addColon, addColon
	}
	return -1, false, false
}

func containsNonASCII(value string) bool {
	for _, r := range value {
		if r > 127 {
			return true
		}
	}
	return false
}

func (p *parser) trimTitleTokens(tokens []*Token) []*Token {
	for len(tokens) > 0 && isTrimBoundaryToken(tokens[0]) {
		tokens = tokens[1:]
	}
	for len(tokens) > 0 && isTrimBoundaryToken(tokens[len(tokens)-1]) {
		if tokens[len(tokens)-1].Value == "~" && containsEarlierTitleDelimiter(tokens[:len(tokens)-1], "~") {
			break
		}
		tokens = tokens[:len(tokens)-1]
	}
	if p.shouldTrimTrailingDot(tokens) {
		tokens = tokens[:len(tokens)-1]
		for len(tokens) > 0 && isTrimBoundaryToken(tokens[len(tokens)-1]) {
			if tokens[len(tokens)-1].Value == "~" && containsEarlierTitleDelimiter(tokens[:len(tokens)-1], "~") {
				break
			}
			tokens = tokens[:len(tokens)-1]
		}
	}
	return tokens
}

func (p *parser) isDotDelimited() bool {
	return !strings.Contains(p.baseName, " ") && !strings.Contains(p.baseName, "_") && strings.Count(p.baseName, ".") >= 2
}

func (p *parser) shouldTrimTrailingDot(tokens []*Token) bool {
	if !p.isDotDelimited() {
		return false
	}
	if len(tokens) == 0 {
		return false
	}
	last := tokens[len(tokens)-1]
	if last.Kind != TokenDelimiter || last.Value != "." {
		return false
	}
	prev := previousRealToken(tokens, len(tokens)-1)
	if prev != nil {
		prevValue := normalizeWord(prev.Value)
		if prevValue == "NO" || prevValue == "VER" || len(prevValue) == 1 {
			return false
		}
	}
	return true
}

func containsEarlierTitleDelimiter(tokens []*Token, value string) bool {
	for _, token := range tokens {
		if token.isContextDelimiter() && token.Value == value {
			return true
		}
	}
	return false
}

func trimTrailingTitleDelimiters(tokens []*Token) []*Token {
	for len(tokens) > 0 && tokens[len(tokens)-1].Kind == TokenDelimiter {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens
}

func (p *parser) trimEpisodeTitleTokens(tokens []*Token) []*Token {
	tokens = p.trimTitleTokens(tokens)
	for len(tokens) > 0 && tokens[len(tokens)-1].Kind == TokenDelimiter && tokens[len(tokens)-1].Value == "." {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens
}

func (p *parser) dotIsTitleInternal(dotIndex int) bool {
	prev := p.prevMeaningful(dotIndex)
	next := p.nextMeaningful(dotIndex)
	if prev == -1 || next == -1 {
		return false
	}
	prevValue := normalizeWord(p.tokens[prev].Value)
	if prevValue == "NO" || prevValue == "VER1" {
		return true
	}
	return false
}

func isTrimBoundaryToken(token *Token) bool {
	if token.isContextDelimiter() {
		return true
	}
	if token.Kind != TokenDelimiter {
		return false
	}
	switch token.Value {
	case " ", "_", "\t", "\n", "\r", ",":
		return true
	default:
		return false
	}
}

func (p *parser) detectEpisodeTitle() {
	lastEpisode := p.lastMainEpisodeMatch()
	if lastEpisode == nil {
		return
	}
	if lastEpisode.Source == "episode-total" {
		return
	}
	if lastEpisode.Source == "bracket-episode-range" {
		return
	}
	start := lastEpisode.TokenTo + 1
	if groupID := p.tokens[lastEpisode.TokenTo].GroupID; groupID >= 0 && groupID < len(p.groups) {
		start = p.groups[groupID].Close + 1
	}
	for start < len(p.tokens) && p.tokens[start].Kind == TokenDelimiter {
		start++
	}
	if start >= len(p.tokens) {
		return
	}
	groupByOpen := p.groupsByOpen()
	skippedMetadata := false
	for start < len(p.tokens) && p.tokens[start].Kind == TokenOpenBracket {
		if p.detectQuotedEpisodeTitle(start) {
			return
		}
		if group, ok := groupByOpen[start]; ok && p.groupContainsOnlyMetadata(group) {
			start = group.Close + 1
			skippedMetadata = true
			for start < len(p.tokens) && (p.tokens[start].Kind == TokenDelimiter || p.tokens[start].isContextDelimiter()) {
				if p.tokens[start].isContextDelimiter() && p.tokens[start].Value == "|" {
					break
				}
				start++
			}
		} else {
			break
		}
	}
	if start >= len(p.tokens) {
		return
	}
	if p.tokens[start].Kind == TokenOpenBracket {
		return
	}
	if p.tokens[start].Kind == TokenCloseBracket {
		return
	}
	if p.tokens[start].isContextDelimiter() {
		if p.tokens[start].Value == "|" && skippedMetadata {
			return
		}
		start++
		for start < len(p.tokens) && p.tokens[start].Kind == TokenDelimiter {
			start++
		}
	} else if p.tokenIsTitleBoundaryAt(start) {
		return
	}

	title := make([]*Token, 0)
	for i := start; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if token.Kind == TokenOpenBracket {
			if group, ok := groupByOpen[i]; ok {
				if len(title) > 0 {
					break
				}
				i = group.Close
				continue
			}
		}
		if token.Kind == TokenDelimiter {
			if len(title) > 0 {
				title = append(title, token)
			}
			continue
		}
		if token.isContextDelimiter() {
			if token.Value == "|" {
				break
			}
			if len(title) > 0 {
				title = append(title, token)
			}
			continue
		}
		if p.tokenIsTitleBoundaryAt(i) {
			break
		}
		title = append(title, token)
	}
	title = p.trimEpisodeTitleTokens(title)
	p.episodeTitle = title
	for _, token := range title {
		token.addPossibility(TagEpisodeTitle, 3, token.Value, "episode-title")
	}
}

func (p *parser) detectQuotedEpisodeTitle(openIndex int) bool {
	groupByOpen := p.groupsByOpen()
	group, ok := groupByOpen[openIndex]
	if !ok {
		return false
	}
	switch group.OpenMark {
	case "「", "『":
	default:
		return false
	}
	title := p.trimTitleTokens(p.tokens[group.Open+1 : group.Close])
	if len(title) == 0 {
		return false
	}
	p.episodeTitle = title
	for _, token := range title {
		token.addPossibility(TagEpisodeTitle, 3, token.Value, "episode-title")
	}
	return true
}

func (p *parser) groupContainsOnlyMetadata(group BracketGroup) bool {
	for i := group.Open + 1; i < group.Close; i++ {
		tok := p.tokens[i]
		if !tok.isTextLike() {
			continue
		}
		if tok.Category == TagUnknown || tok.Category == TagTitle || tok.Category == TagEpisodeTitle {
			return false
		}
	}
	return true
}

func (p *parser) lastMainEpisodeMatch() *termMatch {
	var last *termMatch
	for i := range p.matches {
		if p.matches[i].Tag != TagEpisode {
			continue
		}
		last = &p.matches[i]
	}
	return last
}
