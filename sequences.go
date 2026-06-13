package miyo
import (
	"regexp"
	"strconv"
	"strings"
)

var episodePrefixSpecs []struct {
	regexp *regexp.Regexp
	score  float64
	source string
}

func init() {
	for _, spec := range rawPrefixSpecs {
		episodePrefixSpecs = append(episodePrefixSpecs, struct {
			regexp *regexp.Regexp
			score  float64
			source string
		}{*spec.Target, spec.Score, spec.Source})
	}
}

func (p *parser) scanSequences() {
	// Pass 1: Explicit sequence patterns
	for i, token := range p.tokens {
		if !token.isTextLike() || tokenHasHardMetadata(token) {
			continue
		}
		if p.tokenIsInsideLeadingReleaseGroup(i) {
			continue
		}
		if p.options.ParseEpisode && p.scanBracketEpisodeRange(i) {
			continue
		}
		p.scanExplicitSequencePatterns(i)
	}

	// Pass 2: Bare numbers (heuristic scoring)
	for i, token := range p.tokens {
		if !token.isTextLike() || tokenHasHardMetadata(token) {
			continue
		}
		if p.tokenIsInsideLeadingReleaseGroup(i) {
			continue
		}
		if p.options.ParseEpisode {
			number, releaseVersion, ok := cleanNumber(token.Value)
			if ok {
				score := p.scoreEpisodeCandidates(i, number, true)
				if score > 0 {
					if releaseVersion != "" {
						p.addMatch(TagReleaseVersion, releaseVersion, i, i, 4, "episode-context")
					}
				}
			}
		}
	}
}

func (p *parser) scanExplicitSequencePatterns(i int) bool {
	token := p.tokens[i]
	value := token.Value

	if matches := seasonTypeEpisodeRegexp.FindStringSubmatch(value); matches != nil {
		if p.sequenceTokenBelongsToPipeAlternate(i) {
			if p.options.ParseKeywords {
				p.addMatch(TagAnimeType, canonicalAnimeType(matches[2]), i, i, 4.4, "pipe-alternate-season-type-episode")
			}
			if p.options.ParseEpisode {
				p.addMatch(TagEpisodeAlt, stripLeadingZeros(matches[3]), i, i, 4.5, "pipe-alternate-season-type-episode")
			}
			if matches[4] != "" {
				p.addMatch(TagReleaseVersion, matches[4], i, i, 4.1, "pipe-alternate-season-type-episode")
			}
			return true
		}
		if p.options.ParseSeason {
			p.addMatch(TagSeason, stripLeadingZeros(matches[1]), i, i, 4.6, "season-type-episode")
		}
		if p.options.ParseKeywords {
			p.addMatch(TagAnimeType, canonicalAnimeType(matches[2]), i, i, 4.4, "season-type-episode")
		}
		if p.options.ParseEpisode {
			p.addMatch(TagEpisode, stripLeadingZeros(matches[3]), i, i, 4.6, "season-type-episode")
		}
		if matches[4] != "" {
			p.addMatch(TagReleaseVersion, matches[4], i, i, 4.1, "season-type-episode")
		}
		return true
	}
	if matches := typeEpisodeRegexp.FindStringSubmatch(value); matches != nil {
		if p.options.ParseKeywords {
			p.addMatch(TagAnimeType, canonicalAnimeType(matches[1]), i, i, 4.4, "type-episode")
		}
		if p.options.ParseEpisode {
			p.addMatch(TagEpisode, stripLeadingZeros(matches[2]), i, i, 4.5, "type-episode")
		}
		if matches[3] != "" {
			p.addMatch(TagReleaseVersion, matches[3], i, i, 4, "type-episode")
		}
		return true
	}
	if matches := oneXEpisodeRegexp.FindStringSubmatch(value); matches != nil {
		if p.parentheticalSeasonEpisodeHasAbsoluteEpisodeBefore(i) {
			if p.options.ParseSeason {
				p.addMatch(TagSeason, stripLeadingZeros(matches[1]), i, i, 4.3, "parenthetical-absolute-one-x-episode")
			}
			if p.options.ParseEpisode {
				p.addMatch(TagEpisodeAlt, stripLeadingZeros(matches[2]), i, i, 4.6, "parenthetical-absolute-one-x-episode")
				if matches[3] != "" {
					p.addMatch(TagReleaseVersion, stripLeadingZeros(matches[3]), i, i, 4.1, "parenthetical-absolute-one-x-episode")
				}
			}
			return true
		}
		if p.sequenceTokenBelongsToPipeAlternate(i) {
			if p.options.ParseEpisode {
				p.addMatch(TagEpisodeAlt, stripLeadingZeros(matches[2]), i, i, 4.6, "pipe-alternate-one-x-episode")
				if matches[3] != "" {
					p.addMatch(TagReleaseVersion, stripLeadingZeros(matches[3]), i, i, 4.1, "pipe-alternate-one-x-episode")
				}
			}
			return true
		}
		if p.options.ParseSeason {
			p.addMatch(TagSeason, stripLeadingZeros(matches[1]), i, i, 4.5, "one-x-episode")
		}
		if p.options.ParseEpisode {
			p.addMatch(TagEpisode, stripLeadingZeros(matches[2]), i, i, 4.7, "one-x-episode")
			if matches[3] != "" {
				p.addMatch(TagReleaseVersion, stripLeadingZeros(matches[3]), i, i, 4.1, "one-x-episode")
			}
		}
		return true
	}
	if matches := sxeRegexp.FindStringSubmatch(value); matches != nil {
		if p.sequenceTokenBelongsToPipeAlternate(i) {
			if p.options.ParseEpisode {
				p.addMatch(TagEpisodeAlt, stripLeadingZeros(matches[2]), i, i, 4.7, "pipe-alternate-sxe")
				if matches[3] != "" {
					p.addMatch(TagReleaseVersion, stripLeadingZeros(matches[3]), i, i, 4.2, "pipe-alternate-sxe")
				}
			}
			return true
		}
		if p.options.ParseSeason {
			p.addMatch(TagSeason, stripLeadingZeros(matches[1]), i, i, 4.6, "sxe")
		}
		if p.options.ParseEpisode {
			p.addMatch(TagEpisode, stripLeadingZeros(matches[2]), i, i, 4.8, "sxe")
			if matches[3] != "" {
				p.addMatch(TagReleaseVersion, stripLeadingZeros(matches[3]), i, i, 4.2, "sxe")
			}
		}
		return true
	}
	matchedPrefix := false
	for _, spec := range episodePrefixSpecs {
		if matches := spec.regexp.FindStringSubmatch(value); matches != nil {
			if p.options.ParseEpisode {
				p.addMatch(p.episodeTagForToken(i), stripLeadingZeros(matches[1]), i, i, spec.score, spec.source)
				if len(matches) > 2 && matches[2] != "" {
					p.addMatch(TagReleaseVersion, stripLeadingZeros(matches[2]), i, i, 4, spec.source)
				}
			}
			matchedPrefix = true
			break
		}
	}
	if matchedPrefix {
		return true
	}
	if matches := japaneseSeasonRegexp.FindStringSubmatch(value); matches != nil && p.options.ParseSeason {
		p.addMatch(TagSeason, stripLeadingZeros(matches[1]), i, i, 4.6, "japanese-season")
		return true
	}
	if matches := japaneseEpisodeRegexp.FindStringSubmatch(value); matches != nil && p.options.ParseEpisode {
		p.addMatch(p.episodeTagForToken(i), stripLeadingZeros(matches[1]), i, i, 4.6, "japanese-episode")
		return true
	}
	if matches := releaseInfoVersionRegexp.FindStringSubmatch(value); matches != nil {
		if p.options.ParseKeywords {
			p.addMatch(TagReleaseInformation, canonicalReleaseInformation(matches[1]), i, i, 4.2, "release-info-version")
		}
		p.addMatch(TagReleaseVersion, stripLeadingZeros(matches[2]), i, i, 4.1, "release-info-version")
		return true
	}
	if matches := releaseVersionRegexp.FindStringSubmatch(value); matches != nil {
		p.addMatch(TagReleaseVersion, stripLeadingZeros(matches[1]), i, i, 3.8, "release-version")
		return true
	}
	if p.options.ParseSeason && p.scanBareOrdinalSeason(i) {
		return true
	}
	if p.options.ParseSeason && p.scanBareNumberSeasonBeforeEpisode(i) {
		return true
	}
	if p.options.ParsePart && p.scanPostRangeTitleIndex(i) {
		return true
	}
	if p.options.ParseSeason && p.scanTVSeasonLabel(i) {
		return true
	}
	if p.options.ParseSeason && p.scanConnectedSeasonNumber(i) {
		return true
	}
	if partialEpisodeRegexp.MatchString(value) && p.options.ParseEpisode && p.isEpisodeByContext(i) {
		p.addMatch(p.episodeTagForToken(i), value, i, i, 4.2, "partial-episode")
		return true
	}
	if p.options.ParseEpisode && p.scanFractionalEpisode(i) {
		return true
	}
	if matches := seasonRegexp.FindStringSubmatch(value); matches != nil && p.options.ParseSeason {
		if p.sequenceTokenBelongsToTrailingAlternateGroup(i) || p.sequenceTokenBelongsToPipeAlternate(i) {
			return false
		}
		p.addMatch(TagSeason, stripLeadingZeros(matches[1]), i, i, 4.3, "season-token")
		return true
	}
	if matches := volumeRegexp.FindStringSubmatch(value); matches != nil && p.options.ParseVolume {
		if p.sequenceTokenBelongsToTrailingAlternateGroup(i) || p.sequenceTokenBelongsToPipeAlternate(i) {
			return false
		}
		p.addMatch(TagVolume, stripLeadingZeros(matches[1]), i, i, 4.3, "volume-token")
		return true
	}
	if p.scanSeparatedSequence(i) {
		return true
	}
	p.scanEpisodeTotal(i)
	return false
}

