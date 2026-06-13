package miyo
import "strings"

var bracketPairs = map[string]string{
	"(": ")",
	"[": "]",
	"{": "}",
	"「": "」",
	"『": "』",
	"【": "】",
}

var closeToOpenBracket = map[string]string{
	")": "(",
	"]": "[",
	"}": "{",
	"」": "「",
	"』": "『",
	"】": "【",
}

var knownFileExtensions = map[string]struct{}{
	"3gp": {}, "avi": {}, "divx": {}, "flv": {}, "m2ts": {}, "mkv": {},
	"mov": {}, "mp4": {}, "mpg": {}, "ogm": {}, "rm": {}, "rmvb": {},
	"ts": {}, "webm": {}, "wmv": {},
}

func isOpeningBracket(value string) bool {
	_, ok := bracketPairs[value]
	return ok
}

func isClosingBracket(value string) bool {
	_, ok := closeToOpenBracket[value]
	return ok
}

func matchingClose(open string) string {
	return bracketPairs[open]
}

func isPlainDelimiter(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '_', '.', ',', '。', '、':
		return true
	default:
		return false
	}
}

func isContextDelimiterRune(r rune) bool {
	switch r {
	case '-', '+', '~', '&', '|', '/', '\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2015':
		return true
	default:
		return false
	}
}

func isRangeConnector(value string) bool {
	switch strings.ToLower(value) {
	case "-", "~", "to", "of":
		return true
	default:
		return false
	}
}
