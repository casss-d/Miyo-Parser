package miyo

const Developer = "CassD"

type Metadata struct {
	FileName            string            `json:"file_name,omitempty"`
	Title               string            `json:"title,omitempty"`
	FormattedTitle      string            `json:"formatted_title,omitempty"`
	Year                string            `json:"year,omitempty"`
	SeasonNumber        []string          `json:"season_number,omitempty"`
	PartNumber          []string          `json:"part_number,omitempty"`
	VolumeNumber        []string          `json:"volume_number,omitempty"`
	EpisodeNumber       []string          `json:"episode_number,omitempty"`
	EpisodeNumberAlt    []string          `json:"episode_number_alt,omitempty"`
	OtherEpisodeNumber  []string          `json:"other_episode_number,omitempty"`
	EpisodeTitle        string            `json:"episode_title,omitempty"`
	AnimeType           []string          `json:"anime_type,omitempty"`
	AudioTerm           []string          `json:"audio_term,omitempty"`
	AudioLanguage       []string          `json:"audio_language,omitempty"`
	DeviceCompatibility []string          `json:"device_compatibility,omitempty"`
	FileChecksum        string            `json:"file_checksum,omitempty"`
	FileExtension       string            `json:"file_extension,omitempty"`
	FileIndex           string            `json:"file_index,omitempty"`
	Language            []string          `json:"language,omitempty"`
	ReleaseGroup        string            `json:"release_group,omitempty"`
	ReleaseInformation  []string          `json:"release_information,omitempty"`
	ReleaseVersion      []string          `json:"release_version,omitempty"`
	Source              []string          `json:"source,omitempty"`
	Subtitles           []string          `json:"subtitles,omitempty"`
	SubtitleLanguage    []string          `json:"subtitle_language,omitempty"`
	VideoResolution     string            `json:"video_resolution,omitempty"`
	VideoResolutions    []VideoResolution `json:"video_resolutions,omitempty"`
	VideoTerm           []string          `json:"video_term,omitempty"`
	Unknown             []string          `json:"unknown,omitempty"`
	Series              []SeriesInfo      `json:"series,omitempty"`
}

type SeriesInfo struct {
	Title       string           `json:"title,omitempty"`
	Type        string           `json:"type,omitempty"`
	Year        []SequenceNumber `json:"year,omitempty"`
	Season      []SequenceNumber `json:"season,omitempty"`
	Episode     []SequenceNumber `json:"episode,omitempty"`
	Volume      []SequenceNumber `json:"volume,omitempty"`
	ContentType []SequenceNumber `json:"content_type,omitempty"`
}

type SequenceNumber struct {
	Number         string           `json:"number,omitempty"`
	Part           string           `json:"part,omitempty"`
	Title          string           `json:"title,omitempty"`
	ReleaseVersion string           `json:"release_version,omitempty"`
	Alternative    []SequenceNumber `json:"alternative,omitempty"`
	Start          *SequenceNumber  `json:"start,omitempty"`
	End            *SequenceNumber  `json:"end,omitempty"`
}

type VideoResolution struct {
	Value      string `json:"value,omitempty"`
	Width      int    `json:"video_width,omitempty"`
	Height     int    `json:"video_height,omitempty"`
	ScanMethod string `json:"scan_method,omitempty"`
}