func (p *parser) scanFractionalEpisode(index int) bool {
	if index+2 >= len(p.tokens) {
		return false
	}
	left := p.tokens[index]
	dot := p.tokens[index+1]
	right := p.tokens[index+2]
	if !left.isTextLike() || dot.Kind != TokenDelimiter || dot.Value != "." || !right.isTextLike() {
		return false
	}
	if _, _, ok := cleanNumber(left.Value); !ok {
		return false
	}
	if len(right.Value) != 1 || right.Value[0] < '1' || right.Value[0] > '5' {
		return false
	}
	if !p.isEpisodeByContext(index) {
		return false
	}
	p.addMatch(p.episodeTagForToken(index), left.Value+"."+right.Value, index, index+2, 4.3, "fractional-episode")
	return true
}

func (p *parser) scanBareOrdinalSeason(index int) bool {
	season, ok := ordinalNumber(p.tokens[index].Value)
	if !ok || p.sequenceTokenBelongsToTrailingAlternateGroup(index) {
		return false
	}
	if p.scanSeasonFollowedByEpisode(index, season, "bare-ordinal-season-before-episode") {
		return true
	}
	next := p.nextMeaningful(index)
	if next == -1 || !p.tokens[next].isContextDelimiter() || p.tokens[next].Value != "-" {
		return false
	}
	episodeIndex := p.nextMeaningful(next)
	if episodeIndex == -1 || !p.tokens[episodeIndex].isTextLike() {
		return false
	}
	if _, _, ok := cleanNumber(p.tokens[episodeIndex].Value); !ok {
		return false
	}
	prev := p.prevMeaningful(index)
	if prev == -1 || p.tokens[prev].isContextDelimiter() || tokenHasHardMetadata(p.tokens[prev]) {
		return false
	}
	p.addMatch(TagSeason, season, index, index, 4.3, "bare-ordinal-season")
	return true
}

func (p *parser) scanBareNumberSeasonBeforeEpisode(index int) bool {
	number, _, ok := cleanNumber(p.tokens[index].Value)
	if !ok || len(number) > 2 || len(p.tokens[index].Value) != len(number) {
		return false
	}
	if len(number) == 1 && (number == "0" || number == "1" || number == "2") {
		return false
	}
	next := p.nextMeaningful(index)
	if next == -1 || !p.tokens[next].isContextDelimiter() || p.tokens[next].Value != "-" {
		return false
	}
	return p.scanSeasonFollowedByEpisode(index, number, "bare-number-season-before-episode")
}

func (p *parser) scanBracketEpisodeRange(index int) bool {
	if p.tokens[index].GroupID < 0 || p.tokens[index].GroupID >= len(p.groups) {
		return false
	}
	group := p.groups[p.tokens[index].GroupID]
	if group.Open+1 != index || p.groupLooksLikeYearRange(group) {
		return false
	}
	first, second, ok := p.groupEpisodeRangeEndpoints(group)
	if !ok || first != index {
		return false
	}
	firstNumber, _, firstOK := cleanNumber(p.tokens[first].Value)
	secondNumber, _, secondOK := cleanNumber(p.tokens[second].Value)
	if !firstOK || !secondOK {
		return false
	}
	p.addMatch(TagEpisode, firstNumber, first, first, 4.5, "bracket-episode-range")
	p.addMatch(TagEpisode, secondNumber, second, second, 4.5, "bracket-episode-range")
	return true
}

func (p *parser) scanPostRangeTitleIndex(index int) bool {
	if !p.numberFollowsLeadingEpisodeRangeTitle(index) {
		return false
	}
	number, _, ok := cleanNumber(p.tokens[index].Value)
	if !ok {
		return false
	}
	p.addMatch(TagPart, number, index, index, 3.7, "post-range-title-index")
	return true
}

