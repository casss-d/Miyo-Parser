package miyo

import "sort"

func (p *parser) collectMetadata() {
	p.metadata.Title = buildText(p.titleTokens)
	if p.metadata.Title != "" {
		p.metadata.FormattedTitle = p.metadata.Title
	}
	if p.releaseGroup != "" {
		p.metadata.ReleaseGroup = p.releaseGroup
	}
	if p.episodeTitle != nil {
		p.metadata.EpisodeTitle = buildText(p.episodeTitle)
	}

	for _, match := range p.matchesInSourceOrder() {
		switch match.Tag {
		case TagYear:
			if p.metadata.Year == "" {
				p.metadata.Year = match.Value
			}
		case TagSeason:
			p.metadata.SeasonNumber = appendUnique(p.metadata.SeasonNumber, match.Value)
		case TagEpisode:
			p.metadata.EpisodeNumber = appendUnique(p.metadata.EpisodeNumber, match.Value)
		case TagEpisodeAlt:
			if !containsString(p.metadata.EpisodeNumber, match.Value) {
				p.metadata.EpisodeNumberAlt = appendUnique(p.metadata.EpisodeNumberAlt, match.Value)
			}
		case TagOtherEpisodeNumber:
			p.metadata.OtherEpisodeNumber = appendUnique(p.metadata.OtherEpisodeNumber, match.Value)
		case TagPart:
			p.metadata.PartNumber = appendUnique(p.metadata.PartNumber, match.Value)
		case TagVolume:
			p.metadata.VolumeNumber = appendUnique(p.metadata.VolumeNumber, match.Value)
		case TagAnimeType:
			p.metadata.AnimeType = appendUnique(p.metadata.AnimeType, match.Value)
		case TagAudioTerm:
			p.metadata.AudioTerm = appendUnique(p.metadata.AudioTerm, match.Value)
		case TagDeviceCompatibility:
			p.metadata.DeviceCompatibility = appendUnique(p.metadata.DeviceCompatibility, match.Value)
		case TagFileChecksum:
			if p.metadata.FileChecksum == "" {
				p.metadata.FileChecksum = match.Value
			}
		case TagFileIndex:
			if p.metadata.FileIndex == "" {
				p.metadata.FileIndex = match.Value
			}
		case TagLanguage:
			p.metadata.Language = appendUnique(p.metadata.Language, match.Value)
			if p.languageMatchHasSubtitleContext(match) {
				p.metadata.SubtitleLanguage = appendUnique(p.metadata.SubtitleLanguage, match.Value)
			}
			if p.languageMatchHasAudioContext(match) {
				p.metadata.AudioLanguage = appendUnique(p.metadata.AudioLanguage, match.Value)
			}
		case TagReleaseInformation:
			p.metadata.ReleaseInformation = appendUnique(p.metadata.ReleaseInformation, match.Value)
		case TagReleaseVersion:
			p.metadata.ReleaseVersion = appendUnique(p.metadata.ReleaseVersion, match.Value)
		case TagSource:
			p.metadata.Source = appendUnique(p.metadata.Source, match.Value)
		case TagSubtitles:
			p.metadata.Subtitles = appendUnique(p.metadata.Subtitles, match.Value)
			p.addLanguageFromSubtitleTerm(match.Value)
		case TagVideoResolution:
			if p.metadata.VideoResolution == "" || betterResolution(match.Value, p.metadata.VideoResolution) {
				p.metadata.VideoResolution = match.Value
			}
			if resolution, ok := parseResolution(match.Value); ok {
				p.metadata.VideoResolutions = appendUniqueResolution(p.metadata.VideoResolutions, resolution)
			}
		case TagVideoTerm:
			p.metadata.VideoTerm = appendUnique(p.metadata.VideoTerm, match.Value)
		}
	}

	p.metadata.EpisodeNumberAlt = removeValuesPresentIn(p.metadata.EpisodeNumberAlt, p.metadata.EpisodeNumber)
	if p.metadata.Title != "" && p.metadata.Year != "" {
		p.metadata.FormattedTitle = p.metadata.Title + " (" + p.metadata.Year + ")"
	}
	p.buildSeriesInfo()
}

func (p *parser) matchHasResolvedTitleToken(match termMatch) bool {
	for i := match.TokenFrom; i <= match.TokenTo && i < len(p.tokens); i++ {
		if p.tokens[i].isDelimiter() || p.tokens[i].isBracket() {
			continue
		}
		if p.tokens[i].Category == TagTitle {
			return true
		}
	}
	return false
}

func (p *parser) languageMatchHasSubtitleContext(match termMatch) bool {
	for i := match.TokenFrom; i <= match.TokenTo && i < len(p.tokens); i++ {
		if p.tokens[i].isTextLike() && p.languageTokenHasSubtitleContext(i) {
			return true
		}
	}
	return false
}

func (p *parser) languageMatchHasAudioContext(match termMatch) bool {
	for i := match.TokenFrom; i <= match.TokenTo && i < len(p.tokens); i++ {
		if p.tokens[i].isTextLike() && p.languageTokenHasAudioContext(i) {
			return true
		}
	}
	return false
}

func (p *parser) matchesInSourceOrder() []termMatch {
	matches := append([]termMatch(nil), p.matches...)
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].TokenFrom != matches[j].TokenFrom {
			return matches[i].TokenFrom < matches[j].TokenFrom
		}
		if matches[i].TokenTo != matches[j].TokenTo {
			return matches[i].TokenTo < matches[j].TokenTo
		}
		return matches[i].Tag < matches[j].Tag
	})
	return matches
}

