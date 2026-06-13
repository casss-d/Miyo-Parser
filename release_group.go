package miyo
import (
	"strings"
	"unicode"
)

func (p *parser) detectReleaseGroup() {
	if len(p.groups) == 0 {
		p.detectDashSuffixReleaseGroup()
		return
	}
	first := p.groups[0]
	if first.Starts && !p.groupLooksLikeLeadingBracketTitle(first) && p.groupLooksLikeReleaseGroup(first) {
		groups := p.leadingReleaseGroups()
		value := p.joinReleaseGroupText(groups)
		if value != "" {
			p.releaseGroup = value
			last := groups[len(groups)-1]
			p.addMatch(TagReleaseGroup, value, first.Open+1, last.Close-1, 4.4, "release-group")
			return
		}
	}
	p.detectEmbeddedBracketReleaseGroup()
	if p.releaseGroup == "" {
		p.detectTrailingReleaseGroup()
	}
	if p.releaseGroup == "" {
		p.detectDashSuffixReleaseGroup()
	}
}

func (p *parser) leadingReleaseGroups() []BracketGroup {
	if len(p.groups) == 0 {
		return nil
	}
	first := p.groups[0]
	if !first.Starts || p.groupLooksLikeLeadingBracketTitle(first) || !p.groupLooksLikeReleaseGroup(first) {
		return nil
	}
	groups := []BracketGroup{first}
	previous := first
	for i := 1; i < len(p.groups); i++ {
		group := p.groups[i]
		if !p.groupIsAdjacentAfter(previous, group) {
			break
		}
		if group.OpenMark != "[" || groupLooksLikeQuotedTitleSuffix(group) || p.groupHasMetadata(group) || p.groupLooksLikeYearRange(group) || !p.groupLooksLikeReleaseGroup(group) {
			break
		}
		groups = append(groups, group)
		previous = group
	}
	return groups
}

func (p *parser) groupIsAdjacentAfter(previous BracketGroup, next BracketGroup) bool {
	if next.Open <= previous.Close {
		return false
	}
	hasDelimiter := false
	for i := previous.Close + 1; i < next.Open; i++ {
		if p.tokens[i].Kind != TokenDelimiter {
			return false
		}
		hasDelimiter = true
	}
	return hasDelimiter
}

func (p *parser) joinReleaseGroupText(groups []BracketGroup) string {
	values := make([]string, 0, len(groups))
	for _, group := range groups {
		value := buildReleaseGroupText(tokensInGroup(p.tokens, group))
		if value != "" {
			values = append(values, value)
		}
	}
	return strings.Join(values, " & ")
}

func (p *parser) groupLooksLikeReleaseGroup(group BracketGroup) bool {
	if p.groupLooksLikeYearRange(group) {
		return false
	}
	inside := tokensInGroup(p.tokens, group)
	if len(inside) == 0 {
		return false
	}
	words := 0
	hardWords := 0
	for _, token := range inside {
		if token.isDelimiter() || token.isContextDelimiter() {
			continue
		}
		if tokenHasHardMetadata(token) {
			words++
			hardWords++
			continue
		}
		if token.isTextLike() {
			words++
		}
	}
	return words > 0 && words <= 5 && words != hardWords
}

func (p *parser) groupLooksLikeYearRange(group BracketGroup) bool {
	first := p.nextMeaningful(group.Open)
	if first == -1 || first >= group.Close || !p.tokens[first].isTextLike() {
		return false
	}
	delimiter := p.nextMeaningful(first)
	if delimiter == -1 || delimiter >= group.Close || !p.tokens[delimiter].isContextDelimiter() || !isRangeConnector(p.tokens[delimiter].Value) {
		return false
	}
	second := p.nextMeaningful(delimiter)
	if second == -1 || second >= group.Close || !p.tokens[second].isTextLike() {
		return false
	}
	after := p.nextMeaningful(second)
	return after == group.Close &&
		isYear(p.tokens[first].Value, p.options.YearMin, p.options.YearMax) &&
		isYear(p.tokens[second].Value, p.options.YearMin, p.options.YearMax)
}

func (p *parser) groupLooksLikeLeadingBracketTitle(group BracketGroup) bool {
	if !group.Starts || !p.groupLooksLikeReleaseGroup(group) {
		return false
	}
	next := p.nextMeaningful(group.Close)
	if next == -1 {
		return false
	}
	if _, _, ok := cleanNumber(p.tokens[next].Value); !ok {
		return false
	}
	return p.hasDotDelimitedMetadataAfter(next)
}

func (p *parser) detectTrailingReleaseGroup() {
	if len(p.groups) == 0 {
		return
	}
	group := p.groups[len(p.groups)-1]
	if group.OpenMark == "(" || group.Starts || groupLooksLikeQuotedTitleSuffix(group) || p.groupHasMetadata(group) || !p.groupLooksLikeReleaseGroup(group) {
		return
	}
	value := buildReleaseGroupText(tokensInGroup(p.tokens, group))
	if value == "" {
		return
	}
	p.releaseGroup = value
	p.addMatch(TagReleaseGroup, value, group.Open+1, group.Close-1, 3.4, "trailing-release-group")
}