func (p *parser) scanSeasonFollowedByEpisode(index int, season string, source string) bool {
	if !p.hasTitleTextBeforeSeasonShorthand(index) {
		return false
	}
	episodeIndex := p.seasonShorthandEpisodeIndex(index)
	if episodeIndex == -1 {
		return false
	}
	episode, releaseVersion, ok := cleanNumber(p.tokens[episodeIndex].Value)
	if !ok {
		return false
	}
	if !p.sequenceHasTechnicalBoundaryAfter(episodeIndex) {
		return false
	}
	p.addMatch(TagSeason, season, index, index, 4.3, source)
	if p.options.ParseEpisode {
		p.addMatch(TagEpisode, episode, episodeIndex, episodeIndex, 4.2, source)
		if releaseVersion != "" {
			p.addMatch(TagReleaseVersion, releaseVersion, episodeIndex, episodeIndex, 4, source)
		}
	}
	return true
}

func (p *parser) seasonShorthandEpisodeIndex(index int) int {
	next := p.nextMeaningful(index)
	if next == -1 {
		return -1
	}
	if p.tokens[next].isContextDelimiter() {
		if p.tokens[next].Value != "-" {
			return -1
		}
		next = p.nextMeaningful(next)
	}
	if next == -1 || !p.tokens[next].isTextLike() {
		return -1
	}
	return next
}

func (p *parser) hasTitleTextBeforeSeasonShorthand(index int) bool {
	prev := p.prevMeaningful(index)
	if prev == -1 || p.tokens[prev].isContextDelimiter() || tokenHasHardMetadata(p.tokens[prev]) {
		return false
	}
	return p.tokens[prev].isTextLike()
}

func (p *parser) scanTVSeasonLabel(index int) bool {
	number, _, ok := cleanNumber(p.tokens[index].Value)
	if !ok || !p.numberIsTvSeasonLabel(index) {
		return false
	}
	delimiter := p.prevMeaningful(index)
	prefix := p.prevMeaningful(delimiter)
	if prefix != -1 {
		p.tokens[prefix].addPossibilityWithDescriptor(TagSequencePrefix, TagSeason, 3.5, p.tokens[prefix].Value, "tv-season-label")
	}
	p.addMatch(TagSeason, number, index, index, 4.2, "tv-season-label")
	return true
}

func (p *parser) scanConnectedSeasonNumber(index int) bool {
	connector := p.prevMeaningful(index)
	if connector == -1 || !p.tokens[connector].isContextDelimiter() {
		return false
	}
	switch p.tokens[connector].Value {
	case "+", "&":
	case "-":
		if !p.connectedSeasonDashContinuesList(index) {
			return false
		}
	default:
		return false
	}
	previousNumber := p.prevMeaningful(connector)
	if previousNumber == -1 {
		return false
	}
	if !tokenHasTag(p.tokens[previousNumber], TagSeason) && !p.numberFollowsSeasonPrefix(previousNumber) {
		return false
	}
	// Season + type suffix (e.g. "S1+SP" or "S2+OVA")
	if animeType := isAnimeTypeToken(p.tokens[index].Value); animeType != "" && tokenHasTag(p.tokens[previousNumber], TagSeason) {
		if p.options.ParseKeywords {
			p.addMatch(TagAnimeType, animeType, index, index, 4.1, "connected-season-anime-type")
		}
		return true
	}
	number, _, ok := cleanNumber(p.tokens[index].Value)
	if !ok {
		return false
	}
	p.addMatch(TagSeason, number, index, index, 4.1, "connected-season-number")
	return true
}

func (p *parser) connectedSeasonDashContinuesList(index int) bool {
	next := p.nextMeaningful(index)
	if next == -1 {
		return false
	}
	if p.tokens[next].isContextDelimiter() {
		switch p.tokens[next].Value {
		case "+", "&":
			return true
		default:
			return false
		}
	}
	return tokenHasTag(p.tokens[next], TagAnimeType)
}

func (p *parser) numberFollowsSeasonPrefix(index int) bool {
	prefix := p.prevMeaningful(index)
	return prefix != -1 && isSeasonPrefixKey(normalizeKey(p.tokens[prefix].Value))
}

func (p *parser) scanSeparatedSequence(prefixIndex int) bool {
	prefix := normalizeKey(p.tokens[prefixIndex].Value)
	if p.sequencePrefixBelongsToTrailingAlternateGroup(prefixIndex, prefix) {
		return false
	}
	if p.seasonOrPartPrefixBelongsToPipeAlternate(prefixIndex, prefix) {
		return false
	}
	if isSeasonPrefixKey(prefix) && p.options.ParseSeason {
		if prev := p.prevMeaningful(prefixIndex); prev != -1 {
			if ordinal, ok := ordinalNumber(p.tokens[prev].Value); ok {
				p.tokens[prefixIndex].addPossibilityWithDescriptor(TagSequencePrefix, TagSeason, 3.5, p.tokens[prefixIndex].Value, "sequence-prefix")
				p.addMatch(TagSeason, ordinal, prev, prefixIndex, 4.4, "ordinal-season")
				return true
			}
		}
	}
	next := p.nextMeaningful(prefixIndex)
	if next == -1 || !p.tokens[next].isTextLike() {
		return false
	}
	number, releaseVersion, ok := cleanNumber(p.tokens[next].Value)
	if !ok {
		return false
	}
	switch prefix {
	case "EP", "EPS", "EPISODE", "EPISODES", "CAPITULO", "EPISODIO", "FOLGE":
		if !p.options.ParseEpisode {
			return false
		}
		if p.episodePrefixBelongsToEpisodeTitle(prefixIndex) {
			return false
		}
		p.tokens[prefixIndex].addPossibilityWithDescriptor(TagSequencePrefix, TagEpisode, 3.5, p.tokens[prefixIndex].Value, "sequence-prefix")
		tag := p.episodeTagForToken(next)
		p.addMatch(tag, number, next, next, 4.4, "episode-prefix")
		if releaseVersion != "" {
			p.addMatch(TagReleaseVersion, releaseVersion, next, next, 4, "episode-prefix")
		}
		return true
	case "S", "SEASON", "SAISON", "SEASONS", "SAISONS":
		if !p.options.ParseSeason {
			return false
		}
		if prev := p.prevMeaningful(prefixIndex); prev != -1 {
			if ordinal, ok := ordinalNumber(p.tokens[prev].Value); ok {
				p.tokens[prefixIndex].addPossibilityWithDescriptor(TagSequencePrefix, TagSeason, 3.5, p.tokens[prefixIndex].Value, "sequence-prefix")
				p.addMatch(TagSeason, ordinal, prev, prev, 4.4, "ordinal-season")
				if p.options.ParseEpisode && next != -1 {
					p.addMatch(TagEpisode, number, next, next, 3.9, "episode-after-ordinal-season")
					if releaseVersion != "" {
						p.addMatch(TagReleaseVersion, releaseVersion, next, next, 3.8, "episode-after-ordinal-season")
					}
				}
				return true
			}
		}
		p.tokens[prefixIndex].addPossibilityWithDescriptor(TagSequencePrefix, TagSeason, 3.5, p.tokens[prefixIndex].Value, "sequence-prefix")
		p.addMatch(TagSeason, number, next, next, 4.2, "season-prefix")
		return true
	case "VOL", "VOLUME", "VOLUMES":
		if !p.options.ParseVolume {
			return false
		}
		p.tokens[prefixIndex].addPossibilityWithDescriptor(TagSequencePrefix, TagVolume, 3.5, p.tokens[prefixIndex].Value, "sequence-prefix")
		p.addMatch(TagVolume, number, next, next, 4.2, "volume-prefix")
		return true
	case "PART", "COUR":
		if !p.options.ParsePart {
			return false
		}
		p.tokens[prefixIndex].addPossibilityWithDescriptor(TagSequencePrefix, TagPart, 3.5, p.tokens[prefixIndex].Value, "sequence-prefix")
		p.addMatch(TagPart, number, next, next, 4.2, "part-prefix")
		return true
	}
	return false
}

