package miyo

import (
	"regexp"
	"strconv"
	"strings"
)

var checksumExtractRegexp = regexp.MustCompile(`(?i)(?:^|[^0-9a-f])([0-9a-f]{8})(?:[^0-9a-f]|$)`)

// ExtractChecksum parses the file name (or any string) to extract a potential 8-character hexadecimal checksum.
func ExtractChecksum(value string) string {
	for _, loc := range checksumExtractRegexp.FindAllStringSubmatchIndex(value, -1) {
		if len(loc) < 4 {
			continue
		}
		start := loc[2]
		end := loc[3]
		chk := value[start:end]
		if ContainsHexLetter(chk) {
			return strings.ToUpper(chk)
		}
		if start > 0 && end < len(value) {
			prevChar := value[start-1]
			nextChar := value[end]
			if (prevChar == '[' && nextChar == ']') || (prevChar == '(' && nextChar == ')') || (prevChar == '{' && nextChar == '}') {
				return strings.ToUpper(chk)
			}
		}
	}
	return ""
}

func (p *parser) scanChecksums() {
	for i, token := range p.tokens {
		if !token.isTextLike() || !checksumPattern.MatchString(token.Value) {
			continue
		}
		if !ContainsHexLetter(token.Value) && token.GroupID < 0 {
			continue
		}
		p.addMatch(TagFileChecksum, token.Value, i, i, 5, "checksum")
	}
}

func (p *parser) scanFileIndex() {
	if len(p.tokens) < 4 || !p.tokens[0].isTextLike() || p.tokens[1].Kind != TokenDelimiter || p.tokens[1].Value != "." {
		return
	}
	number, _, ok := cleanNumber(p.tokens[0].Value)
	if !ok {
		return
	}
	next := p.nextMeaningful(1)
	if next == -1 || p.tokens[next].isBracket() || tokenHasHardMetadata(p.tokens[next]) {
		return
	}
	p.addMatch(TagFileIndex, number, 0, 0, 4.8, "file-index")
}

// ContainsHexLetter reports whether value contains at least one hexadecimal
// letter (a–f or A–F). Exported so external tools (e.g. corpus auditors) can
// use the same canonical definition as the parser's checksum scanner.
func ContainsHexLetter(value string) bool {
	for _, r := range value {
		switch r {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F':
			return true
		}
	}
	return false
}

func (p *parser) scanResolutions() {
	for i, token := range p.tokens {
		if !token.isTextLike() {
			continue
		}
		if matches := compactSourceResolutionRegexp.FindStringSubmatch(token.Value); matches != nil {
			p.addMatch(TagSource, canonicalCompactSource(matches[1]), i, i, 4.2, "compact-source-resolution")
			p.addMatch(TagVideoResolution, matches[2], i, i, 4.6, "compact-source-resolution")
			continue
		}
		if _, ok := parseResolution(token.Value); ok {
			if isPhantomAudioCodecHD(p, i, token) {
				continue
			}
			p.addMatch(TagVideoResolution, token.Value, i, i, 5, "resolution")
			continue
		}
		if token.GroupID >= 0 && isBareResolutionNumber(token.Value) {
			p.addMatch(TagVideoResolution, token.Value, i, i, 4.8, "bare-resolution")
		}
	}
}

// isPhantomAudioCodecHD detects standalone "HD"/"FHD" tokens that follow an
// audio-codec name (e.g. "DTS" in "DTS-HD") and should not be treated as
// a resolution alias.
func isPhantomAudioCodecHD(p *parser, index int, token *Token) bool {
	switch normalizeKey(token.Value) {
	case "HD", "FHD":
	default:
		return false
	}
	prev := p.prevMeaningful(index)
	if prev == -1 || !p.tokens[prev].isTextLike() {
		return false
	}
	switch normalizeKey(p.tokens[prev].Value) {
	case "DTS":
		return true
	default:
		return false
	}
}

