package miyo

type Tag string

const (
	TagUnknown             Tag = "unknown"
	TagDelimiter           Tag = "delimiter"
	TagContextDelimiter    Tag = "context_delimiter"
	TagBracket             Tag = "bracket"
	TagTitle               Tag = "title"
	TagEpisodeTitle        Tag = "episode_title"
	TagReleaseGroup        Tag = "release_group"
	TagYear                Tag = "year"
	TagSeason              Tag = "season"
	TagEpisode             Tag = "episode"
	TagEpisodeAlt          Tag = "episode_alt"
	TagOtherEpisodeNumber  Tag = "other_episode_number"
	TagPart                Tag = "part"
	TagVolume              Tag = "volume"
	TagSequencePrefix      Tag = "sequence_prefix"
	TagSequenceRange       Tag = "sequence_range"
	TagAnimeType           Tag = "anime_type"
	TagAudioTerm           Tag = "audio_term"
	TagDeviceCompatibility Tag = "device_compatibility"
	TagFileChecksum        Tag = "file_checksum"
	TagFileExtension       Tag = "file_extension"
	TagFileIndex           Tag = "file_index"
	TagLanguage            Tag = "language"
	TagReleaseInformation  Tag = "release_information"
	TagReleaseVersion      Tag = "release_version"
	TagSource              Tag = "source"
	TagSubtitles           Tag = "subtitles"
	TagVideoResolution     Tag = "video_resolution"
	TagVideoTerm           Tag = "video_term"
)

func (tag Tag) isHardMetadata() bool {
	switch tag {
	case TagYear, TagSeason, TagEpisode, TagEpisodeAlt, TagOtherEpisodeNumber, TagPart, TagVolume,
		TagAnimeType, TagAudioTerm, TagDeviceCompatibility, TagFileChecksum, TagFileExtension,
		TagFileIndex, TagLanguage, TagReleaseInformation, TagReleaseVersion, TagSource,
		TagSubtitles, TagVideoResolution, TagVideoTerm:
		return true
	default:
		return false
	}
}