func isSeasonPrefixKey(key string) bool {
	switch key {
	case "S", "SEASON", "SAISON", "SEASONS", "SAISONS":
		return true
	default:
		return false
	}
}

func (p *parser) sequencePrefixBelongsToTrailingAlternateGroup(prefixIndex int, prefix string) bool {
	switch prefix {
	case "EP", "EPS", "EPISODE", "EPISODES", "CAPITULO", "EPISODIO", "FOLGE",
		"S", "SEASON", "SAISON", "SEASONS", "SAISONS",
		"PART", "COUR":
		return p.sequenceTokenBelongsToTrailingAlternateGroup(prefixIndex)
	default:
		return false
	}
}

func (p *parser) seasonOrPartPrefixBelongsToPipeAlternate(prefixIndex int, prefix string) bool {
	switch prefix {
	case "S", "SEASON", "SAISON", "SEASONS", "SAISONS", "PART", "COUR":
		return p.sequenceTokenBelongsToPipeAlternate(prefixIndex)
	default:
		return false
	}
}

func (p *parser) sequenceTokenBelongsToTrailingAlternateGroup(index int) bool {
	if !p.hasMainEpisodeBefore(index) {
		return false
	}
	groupID := p.tokens[index].GroupID
	if groupID < 0 || groupID >= len(p.groups) {
		return false
	}
	group := p.groups[groupID]
	if group.OpenMark != "(" {
		return false
	}
	return p.groupHasTitleTextBeforeIndex(group, index)
}

func (p *parser) sequenceTokenBelongsToPipeAlternate(index int) bool {
	if !p.hasMainEpisodeSignalBefore(index) {
		return false
	}
	for i := index - 1; i >= 0; i-- {
		token := p.tokens[i]
		if token.Kind == TokenDelimiter {
			continue
		}
		if token.isContextDelimiter() && token.Value == "|" {
			return true
		}
	}
	return false
}

func (p *parser) hasMainEpisodeSignalBefore(index int) bool {
	if p.hasMainEpisodeBefore(index) {
		return true
	}
	for i := 0; i < index; i++ {
		token := p.tokens[i]
		if !token.isTextLike() || tokenHasHardMetadata(token) || p.tokenIsInsideLeadingReleaseGroup(i) {
			continue
		}
		value := token.Value
		if seasonTypeEpisodeRegexp.MatchString(value) ||
			typeEpisodeRegexp.MatchString(value) ||
			oneXEpisodeRegexp.MatchString(value) ||
			sxeRegexp.MatchString(value) ||
			compactEpisodeRegexp.MatchString(value) ||
			hashEpisodeRegexp.MatchString(value) ||
			episodeRegexp.MatchString(value) ||
			japaneseEpisodeRegexp.MatchString(value) {
			return true
		}
		if _, _, ok := cleanNumber(value); ok && p.isEpisodeByContext(i) {
			return true
		}
	}
	return false
}

func (p *parser) groupHasTitleTextBeforeIndex(group BracketGroup, index int) bool {
	for i := group.Open + 1; i < index && i < group.Close; i++ {
		token := p.tokens[i]
		if token.isDelimiter() || token.isBracket() {
			continue
		}
		if !token.isTextLike() || tokenHasHardMetadata(token) {
			continue
		}
		if _, _, ok := cleanNumber(token.Value); ok {
			continue
		}
		if _, ok := ordinalNumber(token.Value); ok {
			continue
		}
		return true
	}
	return false
}

func (p *parser) episodePrefixBelongsToEpisodeTitle(prefixIndex int) bool {
	if !p.hasMainEpisodeBefore(prefixIndex) {
		return false
	}
	prev := p.prevMeaningful(prefixIndex)
	return prev != -1 && p.tokens[prev].isTextLike() && !tokenHasHardMetadata(p.tokens[prev])
}

func (p *parser) scanEpisodeTotal(index int) {
	if !p.options.ParseEpisode || normalizeWord(p.tokens[index].Value) != "OF" {
		return
	}
	prev := p.prevMeaningful(index)
	next := p.nextMeaningful(index)
	if prev == -1 || next == -1 {
		return
	}
	if p.tokens[index].GroupID >= 0 {
		if p.tokens[prev].GroupID != p.tokens[index].GroupID || p.tokens[next].GroupID != p.tokens[index].GroupID {
			return
		}
	}
	episode, _, ok := cleanNumber(p.tokens[prev].Value)
	if !ok {
		return
	}
	total, _, ok := cleanNumber(p.tokens[next].Value)
	if !ok {
		return
	}
	p.addMatch(TagEpisode, episode, prev, prev, 4.5, "episode-total")
	p.addMatch(TagOtherEpisodeNumber, total, next, next, 4.4, "episode-total")
}