func isBareResolutionNumber(value string) bool {
	switch value {
	case "480", "576", "720", "1080", "1440", "2160":
		return true
	default:
		return false
	}
}

func canonicalCompactSource(value string) string {
	switch normalizeKey(value) {
	case "BD":
		return "BD"
	case "DVD":
		return "DVD"
	case "WEB":
		return "WEB"
	default:
		return value
	}
}

func (p *parser) scanYears() {
	for i, token := range p.tokens {
		if !token.isTextLike() || !isYear(token.Value, p.options.YearMin, p.options.YearMax) {
			continue
		}
		if tokenHasTag(token, TagFileChecksum) || tokenHasTag(token, TagVideoResolution) {
			continue
		}
		if p.tokenStartsDate(i) || p.tokenEndsDate(i) {
			p.addMatch(TagYear, token.Value, i, i, 4.4, "date-year")
			continue
		}
		if p.tokenStartsYearRange(i) {
			p.addMatch(TagYear, token.Value, i, i, 4.4, "year-range")
			continue
		}
		if p.tokenEndsYearRange(i) {
			continue
		}
		if token.GroupID >= 0 {
			group := p.groups[token.GroupID]
			if group.OpenMark == "(" || p.isIsolatedGroupNumber(i) {
				p.addMatch(TagYear, token.Value, i, i, 4.5, "year")
			}
			continue
		}
		prev := p.prevMeaningful(i)
		next := p.nextMeaningful(i)
		if prev == -1 || next == -1 {
			p.addMatch(TagYear, token.Value, i, i, 3, "year")
			continue
		}
		if p.tokens[next].Kind == TokenOpenBracket && p.bracketGroupHasMetadata(next) {
			p.addMatch(TagYear, token.Value, i, i, 4.2, "year")
			continue
		}
		if (tokenHasHardMetadata(p.tokens[next]) || isStandaloneTechnicalModeMarker(p.tokens[next])) &&
			(p.hasDotDelimiterBefore(i) || p.hasTechnicalMetadataAfter(i)) {
			p.addMatch(TagYear, token.Value, i, i, 4.2, "year")
			continue
		}
		if p.looseYearBeforeTechnicalStack(i) {
			p.addMatch(TagYear, token.Value, i, i, 4.1, "loose-year-technical-stack")
			continue
		}
		if p.isDelimitedYearWithTechnicalMetadataAfter(i) {
			p.addMatch(TagYear, token.Value, i, i, 4.1, "delimited-year-technical-after")
		}
	}
}

func (p *parser) looseYearBeforeTechnicalStack(index int) bool {
	if p.tokens[index].GroupID >= 0 {
		return false
	}
	prev := p.prevMeaningful(index)
	next := p.nextMeaningful(index)
	if prev == -1 || next == -1 {
		return false
	}
	if p.tokenLooksLikeSequenceAfterLooseYear(next) {
		return false
	}
	if tokenHasHardMetadata(p.tokens[prev]) || p.tokens[prev].isContextDelimiter() {
		return false
	}
	return p.hasLooseYearTechnicalStackAfter(next)
}

func (p *parser) tokenLooksLikeSequenceAfterLooseYear(index int) bool {
	token := p.tokens[index]
	if !token.isTextLike() {
		return false
	}
	value := token.Value
	return seasonTypeEpisodeRegexp.MatchString(value) ||
		typeEpisodeRegexp.MatchString(value) ||
		oneXEpisodeRegexp.MatchString(value) ||
		sxeRegexp.MatchString(value) ||
		compactEpisodeRegexp.MatchString(value) ||
		hashEpisodeRegexp.MatchString(value) ||
		seasonRegexp.MatchString(value) ||
		episodeRegexp.MatchString(value) ||
		japaneseSeasonRegexp.MatchString(value) ||
		japaneseEpisodeRegexp.MatchString(value)
}