func (p *parser) addLanguageFromSubtitleTerm(value string) {
	switch normalizeKey(value) {
	case "ENGLISHSUB", "ENGLISHSUBS":
		p.metadata.Language = appendUnique(p.metadata.Language, "English")
		p.metadata.SubtitleLanguage = appendUnique(p.metadata.SubtitleLanguage, "English")
	case "ENGSUB", "ENGSUBS":
		p.metadata.Language = appendUnique(p.metadata.Language, "Eng")
		p.metadata.SubtitleLanguage = appendUnique(p.metadata.SubtitleLanguage, "Eng")
	}
}

func resolutionPriority(value string) int {
	resolution, ok := parseResolution(value)
	if !ok {
		return 0
	}
	if resolution.Width > 0 && resolution.Height > 0 {
		return 3
	}
	if resolution.Height > 0 {
		return 2
	}
	if resolution.Width > 0 {
		return 1
	}
	return 0
}

func betterResolution(candidate string, current string) bool {
	candidateResolution, candidateOK := parseResolution(candidate)
	currentResolution, currentOK := parseResolution(current)
	if !candidateOK {
		return false
	}
	if !currentOK {
		return true
	}
	candidatePriority := resolutionPriority(candidate)
	currentPriority := resolutionPriority(current)
	if candidatePriority != currentPriority {
		return candidatePriority > currentPriority
	}
	candidatePixels := resolutionPixels(candidateResolution)
	currentPixels := resolutionPixels(currentResolution)
	if candidatePixels != currentPixels {
		return candidatePixels > currentPixels
	}
	// When aliases (HD/FHD/UHD) and numeric (720p/1080p/2160p) or
	// K-notation (4K) resolve to the same pixel count, prefer the
	// more concrete form over the alias.
	if isResolutionAlias(candidate) && !isResolutionAlias(current) {
		return false
	}
	if !isResolutionAlias(candidate) && isResolutionAlias(current) {
		return true
	}
	return false
}

func resolutionPixels(resolution VideoResolution) int {
	if resolution.Width > 0 && resolution.Height > 0 {
		return resolution.Width * resolution.Height
	}
	if resolution.Height > 0 {
		return resolution.Height
	}
	return resolution.Width
}

func (p *parser) buildSeriesInfo() {
	if p.metadata.Title == "" {
		return
	}
	series := SeriesInfo{Title: p.metadata.Title}
	if p.metadata.Year != "" {
		series.Year = append(series.Year, SequenceNumber{Number: p.metadata.Year})
	}
	for _, season := range p.metadata.SeasonNumber {
		series.Season = append(series.Season, SequenceNumber{Number: season})
	}
	for _, episode := range p.metadata.EpisodeNumber {
		entry := SequenceNumber{Number: episode}
		if p.metadata.EpisodeTitle != "" {
			entry.Title = p.metadata.EpisodeTitle
		}
		for _, alternative := range p.metadata.EpisodeNumberAlt {
			entry.Alternative = append(entry.Alternative, SequenceNumber{Number: alternative})
		}
		if len(p.metadata.ReleaseVersion) == 1 {
			entry.ReleaseVersion = p.metadata.ReleaseVersion[0]
		}
		series.Episode = append(series.Episode, entry)
	}
	if rangeEntry, ok := p.episodeRangeEntry(); ok {
		series.Episode = []SequenceNumber{rangeEntry}
	}
	for _, volume := range p.metadata.VolumeNumber {
		series.Volume = append(series.Volume, SequenceNumber{Number: volume})
	}
	if len(p.metadata.AnimeType) > 0 {
		series.Type = p.metadata.AnimeType[0]
	}
	p.metadata.Series = append(p.metadata.Series, series)
}

func (p *parser) episodeRangeEntry() (SequenceNumber, bool) {
	episodeMatches := make([]termMatch, 0, 2)
	for _, match := range p.matches {
		if match.Tag == TagEpisode {
			episodeMatches = append(episodeMatches, match)
		}
	}
	if len(episodeMatches) != 2 {
		return SequenceNumber{}, false
	}
	first := episodeMatches[0]
	second := episodeMatches[1]
	for i := first.TokenTo + 1; i < second.TokenFrom; i++ {
		if !p.tokens[i].isContextDelimiter() {
			continue
		}
		if p.tokens[i].Value != "-" && p.tokens[i].Value != "~" {
			return SequenceNumber{}, false
		}
		return SequenceNumber{
			Start: &SequenceNumber{Number: first.Value},
			End:   &SequenceNumber{Number: second.Value},
		}, true
	}
	return SequenceNumber{}, false
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	if containsString(values, value) {
		return values
	}
	return append(values, value)
}

func containsString(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

func removeValuesPresentIn(values []string, blocked []string) []string {
	if len(values) == 0 || len(blocked) == 0 {
		return values
	}
	result := values[:0]
	for _, value := range values {
		if containsString(blocked, value) {
			continue
		}
		result = append(result, value)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func appendUniqueResolution(values []VideoResolution, value VideoResolution) []VideoResolution {
	for _, existing := range values {
		if existing.Value == value.Value {
			return values
		}
	}
	return append(values, value)
}