// scoreEpisodeCandidates evaluates a set of named, weighted signals for the
// token at index (which must already be known to parse as a number) and returns
// the net episode-likelihood score.  Crucially it also registers a TagEpisode
// Possibility on the token itself — carrying the combined source string that
// lists every signal that fired — so the result is fully visible in DebugResult.
//
// Positive score ⟹ context supports episode interpretation.
// Negative score ⟹ context vetoes episode interpretation.
//
// An additional TagSeason Possibility is registered when the tv-season-label
// signal fires, because that pattern both suppresses episode and suggests season.
func (p *parser) scoreEpisodeCandidates(index int, number string, register bool) float64 {
	token := p.tokens[index]

	type signal struct {
		name  string
		delta float64
		extra Tag // if non-empty, also emit a Possibility for this secondary tag
	}

	var fired []signal

	// ── Negative signals ─────────────────────────────────────────────────────
	// Each one suppresses the episode interpretation by a calibrated amount.
	// Strong vetoes use large magnitudes so they dominate positive signals even
	// when several fire simultaneously.

	// Inside the leading release-group bracket — numbers here belong to the
	// release group name, never to the episode sequence.
	if p.tokenIsInsideLeadingReleaseGroup(index) {
		fired = append(fired, signal{"inside-leading-rg", -5.0, ""})
	}
	// Immediately follows a BATCH tag — it is an index into the batch, not an
	// episode number.
	if p.numberIsBatchIndex(index) {
		fired = append(fired, signal{"batch-index", -4.0, ""})
	}
	// Part of a YYYY-MM-DD date triplet.
	if p.tokenIsInDate(index) {
		fired = append(fired, signal{"in-date", -4.0, ""})
	}
	// "TV 2" / "TV-2" pattern — the number qualifies the TV broadcast label,
	// making it a season number. Emit a season possibility as a side-effect.
	if p.numberIsTvSeasonLabel(index) {
		fired = append(fired, signal{"tv-season-label", -2.0, TagSeason})
	}
	// Follows a closed episode-range group ([01-12]) and is therefore a
	// part/title index, not another episode number.
	if p.numberFollowsLeadingEpisodeRangeTitle(index) {
		fired = append(fired, signal{"leading-range-title", -3.0, ""})
	}
	// Inside a bracket group, preceded by a dash and a short alpha prefix:
	// looks like a catalog code suffix (e.g. "ABC-123"), not an episode.
	if p.numberIsCatalogCodeSuffix(index) {
		fired = append(fired, signal{"catalog-code-suffix", -3.0, ""})
	}
	// Follows a dot from a two-digit number that looks like a channel base
	// (e.g. "2" in "5.1" or "AAC2.0") — this is a channel count suffix.
	if p.numberIsAudioChannelComponent(index) {
		fired = append(fired, signal{"audio-channel", -4.0, ""})
	}
	if p.numberIsAudioChannelBase(index) {
		fired = append(fired, signal{"audio-channel-base", -4.0, ""})
	}
	// Part of a year-to-year range (e.g. "2020-2021").
	if p.tokenIsInYearRange(index) {
		fired = append(fired, signal{"in-year-range", -3.0, ""})
	}
	// Preceded by "No." — ordinal with explicit "No." prefix is a title
	// qualifier, not an episode (e.g. "Vol.1 No.3").
	if p.numberIsNoTitleOrdinal(index) {
		fired = append(fired, signal{"no-title-ordinal", -3.5, ""})
	}
	// Single digit preceded by title-text and followed by a dash and a zero-
	// padded number: the single digit is a title suffix ("Title 2 - 01").
	if p.numberIsSmallTitleSuffixBeforeDashEpisode(index) {
		fired = append(fired, signal{"small-title-prefix", -2.5, ""})
	}
	// Ungrouped number that follows a pipe separator — it belongs to an
	// alternate-language title section, not to the metadata.
	if p.plainNumberBelongsToPipeAlternateTitle(index) {
		fired = append(fired, signal{"pipe-alt-title", -2.5, ""})
	}
	// The specific "U-17" / "U-12" qualifier pattern where a letter prefix is
	// immediately attached before the dash.
	if p.numberIsHyphenatedTitleQualifier(index) {
		fired = append(fired, signal{"hyphenated-qualifier", -2.0, ""})
	}

	// ── Positive signals ─────────────────────────────────────────────────────
	// Each one increases confidence that the token is an episode number.

	// The number is the sole non-whitespace content inside its bracket group —
	// e.g. "[05]" — a strong indicator of an episode number.
	if token.GroupID >= 0 && p.isIsolatedGroupNumber(index) {
		fired = append(fired, signal{"isolated-group", +2.0, ""})
	}
	// The number is connected via a range connector to another number (e.g.
	// "01-12") — it is definitely part of an episode range.
	if p.tokenIsInEpisodeRange(index) {
		fired = append(fired, signal{"episode-range", +2.0, ""})
	}
	// Precedes " - " followed by non-numeric title text — the classic
	// "[Group] Title - 05 - Episode Title" layout.
	if p.numberBeforeDashedEpisodeTitle(index) {
		fired = append(fired, signal{"before-dashed-title", +1.8, ""})
	}
	// Trailing number pattern — e.g. "Title 05" or "Title 05.mkv"
	if p.isTrailingEpisodeCandidate(index) {
		fired = append(fired, signal{"trailing-number", +1.5, ""})
	}

	prev := p.prevMeaningful(index)
	next := p.nextMeaningful(index)

	// Preceded by a context delimiter (dash, tilde, pipe…) — the most common
	// position for a bare episode number in fansub naming.
	if prev != -1 && p.tokens[prev].isContextDelimiter() {
		fired = append(fired, signal{"context-delimiter-prev", +1.5, ""})
	}
	// The token is the first element (no prev), followed by a dash and then
	// non-numeric text — e.g. a file that starts with "01 - Title".
	if prev == -1 && next != -1 && p.tokens[next].isContextDelimiter() {
		afterDelimiter := p.nextMeaningful(next)
		if afterDelimiter != -1 {
			if _, _, ok := cleanNumber(p.tokens[afterDelimiter].Value); !ok {
				fired = append(fired, signal{"leading-number-dash-text", +1.5, ""})
			}
		}
	}
	// Next non-delimiter is an open bracket whose group contains resolved
	// metadata — the number is almost certainly an episode preceding a tech
	// metadata bracket like "[1080p HEVC]".
	if prev != -1 && next != -1 && p.tokens[next].Kind == TokenOpenBracket && p.bracketGroupHasMetadata(next) {
		fired = append(fired, signal{"metadata-bracket-next", +1.5, ""})
	}
	// The number appears in a dot-delimited chain where the tokens before and
	// after are metadata — e.g. "Group.Title.05.1080p.HEVC".
	if p.hasDotDelimiterBefore(index) && p.hasDotDelimitedMetadataAfter(index) {
		fired = append(fired, signal{"dot-delimited-both", +1.5, ""})
	}

	if len(fired) == 0 {
		return 0
	}

	// Accumulate the net episode score and collect signal names for the source
	// string.  The source string makes every fired signal visible in DebugResult.
	var netEpisode float64
	names := make([]string, 0, len(fired))
	for _, s := range fired {
		netEpisode += s.delta
		names = append(names, s.name)
	}

	if register && netEpisode > 0 {
		source := "ep-ctx:" + strings.Join(names, ",")
		tag := p.episodeTagForToken(index)
		p.addMatch(tag, number, index, index, 3.8, source)
	}

	// Register secondary-tag possibilities for signals that imply another tag
	// (currently only tv-season-label → season).  These are affirmative signals
	// that stand on their own regardless of the episode net score.
	if register {
		for _, s := range fired {
			if s.extra == "" {
				continue
			}
			token.addPossibility(s.extra, 2.0, number, "ep-ctx:"+s.name)
		}
	}

	return netEpisode
}

// isEpisodeByContext is a thin boolean wrapper around scoreEpisodeCandidates.
// It exists to preserve the call sites in scanFractionalEpisode and
// hasMainEpisodeSignalBefore that still need a bool result.  New code should
// call scoreEpisodeCandidates directly when it needs the numeric score or wants
// to avoid double-calling addPossibility.
func (p *parser) isEpisodeByContext(index int) bool {
	number, _, ok := cleanNumber(p.tokens[index].Value)
	if !ok {
		if partialEpisodeRegexp.MatchString(p.tokens[index].Value) {
			number = p.tokens[index].Value
		} else {
			return false
		}
	}
	return p.scoreEpisodeCandidates(index, number, false) > 0
}

