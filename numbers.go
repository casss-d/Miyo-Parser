package miyo

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	checksumPattern               = regexp.MustCompile(`(?i)^[a-f0-9]{8}$`)
	compactSourceResolutionRegexp = regexp.MustCompile(`(?i)^(?:JP)?(BD|DVD|WEB)(\d{3,4}p)$`)
	resolutionHeightRegexp        = regexp.MustCompile(`(?i)^(\d{3,4})([pi])$`)
	resolutionPairRegexp          = regexp.MustCompile(`(?i)^(\d{3,4})[x×](\d{3,4})$`)
	resolutionKRegexp             = regexp.MustCompile(`(?i)^(\d)k$`)
	resolutionAliasRegexp         = regexp.MustCompile(`(?i)^(HD|FHD|UHD)$`)
	sxeRegexp                     = regexp.MustCompile(`(?i)^S(\d{1,2})E(\d{1,5})(?:v(\d+))?$`)
	oneXEpisodeRegexp             = regexp.MustCompile(`(?i)^(\d{1,2})x(\d{2,5})(?:v(\d+))?$`)
	typeEpisodeRegexp             = regexp.MustCompile(`(?i)^(NCOP|NCED|OP|ED|OVA|OAD|ONA|SP)(\d{1,5})([a-z])?$`)
	seasonTypeEpisodeRegexp       = regexp.MustCompile(`(?i)^S(\d{1,2})(OVA|OAD|ONA|SP)(\d{1,5})([a-z])?$`)
	seasonRegexp                  = regexp.MustCompile(`(?i)^S(\d{1,2})$`)
	volumeRegexp                  = regexp.MustCompile(`(?i)^vol(?:ume)?(\d{1,4})$`)
	partialEpisodeRegexp          = regexp.MustCompile(`(?i)^\d{1,4}[A-C]$`)
	japaneseSeasonRegexp          = regexp.MustCompile(`^第?(\d{1,2})[期季]$`)
	japaneseEpisodeRegexp         = regexp.MustCompile(`^第?(\d{1,5})[話话集]$`)
	releaseInfoVersionRegexp      = regexp.MustCompile(`(?i)^(REPACK|PROPER)v?(\d+)$`)
	releaseVersionRegexp          = regexp.MustCompile(`(?i)^v(\d+)$`)
	ordinalNumberRegexp           = regexp.MustCompile(`(?i)^(\d+)(st|nd|rd|th)$`)
	numberVersionRegexp           = regexp.MustCompile(`(?i)^(\d{1,5}(?:\.\d+)?)(?:v(\d+))?(?:(TV|BD|DVD|LD|WEB))?$`)

	compactEpisodeRegexp *regexp.Regexp
	hashEpisodeRegexp    *regexp.Regexp
	episodeRegexp        *regexp.Regexp
)

type prefixSpec struct {
	Prefix string
	Score  float64
	Source string
	Target **regexp.Regexp
}

var rawPrefixSpecs = []prefixSpec{
	{Prefix: "EP", Score: 4.4, Source: "compact-episode-prefix", Target: &compactEpisodeRegexp},
	{Prefix: "#", Score: 4.4, Source: "hash-episode", Target: &hashEpisodeRegexp},
	{Prefix: "E", Score: 4.1, Source: "episode-token", Target: &episodeRegexp},
}

func init() {
	for _, spec := range rawPrefixSpecs {
		quoted := regexp.QuoteMeta(spec.Prefix)
		var re *regexp.Regexp
		if spec.Prefix == "#" {
			re = regexp.MustCompile(`^` + quoted + `(\d{1,5})(?:v(\d+))?$`)
		} else {
			re = regexp.MustCompile(`(?i)^` + quoted + `(\d{1,5})(?:v(\d+))?$`)
		}
		*spec.Target = re
	}
}

func cleanNumber(value string) (number string, releaseVersion string, ok bool) {
	matches := numberVersionRegexp.FindStringSubmatch(value)
	if matches == nil {
		return "", "", false
	}
	number = stripLeadingZeros(matches[1])
	if matches[2] != "" {
		releaseVersion = stripLeadingZeros(matches[2])
	}
	return number, releaseVersion, true
}

func stripLeadingZeros(value string) string {
	value = strings.TrimLeft(value, "0")
	if value == "" {
		return "0"
	}
	if strings.HasPrefix(value, ".") {
		return "0" + value
	}
	return value
}

func ordinalNumber(value string) (string, bool) {
	matches := ordinalNumberRegexp.FindStringSubmatch(value)
	if matches == nil {
		return "", false
	}
	return stripLeadingZeros(matches[1]), true
}

func isYear(value string, min int, max int) bool {
	if len(value) != 4 {
		return false
	}
	year, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	return year >= min && year <= max
}

func parseResolution(value string) (VideoResolution, bool) {
	if matches := resolutionAliasRegexp.FindStringSubmatch(value); matches != nil {
		switch normalizeKey(matches[1]) {
		case "HD":
			return VideoResolution{Value: value, Height: 720, ScanMethod: "p"}, true
		case "FHD":
			return VideoResolution{Value: value, Height: 1080, ScanMethod: "p"}, true
		case "UHD":
			return VideoResolution{Value: value, Height: 2160, ScanMethod: "p"}, true
		}
	}
	if matches := resolutionPairRegexp.FindStringSubmatch(value); matches != nil {
		width, _ := strconv.Atoi(matches[1])
		height, _ := strconv.Atoi(matches[2])
		return VideoResolution{Value: value, Width: width, Height: height}, true
	}
	if matches := resolutionHeightRegexp.FindStringSubmatch(value); matches != nil {
		height, _ := strconv.Atoi(matches[1])
		return VideoResolution{Value: value, Height: height, ScanMethod: strings.ToLower(matches[2])}, true
	}
	if matches := resolutionKRegexp.FindStringSubmatch(value); matches != nil {
		width, _ := strconv.Atoi(matches[1])
		if matches[1] == "4" {
			return VideoResolution{Value: value, Width: 3840}, true
		}
		return VideoResolution{Value: value, Width: width * 1000}, true
	}
	return VideoResolution{}, false
}

func isResolutionAlias(value string) bool {
	if matches := resolutionAliasRegexp.FindStringSubmatch(value); matches != nil {
		return true
	}
	return false
}