func (p *parser) hasLooseYearTechnicalStackAfter(start int) bool {
	skippedShortMarkers := 0
	for i := start; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if token.Kind == TokenDelimiter {
			continue
		}
		if token.isContextDelimiter() || token.isBracket() {
			return false
		}
		if isStandaloneTechnicalModeMarker(token) {
			continue
		}
		if tokenHasTag(token, TagVideoResolution) ||
			tokenHasTag(token, TagVideoTerm) ||
			tokenHasTag(token, TagAudioTerm) ||
			tokenHasTag(token, TagSource) ||
			tokenHasTag(token, TagSubtitles) ||
			tokenHasTag(token, TagLanguage) ||
			tokenHasTag(token, TagReleaseInformation) {
			return true
		}
		if token.isTextLike() && token.Value == strings.ToUpper(token.Value) && len(normalizeKey(token.Value)) <= 3 {
			skippedShortMarkers++
			if skippedShortMarkers <= 2 {
				continue
			}
		}
		return false
	}
	return false
}

func (p *parser) hasTechnicalMetadataAfter(index int) bool {
	for i := index + 1; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if token.Kind == TokenDelimiter {
			continue
		}
		if isStandaloneTechnicalModeMarker(token) {
			continue
		}
		if token.isContextDelimiter() || token.Kind == TokenOpenBracket || token.Kind == TokenCloseBracket {
			return false
		}
		return tokenHasTag(token, TagVideoResolution) ||
			tokenHasTag(token, TagVideoTerm) ||
			tokenHasTag(token, TagAudioTerm) ||
			tokenHasTag(token, TagSource) ||
			tokenHasTag(token, TagLanguage) ||
			tokenHasTag(token, TagSubtitles) ||
			tokenHasTag(token, TagReleaseInformation)
	}
	return false
}

func (p *parser) tokenIsInYearRange(index int) bool {
	return p.tokenStartsYearRange(index) || p.tokenEndsYearRange(index)
}

func (p *parser) tokenIsInDate(index int) bool {
	return p.tokenStartsDate(index) || p.tokenIsMiddleOfDate(index) || p.tokenEndsDate(index)
}

func (p *parser) tokenStartsDate(index int) bool {
	if !p.tokens[index].isTextLike() || !isYear(p.tokens[index].Value, p.options.YearMin, p.options.YearMax) {
		return false
	}
	firstDelimiter := p.nextDateComponent(index)
	month := p.nextDateComponent(firstDelimiter)
	secondDelimiter := p.nextDateComponent(month)
	day := p.nextDateComponent(secondDelimiter)
	return p.datePartsAreValid(index, firstDelimiter, month, secondDelimiter, day)
}

func (p *parser) tokenIsMiddleOfDate(index int) bool {
	firstDelimiter := p.prevDateComponent(index)
	year := p.prevDateComponent(firstDelimiter)
	secondDelimiter := p.nextDateComponent(index)
	day := p.nextDateComponent(secondDelimiter)
	return p.datePartsAreValid(year, firstDelimiter, index, secondDelimiter, day)
}

func (p *parser) tokenEndsDate(index int) bool {
	secondDelimiter := p.prevDateComponent(index)
	month := p.prevDateComponent(secondDelimiter)
	firstDelimiter := p.prevDateComponent(month)
	year := p.prevDateComponent(firstDelimiter)
	return p.datePartsAreValid(year, firstDelimiter, month, secondDelimiter, index)
}

