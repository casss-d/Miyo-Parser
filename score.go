package miyo

var tagTieBreakOrder = map[Tag]int{
	TagBracket:             100,
	TagDelimiter:           100,
	TagContextDelimiter:    100,
	TagFileChecksum:        95,
	TagFileIndex:           92,
	TagVideoResolution:     90,
	TagEpisode:             85,
	TagEpisodeAlt:          84,
	TagSeason:              83,
	TagYear:                82,
	TagOtherEpisodeNumber:  81,
	TagPart:                80,
	TagVolume:              79,
	TagSequenceRange:       78,
	TagSequencePrefix:      77,
	TagAnimeType:           76,
	TagReleaseInformation:  75,
	TagSource:              74,
	TagVideoTerm:           73,
	TagAudioTerm:           72,
	TagDeviceCompatibility: 71,
	TagLanguage:            70,
	TagSubtitles:           69,
	TagReleaseVersion:      68,
	TagReleaseGroup:        67,
	TagFileExtension:       66,
	TagTitle:               50,
	TagEpisodeTitle:        49,
	TagUnknown:             0,
}

func resolveTokens(tokens []*Token) {
	for _, token := range tokens {
		if len(token.Possibilities) == 0 {
			token.Category = TagUnknown
			continue
		}
		bestSet := false
		var best Possibility
		for _, possibility := range token.Possibilities {
			if !bestSet || possibility.Score > best.Score || (possibility.Score == best.Score && tieBreak(possibility.Tag, best.Tag)) {
				best = possibility
				bestSet = true
			}
		}
		token.Category = best.Tag
	}
}

func tieBreakRank(tag Tag) (int, bool) {
	rank, ok := tagTieBreakOrder[tag]
	return rank, ok
}

func tieBreak(candidate Tag, current Tag) bool {
	candidateRank, candidateKnown := tieBreakRank(candidate)
	currentRank, currentKnown := tieBreakRank(current)
	if candidateRank != currentRank {
		return candidateRank > currentRank
	}
	if candidateKnown != currentKnown {
		return candidateKnown
	}
	return candidate < current
}