func (p *parser) parentheticalSeasonEpisodeHasAbsoluteEpisodeBefore(index int) bool {
	token := p.tokens[index]
	if token.GroupID < 0 || token.GroupID >= len(p.groups) {
		return false
	}
	group := p.groups[token.GroupID]
	if group.OpenMark != "(" {
		return false
	}
	beforeGroup := p.prevMeaningful(group.Open)
	if beforeGroup == -1 || !p.tokens[beforeGroup].isTextLike() {
		return false
	}
	if _, _, ok := cleanNumber(p.tokens[beforeGroup].Value); !ok {
		return false
	}
	beforeNumber := p.prevMeaningful(beforeGroup)
	return beforeNumber != -1 && p.tokens[beforeNumber].isTextLike() && !tokenHasHardMetadata(p.tokens[beforeNumber])
}

func (p *parser) numberIsBatchIndex(index int) bool {
	prev := p.prevMeaningful(index)
	return prev != -1 &&
		normalizeKey(p.tokens[prev].Value) == "BATCH" &&
		tokenHasTag(p.tokens[prev], TagReleaseInformation)
}

func (p *parser) isTrailingEpisodeCandidate(index int) bool {
	if p.tokens[index].GroupID >= 0 {
		return false
	}
	if p.hasMainEpisode() {
		return false
	}
	// Check if there are any meaningful text/number tokens after this index
	for i := index + 1; i < len(p.tokens); i++ {
		t := p.tokens[i]
		if t.isDelimiter() || t.isContextDelimiter() || t.isBracket() || t.GroupID >= 0 {
			continue
		}
		if tokenHasTag(t, TagFileExtension) || tokenHasTag(t, TagFileChecksum) || tokenHasHardMetadata(t) || tokenHasTag(t, TagReleaseGroup) || p.isTrailingReleaseGroupToken(i) {
			continue
		}
		// Any other text token means it is not trailing
		return false
	}
	// Make sure there is some title-like text before it to represent the title
	hasTitleText := false
	for i := 0; i < index; i++ {
		t := p.tokens[i]
		if t.isTextLike() && !tokenHasHardMetadata(t) {
			hasTitleText = true
			break
		}
	}
	return hasTitleText
}

func (p *parser) numberFollowsLeadingEpisodeRangeTitle(index int) bool {
	if p.tokens[index].GroupID >= 0 {
		return false
	}
	if _, _, ok := cleanNumber(p.tokens[index].Value); !ok {
		return false
	}
	prev := p.prevMeaningful(index)
	if prev == -1 || !p.tokens[prev].isTextLike() || tokenHasHardMetadata(p.tokens[prev]) {
		return false
	}
	if !p.sequenceHasTechnicalBoundaryAfter(index) {
		return false
	}
	for _, group := range p.groups {
		if group.Open > index {
			break
		}
		if group.Close >= index || group.Starts {
			continue
		}
		if p.groupIsEpisodeRange(group) {
			return true
		}
	}
	return false
}

func (p *parser) groupIsEpisodeRange(group BracketGroup) bool {
	_, _, ok := p.groupEpisodeRangeEndpoints(group)
	return ok
}

func (p *parser) groupEpisodeRangeEndpoints(group BracketGroup) (int, int, bool) {
	first := p.nextMeaningful(group.Open)
	if first == -1 || first >= group.Close || !p.tokens[first].isTextLike() {
		return -1, -1, false
	}
	delimiter := p.nextMeaningful(first)
	if delimiter == -1 || delimiter >= group.Close || !p.tokens[delimiter].isContextDelimiter() || !isRangeConnector(p.tokens[delimiter].Value) {
		return -1, -1, false
	}
	second := p.nextMeaningful(delimiter)
	if second == -1 || second >= group.Close || !p.tokens[second].isTextLike() {
		return -1, -1, false
	}
	after := p.nextMeaningful(second)
	if after != group.Close {
		return -1, -1, false
	}
	_, _, firstOK := cleanNumber(p.tokens[first].Value)
	_, _, secondOK := cleanNumber(p.tokens[second].Value)
	if !firstOK || !secondOK {
		return -1, -1, false
	}
	return first, second, true
}

func (p *parser) numberBeforeDashedEpisodeTitle(index int) bool {
	number, _, ok := cleanNumber(p.tokens[index].Value)
	if !ok || len(number) < 3 {
		return false
	}
	prev := p.prevMeaningful(index)
	if prev == -1 || !p.tokens[prev].isTextLike() || tokenHasHardMetadata(p.tokens[prev]) {
		return false
	}
	next := p.nextMeaningful(index)
	if next == -1 || !p.tokens[next].isContextDelimiter() || p.tokens[next].Value != "-" {
		return false
	}
	if p.tokens[next].Start == p.tokens[index].End {
		return false
	}
	afterDash := p.nextMeaningful(next)
	if afterDash == -1 || !p.tokens[afterDash].isTextLike() || tokenHasHardMetadata(p.tokens[afterDash]) {
		return false
	}
	if p.tokens[afterDash].Start == p.tokens[next].End {
		return false
	}
	if _, _, ok := cleanNumber(p.tokens[afterDash].Value); ok {
		return false
	}
	return p.episodeTitlePhraseHasTechnicalBoundaryAfter(afterDash)
}

func (p *parser) sequenceHasTechnicalBoundaryAfter(index int) bool {
	for i := index + 1; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if token.Kind == TokenDelimiter || token.isContextDelimiter() || token.Kind == TokenCloseBracket {
			continue
		}
		if token.Kind == TokenOpenBracket {
			return p.bracketGroupHasMetadata(i)
		}
		return tokenHasHardMetadata(token)
	}
	return false
}

func (p *parser) episodeTitlePhraseHasTechnicalBoundaryAfter(index int) bool {
	for i := index + 1; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if token.Kind == TokenDelimiter || token.Kind == TokenCloseBracket {
			continue
		}
		if token.isContextDelimiter() {
			if token.Value == "|" {
				return false
			}
			continue
		}
		if token.Kind == TokenOpenBracket {
			return p.bracketGroupHasMetadata(i)
		}
		if tokenHasHardMetadata(token) {
			return true
		}
	}
	return false
}

func (p *parser) plainNumberBelongsToPipeAlternateTitle(index int) bool {
	if p.tokens[index].GroupID >= 0 || !p.hasPipeBefore(index) {
		return false
	}
	prev := p.prevMeaningful(index)
	return prev != -1 && p.tokens[prev].isTextLike() && !tokenHasHardMetadata(p.tokens[prev])
}

func (p *parser) hasPipeBefore(index int) bool {
	for i := index - 1; i >= 0; i-- {
		token := p.tokens[i]
		if token.Kind == TokenDelimiter {
			continue
		}
		if token.isContextDelimiter() && token.Value == "|" {
			return true
		}
	}
	return false
}

