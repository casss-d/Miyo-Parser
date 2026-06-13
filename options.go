package miyo

type Options struct {
	ParseTitle           bool
	ParseEpisode         bool
	ParseEpisodeTitle    bool
	ParseSeason          bool
	ParsePart            bool
	ParseVolume          bool
	ParseYear            bool
	ParseFileExtension   bool
	ParseFileChecksum    bool
	ParseReleaseGroup    bool
	ParseVideoResolution bool
	ParseKeywords        bool
	ParseAlternatives    bool

	// Use Disable* fields to turn off defaults. Plain false Parse* values are
	// indistinguishable from omitted fields in Go struct literals.
	DisableTitle           bool
	DisableEpisode         bool
	DisableEpisodeTitle    bool
	DisableSeason          bool
	DisablePart            bool
	DisableVolume          bool
	DisableYear            bool
	DisableFileExtension   bool
	DisableFileChecksum    bool
	DisableReleaseGroup    bool
	DisableVideoResolution bool
	DisableKeywords        bool
	DisableAlternatives    bool

	ParseFolderContext   bool
	Folder               string
	YearMin              int
	YearMax              int
	Debug                bool
	CompatibilityMode    bool
	DisableCompatibility bool
}

func defaultOptions() Options {
	return Options{
		ParseTitle:           true,
		ParseEpisode:         true,
		ParseEpisodeTitle:    true,
		ParseSeason:          true,
		ParsePart:            true,
		ParseVolume:          true,
		ParseYear:            true,
		ParseFileExtension:   true,
		ParseFileChecksum:    true,
		ParseReleaseGroup:    true,
		ParseVideoResolution: true,
		ParseKeywords:        true,
		ParseAlternatives:    true,
		YearMin:              1900,
		YearMax:              2099,
		CompatibilityMode:    true,
	}
}

func withDefaults(options Options) Options {
	defaults := defaultOptions()

	// A zero Options value should behave like default parsing.
	if options == (Options{}) {
		return defaults
	}

	if options.DisableTitle {
		defaults.ParseTitle = false
	}
	if options.DisableEpisode {
		defaults.ParseEpisode = false
	}
	if options.DisableEpisodeTitle {
		defaults.ParseEpisodeTitle = false
	}
	if options.DisableSeason {
		defaults.ParseSeason = false
	}
	if options.DisablePart {
		defaults.ParsePart = false
	}
	if options.DisableVolume {
		defaults.ParseVolume = false
	}
	if options.DisableYear {
		defaults.ParseYear = false
	}
	if options.DisableFileExtension {
		defaults.ParseFileExtension = false
	}
	if options.DisableFileChecksum {
		defaults.ParseFileChecksum = false
	}
	if options.DisableReleaseGroup {
		defaults.ParseReleaseGroup = false
	}
	if options.DisableVideoResolution {
		defaults.ParseVideoResolution = false
	}
	if options.DisableKeywords {
		defaults.ParseKeywords = false
	}
	if options.DisableAlternatives {
		defaults.ParseAlternatives = false
	}
	if options.ParseFolderContext {
		defaults.ParseFolderContext = true
	}
	if options.Folder != "" {
		defaults.Folder = options.Folder
		defaults.ParseFolderContext = true
	}
	if options.YearMin != 0 {
		defaults.YearMin = options.YearMin
	}
	if options.YearMax != 0 {
		defaults.YearMax = options.YearMax
	}
	if options.Debug {
		defaults.Debug = true
	}
	if options.DisableCompatibility {
		defaults.CompatibilityMode = false
	}

	return defaults
}