func (p *parser) detectEmbeddedBracketReleaseGroup() {
	for _, group := range p.groups {
		if group.Starts || groupLooksLikeQuotedTitleSuffix(group) || p.groupLooksLikeYearRange(group) || !p.groupHasMetadata(group) {
			continue
		}
		for delimiterIndex := group.Close - 1; delimiterIndex > group.Open; delimiterIndex-- {
			token := p.tokens[delimiterIndex]
			if !token.isContextDelimiter() || token.Value != "-" {
				continue
			}
			start := p.nextMeaningful(delimiterIndex)
			if start == -1 || start >= group.Close {
				continue
			}
			if !p.trailingSegmentLooksLikeReleaseGroup(start, group.Close-1) {
				continue
			}
			value := buildReleaseGroupText(p.tokens[start:group.Close])
			if value == "" {
				continue
			}
			p.releaseGroup = value
			p.addMatch(TagReleaseGroup, value, start, group.Close-1, 3.6, "embedded-bracket-release-group")
			return
		}
	}
}

func (p *parser) detectDashSuffixReleaseGroup() {
	for delimiterIndex := len(p.tokens) - 1; delimiterIndex >= 0; delimiterIndex-- {
		token := p.tokens[delimiterIndex]
		if !token.isContextDelimiter() || token.Value != "-" {
			continue
		}
		before := p.prevMeaningful(delimiterIndex)
		if before == -1 || !p.tokenIsTechnicalReleaseBoundary(before) || !p.hasTechnicalMetadataBefore(delimiterIndex) {
			continue
		}
		start := p.nextMeaningful(delimiterIndex)
		if start == -1 {
			continue
		}
		end := p.releaseGroupSegmentEnd(start)
		// When the RG candidate consists entirely of non-ASCII characters
		// (e.g. CJK text), it is almost certainly a title or descriptor,
		// not a release group name.  (Legitimate dash-suffix RGs use Latin
		// letters, digits, or a mix thereof.)
		if p.trailingSegmentIsNonASCII(start, end) {
			continue
		}
		if end < start || !p.trailingSegmentLooksLikeReleaseGroup(start, end) {
			continue
		}
		value := buildReleaseGroupText(p.tokens[start : end+1])
		if value == "" {
			continue
		}
		p.releaseGroup = value
		p.addMatch(TagReleaseGroup, value, start, end, 3.2, "dash-suffix-release-group")
		return
	}
}

func (p *parser) tokenIsTechnicalReleaseBoundary(index int) bool {
	token := p.tokens[index]
	if token.Kind == TokenCloseBracket {
		return true
	}
	if isStandaloneTechnicalModeMarker(token) {
		return true
	}
	return tokenHasTag(token, TagVideoResolution) ||
		tokenHasTag(token, TagVideoTerm) ||
		tokenHasTag(token, TagAudioTerm) ||
		tokenHasTag(token, TagSource)
}

func (p *parser) hasTechnicalMetadataBefore(index int) bool {
	for i := 0; i < index; i++ {
		if tokenHasTag(p.tokens[i], TagVideoResolution) ||
			tokenHasTag(p.tokens[i], TagVideoTerm) ||
			tokenHasTag(p.tokens[i], TagAudioTerm) ||
			tokenHasTag(p.tokens[i], TagSource) {
			return true
		}
	}
	return false
}

func (p *parser) lastMeaningful() int {
	for i := len(p.tokens) - 1; i >= 0; i-- {
		if p.tokens[i].Kind != TokenDelimiter {
			return i
		}
	}
	return -1
}

func (p *parser) releaseGroupSegmentEnd(start int) int {
	end := p.lastMeaningful()
	for i := start; i < len(p.tokens); i++ {
		if p.tokens[i].isContextDelimiter() && p.tokens[i].Value == "|" {
			prev := p.prevMeaningful(i)
			if prev >= start {
				return prev
			}
			return end
		}
		if p.tokens[i].Kind != TokenOpenBracket {
			continue
		}
		prev := p.prevMeaningful(i)
		if prev >= start {
			return prev
		}
		return end
	}
	return end
}

func (p *parser) trailingSegmentLooksLikeReleaseGroup(start int, end int) bool {
	words := 0
	letterWords := 0
	for i := start; i <= end; i++ {
		token := p.tokens[i]
		if token.isDelimiter() || token.isContextDelimiter() {
			continue
		}
		if token.isBracket() || tokenHasHardMetadata(token) {
			return false
		}
		if token.isTextLike() {
			words++
			if tokenContainsLetter(token) {
				letterWords++
			}
		}
	}
	return words > 0 && words <= 5 && letterWords > 0
}

// trailingSegmentIsNonASCII returns true when every text token in the range
// (start..end) consists entirely of non-ASCII letters, making it a poor
// candidate for a release group name.  Legitimate release groups use Latin
// letters/digits; CJK titles and descriptors should not be extracted as RGs.
func (p *parser) trailingSegmentIsNonASCII(start int, end int) bool {
	for i := start; i <= end; i++ {
		token := p.tokens[i]
		if token.isDelimiter() || token.isContextDelimiter() {
			continue
		}
		if !token.isTextLike() {
			continue
		}
		hasASCIILetter := false
		for _, r := range token.Value {
			if r > 0 && r < 128 && unicode.IsLetter(r) {
				hasASCIILetter = true
				break
			}
		}
		if !hasASCIILetter {
			return true
		}
	}
	return false
}

func tokenContainsLetter(token *Token) bool {
	for _, r := range token.Value {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