func (p *parser) tokenIsInsideLeadingReleaseGroup(index int) bool {
	groupID := p.tokens[index].GroupID
	if groupID < 0 || groupID >= len(p.groups) {
		return false
	}
	group := p.groups[groupID]
	return group.Starts &&
		!p.groupLooksLikeLeadingBracketTitle(group) &&
		p.groupLooksLikeReleaseGroup(group)
}

func (p *parser) numberIsNoTitleOrdinal(index int) bool {
	prev := p.prevMeaningful(index)
	if prev == -1 || normalizeWord(p.tokens[prev].Value) != "NO" {
		return false
	}
	return p.hasDotDelimiterBefore(index)
}

func (p *parser) numberIsSmallTitleSuffixBeforeDashEpisode(index int) bool {
	number, _, ok := cleanNumber(p.tokens[index].Value)
	if !ok || len(number) < 1 || len(number) > 3 || len(p.tokens[index].Value) > 3 {
		return false
	}
	if p.tokens[index].Value[0] == '0' {
		return false
	}
	prev := p.prevMeaningful(index)
	next := p.nextMeaningful(index)
	if prev == -1 || next == -1 || p.tokens[next].Value != "-" {
		return false
	}
	if p.tokens[prev].isContextDelimiter() || tokenHasHardMetadata(p.tokens[prev]) {
		return false
	}
	afterDash := p.nextMeaningful(next)
	if afterDash == -1 || !p.tokens[afterDash].isTextLike() {
		return false
	}
	afterDashClean, _, ok := cleanNumber(p.tokens[afterDash].Value)
	if !ok || len(p.tokens[afterDash].Value) < 2 {
		return false
	}

	if len(number) == 1 {
		return true
	}

	// For multi-digit suffixes (e.g. 21 in "Eyeshield 21 - 001"):
	// 1. If the episode after dash is zero-padded (e.g. starts with "0")
	if p.tokens[afterDash].Value[0] == '0' {
		return true
	}

	// 2. Or if the suffix is mathematically greater than or equal to the episode number
	n1, err1 := strconv.Atoi(number)
	n2, err2 := strconv.Atoi(afterDashClean)
	if err1 == nil && err2 == nil && n1 >= n2 {
		return true
	}

	return false
}

func (p *parser) numberIsCatalogCodeSuffix(index int) bool {
	if p.tokens[index].GroupID < 0 {
		return false
	}
	prev := p.prevMeaningful(index)
	if prev == -1 || p.tokens[prev].Value != "-" || p.tokens[prev].GroupID != p.tokens[index].GroupID {
		return false
	}
	prefix := p.prevMeaningful(prev)
	if prefix == -1 || p.tokens[prefix].GroupID != p.tokens[index].GroupID {
		return false
	}
	key := normalizeKey(p.tokens[prefix].Value)
	return len(key) >= 3 && len(key) <= 8
}

func (p *parser) numberIsAudioChannelComponent(index int) bool {
	number, _, ok := cleanNumber(p.tokens[index].Value)
	if !ok || len(number) > 2 {
		return false
	}
	left := p.leftNumberSeparatedByDot(index)
	if left == -1 {
		return false
	}
	beforeLeft := p.prevMeaningful(left)
	next := p.nextMeaningful(index)
	if p.tokenHasAudioCodecWithChannelBase(left) {
		return true
	}
	if beforeLeft != -1 && tokenHasTag(p.tokens[beforeLeft], TagAudioTerm) {
		return true
	}
	return next != -1 && tokenHasTag(p.tokens[next], TagAudioTerm)
}

func (p *parser) leftNumberSeparatedByDot(index int) int {
	if index < 2 || p.tokens[index-1].Kind != TokenDelimiter || p.tokens[index-1].Value != "." {
		return -1
	}
	left := index - 2
	for left >= 0 && p.tokens[left].Kind == TokenDelimiter && (p.tokens[left].Value == " " || p.tokens[left].Value == "_") {
		left--
	}
	if left < 0 || !p.tokens[left].isTextLike() {
		return -1
	}
	if p.tokens[index].GroupID >= 0 && p.tokens[left].GroupID != p.tokens[index].GroupID {
		return -1
	}
	if leftNumber, _, ok := cleanNumber(p.tokens[left].Value); !ok || len(leftNumber) > 2 {
		return -1
	}
	return left
}

// hasAudioCodecInChannelChain walks backwards from index past channel-chain
// tokens (numbers, dots, slashes) to find a leading audio codec (AAC, FLAC, etc.).
// This handles compound specs like "AAC 5.1/2.0" where the channel base "2"
// is not immediately preceded by the audio codec.
func (p *parser) hasAudioCodecInChannelChain(index int) bool {
	for i := index - 1; i >= 0; i-- {
		tok := p.tokens[i]
		if tok.Kind == TokenDelimiter && tok.Value == "." {
			continue
		}
		if tok.isContextDelimiter() && tok.Value == "/" {
			continue
		}
		if tok.Kind == TokenDelimiter {
			continue
		}
		if tok.isTextLike() {
			num, _, ok := cleanNumber(tok.Value)
			if ok && len(num) <= 2 {
				continue
			}
			if tokenHasTag(tok, TagAudioTerm) {
				return true
			}
		}
		return false
	}
	return false
}

func (p *parser) tokenHasAudioCodecWithChannelBase(index int) bool {
	value := p.tokens[index].Value
	if len(value) < 2 {
		return false
	}
	last := value[len(value)-1]
	if last < '0' || last > '9' {
		return false
	}
	key := normalizeKey(value[:len(value)-1])
	switch key {
	case "AAC", "AC3", "DDP", "DTS", "EAC3", "FLAC", "OPUS", "PCM", "TRUEHD":
		return true
	default:
		return false
	}
}

func (p *parser) numberIsHyphenatedTitleQualifier(index int) bool {
	prev := p.prevMeaningful(index)
	if prev == -1 || !p.tokens[prev].isContextDelimiter() || p.tokens[prev].Value != "-" {
		return false
	}
	before := p.prevMeaningful(prev)
	if before == -1 || !p.tokens[before].isTextLike() {
		return false
	}
	if p.tokens[index].GroupID >= 0 && p.tokens[before].GroupID != p.tokens[index].GroupID {
		return false
	}
	switch normalizeWord(p.tokens[before].Value) {
	case "U":
		return true
	default:
		return false
	}
}

func (p *parser) numberIsTvSeasonLabel(index int) bool {
	token := p.tokens[index]
	prev := p.prevMeaningful(index)
	if prev == -1 {
		return false
	}
	if token.GroupID >= 0 && p.tokens[prev].GroupID != token.GroupID {
		return false
	}
	if isTVSeasonPrefixToken(p.tokens[prev]) {
		return true
	}
	if !p.tokens[prev].isContextDelimiter() {
		return false
	}
	beforeDelimiter := p.prevMeaningful(prev)
	if beforeDelimiter == -1 {
		return false
	}
	if token.GroupID >= 0 && p.tokens[beforeDelimiter].GroupID != token.GroupID {
		return false
	}
	return isTVSeasonPrefixToken(p.tokens[beforeDelimiter])
}

