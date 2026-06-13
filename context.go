package miyo
// contextualSupportScore returns a numeric score indicating how much the context
// supports the token span [start, end] being interpreted as metadata rather than title text.
func (p *parser) contextualSupportScore(start int, end int) float64 {
	score := 0.0

	// 1. Inside leading release group -> strong negative metadata support for normal elements
	if p.tokenIsInsideLeadingReleaseGroup(start) {
		score -= 3.0
	}

	// 2. Has hard metadata neighbor in the same group -> strong positive
	if p.hasHardMetadataNeighbor(start, end) {
		score += 2.0
	}

	// 3. Part of a recognized technical stack -> strong positive
	if p.isPartOfTechnicalStack(start, end) {
		score += 1.5
	}

	// 4. Isolated in its own bracket group -> positive
	if p.isIsolatedInBracketGroup(start, end) {
		score += 1.5
	}

	// 5. Inside a bracket group that contains other hard metadata -> positive
	if p.groupHasOtherHardMetadata(start, end) {
		score += 1.5
	}

	// 6. Has title-looking neighbors -> negative
	if p.hasTitleLookingNeighbors(start, end) {
		score -= 1.5
	}

	return score
}

// tokenIsHardMetadata checks if a token is marked as hard metadata or has a hard metadata possibility/lexicon entry.
func (p *parser) tokenIsHardMetadata(index int) bool {
	if index < 0 || index >= len(p.tokens) {
		return false
	}
	token := p.tokens[index]
	if token.Category.isHardMetadata() {
		return true
	}
	if entry, ok := p.lexicon[normalizeKey(token.Value)]; ok && entry.Tag.isHardMetadata() {
		return true
	}
	for _, possibility := range token.Possibilities {
		if possibility.Tag.isHardMetadata() {
			return true
		}
	}
	val := token.Value
	if isYear(val, p.options.YearMin, p.options.YearMax) {
		return true
	}
	if seasonRegexp.MatchString(val) || seasonTypeEpisodeRegexp.MatchString(val) || typeEpisodeRegexp.MatchString(val) {
		return true
	}
	return false
}

// hasHardMetadataNeighbor checks if the span has a hard metadata neighbor in the same bracket group.
func (p *parser) hasHardMetadataNeighbor(start int, end int) bool {
	groupID := p.tokens[start].GroupID
	for _, neighbor := range []int{p.prevMeaningful(start), p.nextMeaningful(end)} {
		if neighbor == -1 {
			continue
		}
		if groupID >= 0 && p.tokens[neighbor].GroupID != groupID {
			continue
		}
		if p.tokenIsHardMetadata(neighbor) {
			return true
		}
	}
	return false
}

// isPartOfTechnicalStack checks if the span is adjacent to other technical metadata or in a dot-delimited tech stack.
func (p *parser) isPartOfTechnicalStack(start int, end int) bool {
	groupID := p.tokens[start].GroupID
	// Adjacent technical metadata
	for _, neighbor := range []int{p.prevMeaningful(start), p.nextMeaningful(end)} {
		if neighbor == -1 {
			continue
		}
		if groupID >= 0 && p.tokens[neighbor].GroupID != groupID {
			continue
		}
		if p.tokenIsHardMetadata(neighbor) || isStandaloneTechnicalModeMarker(p.tokens[neighbor]) {
			return true
		}
	}
	// Dot-delimited metadata chain
	if p.hasDotDelimiterBefore(start) && p.hasDotDelimitedMetadataAfter(end) {
		return true
	}
	return false
}

// isIsolatedInBracketGroup checks if the span is the sole non-delimiter text in its bracket group.
func (p *parser) isIsolatedInBracketGroup(start int, end int) bool {
	token := p.tokens[start]
	if token.GroupID < 0 {
		return false
	}
	group := p.groups[token.GroupID]
	for i := group.Open + 1; i < group.Close; i++ {
		if (i >= start && i <= end) || p.tokens[i].isDelimiter() {
			continue
		}
		return false
	}
	return true
}

// hasTitleLookingNeighbors checks if neighbor text tokens look like ordinary title text.
func (p *parser) hasTitleLookingNeighbors(start int, end int) bool {
	prev := p.prevMeaningful(start)
	next := p.nextMeaningful(end)

	titleCount := 0
	for _, neighbor := range []int{prev, next} {
		if neighbor == -1 {
			continue
		}
		if p.tokens[neighbor].GroupID >= 0 {
			continue
		}
		tok := p.tokens[neighbor]
		if tok.isTextLike() && !p.tokenIsHardMetadata(neighbor) {
			titleCount++
		}
	}
	return titleCount > 0
}

func (p *parser) groupHasOtherHardMetadata(start int, end int) bool {
	token := p.tokens[start]
	if token.GroupID < 0 {
		return false
	}
	group := p.groups[token.GroupID]
	for i := group.Open + 1; i < group.Close; i++ {
		if i >= start && i <= end {
			continue
		}
		if p.tokenIsHardMetadata(i) {
			return true
		}
	}
	return false
}
