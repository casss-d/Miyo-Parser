package miyo
import (
	"path/filepath"
	"regexp"
	"strings"
)

type parser struct {
	filename     string
	baseName     string
	options      Options
	tokens       []*Token
	groups       []BracketGroup
	lexicon      map[string]LexiconEntry
	matches      []termMatch
	releaseGroup string
	titleTokens  []*Token
	episodeTitle []*Token
	metadata     *Metadata
}

func Parse(filename string) *Metadata {
	return ParseWithOptions(filename, Options{})
}

func ParseWithOptions(filename string, options Options) *Metadata {
	meta, _ := ParseDebug(filename, options)
	return meta
}

func ParsePath(path string, options Options) *Metadata {
	options = withDefaults(options)
	if options.Folder == "" {
		options.Folder = filepath.Dir(path)
	}
	return ParseWithOptions(filepath.Base(path), options)
}

func ParseDebug(filename string, options Options) (*Metadata, *DebugResult) {
	p := newParser(filename, options)
	p.parse()
	return p.metadata, newDebugResult(p.tokens, p.matches)
}

func newParser(filename string, options Options) *parser {
	options = withDefaults(options)
	base, extension := splitFileExtension(filename, options.ParseFileExtension)
	meta := &Metadata{
		FileName:       filename,
		FileExtension:  extension,
		FormattedTitle: "",
	}
	return &parser{
		filename: filename,
		baseName: base,
		options:  options,
		lexicon:  defaultLexicon,
		metadata: meta,
	}
}

func (p *parser) parse() {
	p.tokens = tokenize(p.baseName)
	p.groups = analyzeBrackets(p.tokens)

	if p.options.ParseFileChecksum {
		p.scanChecksums()
	}
	if p.options.ParseFileExtension {
		p.scanFileExtensions()
	}
	p.scanFileIndex()
	if p.options.ParseVideoResolution {
		p.scanResolutions()
	}
	if p.options.ParseKeywords {
		p.scanLexicon()
		p.scanBroadcastTags()
		p.scanEmbeddedLanguages()
		p.scanEmbeddedAudioTerms()
		p.scanStandaloneDualAudioTerms()
		p.scanEmbeddedVideoTerms()
		p.scanEmbeddedSubtitleTerms()
	}
	if p.options.ParseYear {
		p.scanYears()
	}
	if p.options.ParseSeason || p.options.ParseEpisode {
		p.scanSequences()
	}

	resolveTokens(p.tokens)
	p.filterMatches()

	if p.options.ParseReleaseGroup {
		p.detectReleaseGroup()
	}
	if p.options.ParseTitle {
		p.detectTitle()
	}
	if p.options.ParseEpisodeTitle {
		p.detectEpisodeTitle()
	}

	resolveTokens(p.tokens)
	p.filterMatches()
	p.collectMetadata()
}

func splitFileExtension(filename string, enabled bool) (string, string) {
	if !enabled {
		return filename, ""
	}
	idx := strings.LastIndex(filename, ".")
	if idx < 0 || idx == len(filename)-1 {
		return filename, ""
	}
	extension := strings.ToLower(filename[idx+1:])
	if _, ok := knownFileExtensions[extension]; !ok {
		return filename, ""
	}
	return filename[:idx], extension
}

func (p *parser) addMatch(tag Tag, value string, from int, to int, score float64, source string) {
	if from < 0 || to < from || to >= len(p.tokens) {
		return
	}
	match := termMatch{
		Tag:       tag,
		Value:     value,
		TokenFrom: from,
		TokenTo:   to,
		Score:     score,
		Source:    source,
	}
	p.matches = append(p.matches, match)
	for i := from; i <= to; i++ {
		if p.tokens[i].isBracket() {
			continue
		}
		tokenScore := score
		if p.tokens[i].isDelimiter() {
			tokenScore -= 0.5
		}
		p.tokens[i].addPossibility(tag, tokenScore, value, source)
	}
}

func (p *parser) filterMatches() {
	active := p.matches[:0]
	for _, match := range p.matches {
		if p.matchHasResolvedTitleToken(match) {
			continue
		}
		active = append(active, match)
	}
	p.matches = active
}

func (p *parser) prevMeaningful(index int) int {
	for i := index - 1; i >= 0; i-- {
		if p.tokens[i].Kind == TokenDelimiter {
			continue
		}
		return i
	}
	return -1
}

func (p *parser) nextMeaningful(index int) int {
	for i := index + 1; i < len(p.tokens); i++ {
		if p.tokens[i].Kind == TokenDelimiter {
			continue
		}
		return i
	}
	return -1
}

func (p *parser) groupsByOpen() map[int]BracketGroup {
	result := make(map[int]BracketGroup, len(p.groups))
	for _, group := range p.groups {
		result[group.Open] = group
	}
	return result
}

func tokenHasTag(token *Token, tag Tag) bool {
	_, ok := token.possibility(tag)
	return ok || token.Category == tag
}

func tokenHasHardMetadata(token *Token) bool {
	if token.Category.isHardMetadata() {
		return true
	}
	for _, possibility := range token.Possibilities {
		if possibility.Tag.isHardMetadata() {
			return true
		}
	}
	return false
}

var broadcastTagRegexp = regexp.MustCompile(`^[★]?(?:\d{1,2}月)?新番[★]?$`)

// cjkBroadcastSuffixes are substrings that mark a CJK token as a broadcast
// season label (e.g. "哆啦A梦新番" → "新番" suffix). They are only matched when
// they appear at the end of the token value so that normal anime words are not
// accidentally flagged.
var cjkBroadcastSuffixes = []string{"新番", "週新番", "周新番", "月新番"}

func (p *parser) scanBroadcastTags() {
	for i, token := range p.tokens {
		if !token.isTextLike() || tokenHasHardMetadata(token) {
			continue
		}
		if broadcastTagRegexp.MatchString(token.Value) {
			p.addMatch(TagReleaseInformation, token.Value, i, i, 4.5, "broadcast-tag")
			continue
		}
		// Detect embedded CJK broadcast suffixes (e.g. "哆啦A梦新番").
		// We only flag the suffix itself; the prefix becomes an additional
		// synthetic title token trimmed by stripEmbeddedBroadcastSuffix later.
		for _, suffix := range cjkBroadcastSuffixes {
			if strings.HasSuffix(token.Value, suffix) && len(token.Value) > len(suffix) {
				// Rewrite: trim the suffix so it doesn't appear in the title.
				token.Value = token.Value[:len(token.Value)-len(suffix)]
				break
			}
		}
	}
}

func (p *parser) scanFileExtensions() {
	for i, token := range p.tokens {
		if !token.isTextLike() || tokenHasHardMetadata(token) {
			continue
		}
		if _, ok := knownFileExtensions[strings.ToLower(token.Value)]; ok {
			p.addMatch(TagFileExtension, strings.ToLower(token.Value), i, i, 5.0, "file-extension")
		}
	}
}