func isTVSeasonPrefixToken(token *Token) bool {
	if token == nil || !token.isTextLike() {
		return false
	}
	switch normalizeWord(token.Value) {
	case "TV", "ТВ":
		return true
	default:
		return false
	}
}

func (p *parser) bracketGroupHasMetadata(openIndex int) bool {
	for _, group := range p.groups {
		if group.Open != openIndex {
			continue
		}
		return p.groupHasMetadata(group)
	}
	return false
}

func isAnimeTypeToken(value string) string {
	switch normalizeWord(value) {
	case "SP", "OVA", "OAD", "ONA", "OP", "ED", "NCOP", "NCED":
		return canonicalAnimeType(value)
	default:
		return ""
	}
}

func canonicalAnimeType(value string) string {
	switch normalizeWord(value) {
	case "NCOP":
		return "NCOP"
	case "NCED":
		return "NCED"
	case "OP":
		return "OP"
	case "ED":
		return "ED"
	case "OVA":
		return "OVA"
	case "OAD":
		return "OAD"
	case "ONA":
		return "ONA"
	case "SP":
		return "SP"
	default:
		return value
	}
}

func canonicalReleaseInformation(value string) string {
	switch normalizeKey(value) {
	case "REPACK":
		return "REPACK"
	case "PROPER":
		return "PROPER"
	default:
		return value
	}
}

func (p *parser) hasDotDelimiterBefore(index int) bool {
	for i := index - 1; i >= 0; i-- {
		if p.tokens[i].Kind == TokenDelimiter {
			if p.tokens[i].Value == "." {
				return true
			}
			if p.tokens[i].Value == " " || p.tokens[i].Value == "_" {
				continue
			}
		}
		return false
	}
	return false
}

func (p *parser) hasDotDelimitedMetadataAfter(index int) bool {
	seenDot := false
	for i := index + 1; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if token.Kind == TokenDelimiter {
			if token.Value == "." {
				seenDot = true
			}
			if token.Value == " " || token.Value == "_" {
				continue
			}
			if token.Value != "." {
				return false
			}
			continue
		}
		if !seenDot {
			return false
		}
		if token.Kind == TokenOpenBracket {
			return true
		}
		return tokenHasHardMetadata(token)
	}
	return seenDot
}

func (p *parser) isIsolatedGroupNumber(index int) bool {
	token := p.tokens[index]
	if token.GroupID < 0 {
		return false
	}
	group := p.groups[token.GroupID]
	for i := group.Open + 1; i < group.Close; i++ {
		if i == index || p.tokens[i].isDelimiter() {
			continue
		}
		return false
	}
	return true
}

func (p *parser) episodeTagForToken(index int) Tag {
	if p.tokenIsInEpisodeRange(index) {
		return TagEpisode
	}
	if p.options.ParseAlternatives && p.hasMainEpisodeSignalBefore(index) {
		if p.tokens[index].GroupID >= 0 || p.sequenceTokenBelongsToPipeAlternate(index) {
			return TagEpisodeAlt
		}
	}
	return TagEpisode
}

func (p *parser) tokenIsInEpisodeRange(index int) bool {
	for _, neighbor := range []int{p.prevMeaningful(index), p.nextMeaningful(index)} {
		if neighbor == -1 {
			continue
		}
		if !p.tokens[neighbor].isContextDelimiter() {
			continue
		}
		other := p.prevMeaningful(neighbor)
		if other == index {
			other = p.nextMeaningful(neighbor)
		}
		other = p.rangeEndpointAcrossRepeatedConnectors(index, other)
		if other == -1 || !p.tokens[other].isTextLike() {
			continue
		}
		if p.tokens[index].GroupID >= 0 && p.tokens[other].GroupID != p.tokens[index].GroupID {
			continue
		}
		if p.tokenIsInYearRange(index) || p.tokenIsInYearRange(other) {
			continue
		}
		if _, _, ok := cleanNumber(p.tokens[other].Value); ok {
			return true
		}
	}
	return false
}

func (p *parser) rangeEndpointAcrossRepeatedConnectors(index int, other int) int {
	for other != -1 && p.tokens[other].isContextDelimiter() && isRangeConnector(p.tokens[other].Value) {
		if other < index {
			other = p.prevMeaningful(other)
		} else {
			other = p.nextMeaningful(other)
		}
	}
	return other
}

func (p *parser) hasMainEpisodeBefore(index int) bool {
	for i := 0; i < index; i++ {
		if tokenHasTag(p.tokens[i], TagEpisode) {
			return true
		}
	}
	return false
}

func (p *parser) hasMainEpisode() bool {
	for _, token := range p.tokens {
		if tokenHasTag(token, TagEpisode) {
			return true
		}
	}
	return false
}

func (p *parser) isTrailingReleaseGroupToken(index int) bool {
	for i := len(p.tokens) - 1; i >= 0; i-- {
		token := p.tokens[i]
		if token.isContextDelimiter() && token.Value == "-" {
			before := p.prevMeaningful(i)
			if before != -1 && p.tokenIsTechnicalReleaseBoundary(before) && p.hasTechnicalMetadataBefore(i) {
				start := p.nextMeaningful(i)
				if start != -1 {
					end := p.releaseGroupSegmentEnd(start)
					if index >= start && index <= end {
						return true
					}
				}
			}
			break
		}
	}
	return false
}

func (p *parser) numberIsAudioChannelBase(index int) bool {
	number, _, ok := cleanNumber(p.tokens[index].Value)
	if !ok || len(number) > 2 {
		return false
	}
	nextDot := index + 1
	if nextDot >= len(p.tokens) || p.tokens[nextDot].Kind != TokenDelimiter || p.tokens[nextDot].Value != "." {
		return false
	}
	right := nextDot + 1
	if right >= len(p.tokens) || !p.tokens[right].isTextLike() {
		return false
	}
	rightNumber, _, ok := cleanNumber(p.tokens[right].Value)
	if !ok || len(rightNumber) != 1 || rightNumber[0] < '0' || rightNumber[0] > '2' {
		return false
	}

	prev := p.prevMeaningful(index)
	nextAfterRight := p.nextMeaningful(right)

	if prev != -1 && tokenHasTag(p.tokens[prev], TagAudioTerm) {
		return true
	}
	if p.tokenHasAudioCodecWithChannelBase(index) {
		return true
	}
	if nextAfterRight != -1 && tokenHasTag(p.tokens[nextAfterRight], TagAudioTerm) {
		return true
	}
	// Walk backwards past compound channel specs like "5.1/2.0" or "7.1/5.1"
	// to find the leading audio codec (e.g. AAC, FLAC, E-AC-3).
	if p.hasAudioCodecInChannelChain(index) {
		return true
	}
	return false
}