func (p *parser) datePartsAreValid(year int, firstDelimiter int, month int, secondDelimiter int, day int) bool {
	if year == -1 || firstDelimiter == -1 || month == -1 || secondDelimiter == -1 || day == -1 {
		return false
	}
	if !p.tokens[year].isTextLike() || !isYear(p.tokens[year].Value, p.options.YearMin, p.options.YearMax) {
		return false
	}
	firstToken := p.tokens[firstDelimiter]
	secondToken := p.tokens[secondDelimiter]
	if !firstToken.isDelimiter() || !secondToken.isDelimiter() {
		return false
	}
	firstVal := firstToken.Value
	secondVal := secondToken.Value
	if (firstVal != "-" && firstVal != "." && firstVal != "_" && firstVal != "/") || firstVal != secondVal {
		return false
	}
	if !p.tokens[month].isTextLike() || !p.tokens[day].isTextLike() {
		return false
	}
	monthValue, _, monthOK := cleanNumber(p.tokens[month].Value)
	dayValue, _, dayOK := cleanNumber(p.tokens[day].Value)
	if !monthOK || !dayOK || len(monthValue) > 2 || len(dayValue) > 2 {
		return false
	}
	if !validDateComponent(monthValue, 1, 12) || !validDateComponent(dayValue, 1, 31) {
		return false
	}
	if p.tokens[year].GroupID >= 0 && (p.tokens[month].GroupID != p.tokens[year].GroupID || p.tokens[day].GroupID != p.tokens[year].GroupID) {
		return false
	}
	return true
}

func validDateComponent(value string, min int, max int) bool {
	number, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	return number >= min && number <= max
}

func (p *parser) tokenStartsYearRange(index int) bool {
	return p.yearRangeNeighbor(index, true)
}

func (p *parser) tokenEndsYearRange(index int) bool {
	return p.yearRangeNeighbor(index, false)
}

func (p *parser) yearRangeNeighbor(index int, forward bool) bool {
	token := p.tokens[index]
	if !token.isTextLike() || !isYear(token.Value, p.options.YearMin, p.options.YearMax) {
		return false
	}
	delimiter := p.nextMeaningful(index)
	if !forward {
		delimiter = p.prevMeaningful(index)
	}
	if delimiter == -1 {
		return false
	}
	isConn := p.tokens[delimiter].isContextDelimiter() && isRangeConnector(p.tokens[delimiter].Value)
	if !isConn && p.tokens[delimiter].Value != "." {
		return false
	}
	other := p.nextMeaningful(delimiter)
	if !forward {
		other = p.prevMeaningful(delimiter)
	}
	if other == -1 || !p.tokens[other].isTextLike() || !isYear(p.tokens[other].Value, p.options.YearMin, p.options.YearMax) {
		return false
	}
	if token.GroupID >= 0 && p.tokens[other].GroupID != token.GroupID {
		return false
	}
	return true
}

func (p *parser) hasTechnicalMetadataAfterAnywhere(index int) bool {
	for i := index; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if tokenHasTag(token, TagVideoResolution) ||
			tokenHasTag(token, TagVideoTerm) ||
			tokenHasTag(token, TagAudioTerm) ||
			tokenHasTag(token, TagSource) ||
			tokenHasTag(token, TagLanguage) ||
			tokenHasTag(token, TagSubtitles) ||
			tokenHasTag(token, TagReleaseInformation) {
			return true
		}
	}
	return false
}

func (p *parser) isDelimitedYearWithTechnicalMetadataAfter(index int) bool {
	if p.tokens[index].GroupID >= 0 {
		return false
	}
	prev := p.prevMeaningful(index)
	next := p.nextMeaningful(index)
	if next == -1 {
		return false
	}
	if p.tokenLooksLikeSequenceAfterLooseYear(next) {
		return false
	}
	if prev != -1 && tokenHasHardMetadata(p.tokens[prev]) {
		return false
	}
	return p.hasTechnicalMetadataAfterAnywhere(next)
}

func (p *parser) nextDateComponent(index int) int {
	if index == -1 {
		return -1
	}
	for i := index + 1; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if token.Kind == TokenDelimiter && (token.Value == " " || token.Value == "\t" || token.Value == "\r" || token.Value == "\n") {
			continue
		}
		return i
	}
	return -1
}

func (p *parser) prevDateComponent(index int) int {
	if index == -1 {
		return -1
	}
	for i := index - 1; i >= 0; i-- {
		token := p.tokens[i]
		if token.Kind == TokenDelimiter && (token.Value == " " || token.Value == "\t" || token.Value == "\r" || token.Value == "\n") {
			continue
		}
		return i
	}
	return -1
}
