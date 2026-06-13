package miyo

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestParseFansubReleaseWithTechnicalMetadata(t *testing.T) {
	meta := Parse("[TaigaSubs]_Toradora!_(2008)_-_01v2_-_Tiger_and_Dragon_[1280x720_H.264_FLAC][1234ABCD].mkv")

	assertEqual(t, meta.FileName, "[TaigaSubs]_Toradora!_(2008)_-_01v2_-_Tiger_and_Dragon_[1280x720_H.264_FLAC][1234ABCD].mkv")
	assertEqual(t, meta.ReleaseGroup, "TaigaSubs")
	assertEqual(t, meta.Title, "Toradora!")
	assertEqual(t, meta.FormattedTitle, "Toradora! (2008)")
	assertEqual(t, meta.Year, "2008")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"1"})
	assertEqual(t, meta.EpisodeTitle, "Tiger and Dragon")
	assertSliceEqual(t, meta.ReleaseVersion, []string{"2"})
	assertEqual(t, meta.VideoResolution, "1280x720")
	assertSliceContains(t, meta.VideoTerm, "H.264")
	assertSliceContains(t, meta.AudioTerm, "FLAC")
	assertEqual(t, meta.FileChecksum, "1234ABCD")
	assertEqual(t, meta.FileExtension, "mkv")
}

func ExampleParse() {
	meta := Parse("[SubsPlease] Sousou no Frieren - 14 [1080p].mkv")

	fmt.Println(meta.Title)
	fmt.Println(meta.EpisodeNumber[0])
	fmt.Println(meta.VideoResolution)

	// Output:
	// Sousou no Frieren
	// 14
	// 1080p
}

func BenchmarkParseTypicalRelease(b *testing.B) {
	filename := "maboroshi 2023 MULTi AD 1080p NF WEB-DL DDP5.1 DV HDR10 H.265-Tsundere-Raws (VF, FRENCH, SUBFRENCH, VOSTFR)"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Parse(filename)
	}
}

func TestParseTitleKeepsBracketedTitleFragment(t *testing.T) {
	meta := Parse("Evangelion_1.0_You_Are_[Not]_Alone_(1080p)_[@Home]")

	assertEqual(t, meta.Title, "Evangelion 1.0 You Are [Not] Alone")
	assertEqual(t, meta.ReleaseGroup, "@Home")
	assertEqual(t, meta.VideoResolution, "1080p")
	assertSliceEqual(t, meta.EpisodeNumber, nil)
}

func TestParseEpisodePrefixBeforeTitle(t *testing.T) {
	meta := Parse("Episode 14 Ore no Imouto ga Konnani Kawaii Wake ga Nai. (saison 2) VOSTFR")

	assertEqual(t, meta.Title, "Ore no Imouto ga Konnani Kawaii Wake ga Nai.")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"14"})
	assertSliceEqual(t, meta.SeasonNumber, []string{"2"})
	assertSliceContains(t, meta.Language, "VOSTFR")
}

func TestParseTitleNumberWithoutEpisodeSignalStaysTitle(t *testing.T) {
	meta := Parse("The iDOLM@STER 765 Pro to Iu Monogatari.mkv")

	assertEqual(t, meta.Title, "The iDOLM@STER 765 Pro to Iu Monogatari")
	assertSliceEqual(t, meta.EpisodeNumber, nil)
	assertEqual(t, meta.FileExtension, "mkv")
}

func TestParseSeasonEpisodeAlternativeAndEpisodeTitle(t *testing.T) {
	meta := Parse("Fairy Tail - S06E32 - Tartaros Arc Iron Fist of the Fire Dragon [Episode 83]")

	assertEqual(t, meta.Title, "Fairy Tail")
	assertSliceEqual(t, meta.SeasonNumber, []string{"6"})
	assertSliceEqual(t, meta.EpisodeNumber, []string{"32"})
	assertSliceEqual(t, meta.EpisodeNumberAlt, []string{"83"})
	assertEqual(t, meta.EpisodeTitle, "Tartaros Arc Iron Fist of the Fire Dragon")
}

func TestParseLargeBatchMetadata(t *testing.T) {
	meta := Parse("[Anime Time] Sword Art Online (S01+S02+S03+S04+Alternative+Movies+Specials+OVAs) [BD] [Dual Audio][1080p][HEVC 10bit x265][AAC][Eng Sub] [Batch] (SAO)")

	assertEqual(t, meta.ReleaseGroup, "Anime Time")
	assertEqual(t, meta.Title, "Sword Art Online")
	assertSliceEqual(t, meta.SeasonNumber, []string{"1", "2", "3", "4"})
	assertSliceContains(t, meta.AnimeType, "Movies")
	assertSliceContains(t, meta.AnimeType, "Specials")
	assertSliceContains(t, meta.AnimeType, "OVAs")
	assertSliceContains(t, meta.Source, "BD")
	assertSliceContains(t, meta.AudioTerm, "Dual Audio")
	assertSliceContains(t, meta.AudioTerm, "AAC")
	assertSliceContains(t, meta.Language, "Eng")
	assertSliceContains(t, meta.Subtitles, "Eng-Sub")
	assertEqual(t, meta.VideoResolution, "1080p")
	assertSliceContains(t, meta.VideoTerm, "HEVC")
	assertSliceContains(t, meta.VideoTerm, "10bit")
	assertSliceContains(t, meta.VideoTerm, "x265")
	assertSliceContains(t, meta.ReleaseInformation, "Batch")
}

func TestParseBracketedTitleBeforeDotDelimitedEpisode(t *testing.T) {
	meta := Parse("[Keroro].148.[Xvid.mp3].[FE68D5F1].avi")

	assertEqual(t, meta.ReleaseGroup, "")
	assertEqual(t, meta.Title, "Keroro")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"148"})
	assertSliceContains(t, meta.VideoTerm, "Xvid")
	assertSliceContains(t, meta.AudioTerm, "MP3")
	assertEqual(t, meta.FileChecksum, "FE68D5F1")
	assertEqual(t, meta.FileExtension, "avi")
}

func TestParseWithDebugOptionStillUsesDefaultParseFlags(t *testing.T) {
	meta := ParseWithOptions("[SubsPlease] Sousou no Frieren - 14 [1080p].mkv", Options{Debug: true})

	assertEqual(t, meta.ReleaseGroup, "SubsPlease")
	assertEqual(t, meta.Title, "Sousou no Frieren")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"14"})
	assertEqual(t, meta.VideoResolution, "1080p")
}

func TestParseWithSingleParseOptionKeepsDefaultParseFlags(t *testing.T) {
	meta := ParseWithOptions("[SubsPlease] Sousou no Frieren - 14 [1080p][1234ABCD].mkv", Options{ParseFileChecksum: true})

	assertEqual(t, meta.ReleaseGroup, "SubsPlease")
	assertEqual(t, meta.Title, "Sousou no Frieren")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"14"})
	assertEqual(t, meta.VideoResolution, "1080p")
	assertEqual(t, meta.FileChecksum, "1234ABCD")
	assertEqual(t, meta.FileExtension, "mkv")
}

func TestParseWithExplicitDisableOptions(t *testing.T) {
	meta := ParseWithOptions("[SubsPlease] Sousou no Frieren - 14 [1080p].mkv", Options{
		DisableTitle:   true,
		DisableEpisode: true,
	})

	assertEqual(t, meta.ReleaseGroup, "SubsPlease")
	assertEqual(t, meta.Title, "")
	assertSliceEqual(t, meta.EpisodeNumber, nil)
	assertEqual(t, meta.VideoResolution, "1080p")
	assertEqual(t, meta.FileExtension, "mkv")
}

func TestCleanNumberAcceptsFractionalEpisodes(t *testing.T) {
	number, releaseVersion, ok := cleanNumber("07.5v2")

	if !ok {
		t.Fatalf("cleanNumber rejected fractional episode")
	}
	assertEqual(t, number, "7.5")
	assertEqual(t, releaseVersion, "2")

	number, releaseVersion, ok = cleanNumber("00.5")
	if !ok {
		t.Fatalf("cleanNumber rejected leading-zero fractional episode")
	}
	assertEqual(t, number, "0.5")
	assertEqual(t, releaseVersion, "")
}

func TestTokenPossibilityReplacementPreservesHighestScore(t *testing.T) {
	token := newToken("01", 0, TokenText)

	token.addPossibility(TagEpisode, 4, "1", "first")
	token.addPossibility(TagTitle, 3.2, "01", "title")
	token.addPossibility(TagEpisode, 3, "wrong", "lower")
	token.addPossibility(TagEpisode, 4.5, "1", "higher")

	possibility, ok := token.possibility(TagEpisode)
	if !ok {
		t.Fatalf("expected episode possibility")
	}
	assertEqual(t, possibility.Value, "1")
	assertEqual(t, possibility.Score, 4.5)
	resolveTokens([]*Token{token})
	assertEqual(t, token.Category, TagEpisode)
	assertEqual(t, token.resolvedValue(), "1")
}

func TestParseLanguageWordsInTitleWithoutMediaContext(t *testing.T) {
	meta := Parse("English Teacher - 01 [1080p].mkv")

	assertEqual(t, meta.Title, "English Teacher")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"1"})
	assertSliceEqual(t, meta.Language, nil)
	assertSliceEqual(t, meta.SubtitleLanguage, nil)
	assertSliceEqual(t, meta.AudioLanguage, nil)
}

func TestParseSubtitleLanguageContextSeparatelyFromAudio(t *testing.T) {
	meta := Parse("[Group] Anime Title - 01 [Dual Audio][Eng Sub][1080p].mkv")

	assertEqual(t, meta.ReleaseGroup, "Group")
	assertEqual(t, meta.Title, "Anime Title")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"1"})
	assertSliceContains(t, meta.AudioTerm, "Dual Audio")
	assertSliceContains(t, meta.Subtitles, "Eng-Sub")
	assertSliceContains(t, meta.Language, "Eng")
	assertSliceContains(t, meta.SubtitleLanguage, "Eng")
	assertSliceEqual(t, meta.AudioLanguage, nil)
}

func TestParseDubLanguageContextSeparatelyFromSubtitles(t *testing.T) {
	meta := Parse("[Group] Anime Title - 01 [English Dub][1080p].mkv")

	assertEqual(t, meta.Title, "Anime Title")
	assertSliceContains(t, meta.Language, "English")
	assertSliceContains(t, meta.AudioLanguage, "English")
	assertSliceEqual(t, meta.SubtitleLanguage, nil)
}

func TestParseExplicitFiveDigitEpisodeNumber(t *testing.T) {
	meta := Parse("Long Runner EP12345 [1080p].mkv")

	assertEqual(t, meta.Title, "Long Runner")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"12345"})
	assertEqual(t, meta.VideoResolution, "1080p")
}

func TestDatePartsRequireValidMonthAndDay(t *testing.T) {
	invalid := parserForDateTest("Show 2024-99-99 01")
	if invalid.tokenStartsDate(tokenIndexForDateTest(t, invalid.tokens, "2024")) {
		t.Fatalf("2024-99-99 should not be accepted as a date")
	}

	valid := parserForDateTest("Show 2024-12-31 01")
	if !valid.tokenStartsDate(tokenIndexForDateTest(t, valid.tokens, "2024")) {
		t.Fatalf("2024-12-31 should be accepted as a date")
	}
}

func TestTieBreakRankCoversAllTags(t *testing.T) {
	for _, tag := range []Tag{
		TagUnknown, TagDelimiter, TagContextDelimiter, TagBracket, TagTitle, TagEpisodeTitle,
		TagReleaseGroup, TagYear, TagSeason, TagEpisode, TagEpisodeAlt, TagOtherEpisodeNumber,
		TagPart, TagVolume, TagSequencePrefix, TagSequenceRange, TagAnimeType, TagAudioTerm,
		TagDeviceCompatibility, TagFileChecksum, TagFileExtension, TagFileIndex, TagLanguage,
		TagReleaseInformation, TagReleaseVersion, TagSource, TagSubtitles, TagVideoResolution,
		TagVideoTerm,
	} {
		if _, ok := tieBreakRank(tag); !ok {
			t.Fatalf("tie-break rank missing for %s", tag)
		}
	}
}

func TestParseOrdinalSeasonEpisodeAndTrailingReleaseGroup(t *testing.T) {
	meta := Parse("Hayate no Gotoku 2nd Season 24 (Blu-Ray 1080p) [Chihiro]")

	assertEqual(t, meta.Title, "Hayate no Gotoku")
	assertSliceEqual(t, meta.SeasonNumber, []string{"2"})
	assertSliceEqual(t, meta.EpisodeNumber, []string{"24"})
	assertSliceContains(t, meta.Source, "Blu-Ray")
	assertEqual(t, meta.VideoResolution, "1080p")
	assertEqual(t, meta.ReleaseGroup, "Chihiro")
}

func TestParseNoDotTitleNumberDoesNotBecomeEpisode(t *testing.T) {
	meta := Parse("[BluDragon] Blue Submarine No.6 (DVD, R2, Dual Audio) V3")

	assertEqual(t, meta.ReleaseGroup, "BluDragon")
	assertEqual(t, meta.Title, "Blue Submarine No.6")
	assertSliceEqual(t, meta.EpisodeNumber, nil)
	assertSliceContains(t, meta.Source, "DVD")
	assertSliceContains(t, meta.AudioTerm, "Dual Audio")
	assertSliceEqual(t, meta.ReleaseVersion, []string{"3"})
}

func TestParseOneXSeasonEpisodeAndEpisodeTitle(t *testing.T) {
	meta := Parse("After War Gundam X - 1x03 - My Mount is Fierce!.mkv")

	assertEqual(t, meta.Title, "After War Gundam X")
	assertSliceEqual(t, meta.SeasonNumber, []string{"1"})
	assertSliceEqual(t, meta.EpisodeNumber, []string{"3"})
	assertEqual(t, meta.EpisodeTitle, "My Mount is Fierce!")
	assertEqual(t, meta.FileExtension, "mkv")
}

func TestParseEpisodeRangeAndEpisodeList(t *testing.T) {
	rangeMeta := Parse("[HorribleSubs] Tsukimonogatari - (01-04) [1080p].mkv")
	assertEqual(t, rangeMeta.Title, "Tsukimonogatari")
	assertSliceEqual(t, rangeMeta.EpisodeNumber, []string{"1", "4"})
	if len(rangeMeta.Series) == 0 || len(rangeMeta.Series[0].Episode) == 0 ||
		rangeMeta.Series[0].Episode[0].Start == nil || rangeMeta.Series[0].Episode[0].End == nil {
		t.Fatalf("expected structured episode range, got %#v", rangeMeta.Series)
	}
	assertEqual(t, rangeMeta.Series[0].Episode[0].Start.Number, "1")
	assertEqual(t, rangeMeta.Series[0].Episode[0].End.Number, "4")

	listMeta := Parse("[HorribleSubs] Momokuri - 01+02 [720p]")
	assertEqual(t, listMeta.Title, "Momokuri")
	assertSliceEqual(t, listMeta.EpisodeNumber, []string{"1", "2"})
	if len(listMeta.Series) == 0 || len(listMeta.Series[0].Episode) != 2 {
		t.Fatalf("expected two episode entries, got %#v", listMeta.Series)
	}
}

func TestParsePartialAndFractionalEpisodes(t *testing.T) {
	partial := Parse("[HorribleSubs] Gintama - 111C [1080p].mkv")
	assertEqual(t, partial.Title, "Gintama")
	assertSliceEqual(t, partial.EpisodeNumber, []string{"111C"})

	fractional := Parse("[Zurako] Sora no Woto - 07.5 - Drinking Party - Fortress Battle (BD 1080p AAC) [F7DF16F7].mkv")
	assertEqual(t, fractional.Title, "Sora no Woto")
	assertSliceEqual(t, fractional.EpisodeNumber, []string{"07.5"})
	assertEqual(t, fractional.EpisodeTitle, "Drinking Party - Fortress Battle")
}

func TestParseJapaneseSeasonAndEpisodeCounters(t *testing.T) {
	meta := Parse("呪術廻戦 第2期 01話")

	assertEqual(t, meta.Title, "呪術廻戦")
	assertSliceEqual(t, meta.SeasonNumber, []string{"2"})
	assertSliceEqual(t, meta.EpisodeNumber, []string{"1"})
}

func TestParseLeadingFileIndexSeparatelyFromTitleAndEpisode(t *testing.T) {
	meta := Parse("01. Cowboy Bebop - 05.mkv")

	assertEqual(t, meta.FileIndex, "1")
	assertEqual(t, meta.Title, "Cowboy Bebop")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"5"})
	assertEqual(t, meta.FileExtension, "mkv")
}

func TestParseEpisodePrefixOnlyReleaseAsEpisodeTitle(t *testing.T) {
	meta := Parse("Ep. 01 - The Boy in the Iceberg")

	assertEqual(t, meta.Title, "")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"1"})
	assertEqual(t, meta.EpisodeTitle, "The Boy in the Iceberg")
}

func TestParseSlashSeparatedAudioCodecStack(t *testing.T) {
	meta := Parse("[DemiHuman] SPY x FAMILY CODE: White (2023) REPACK (BD Remux 1080p AVC FLAC/E-AC-3/AAC 5.1/2.0) [Dual-Audio] [Multi-Audio] [Multi-Sub] [OV] [DF] [Synchro] [OmU] [OmeU] [RUS(int)] [MOVIE] | 劇場版 Gekijouban SPY×FAMILY CODE: White | Семья шпиона — Код: Белый")

	assertEqual(t, meta.ReleaseGroup, "DemiHuman")
	assertEqual(t, meta.Title, "SPY x FAMILY CODE: White")
	assertSliceContains(t, meta.Source, "BD")
	assertSliceContains(t, meta.ReleaseInformation, "Remux")
	assertEqual(t, meta.VideoResolution, "1080p")
	assertSliceContains(t, meta.VideoTerm, "AVC")
	assertSliceContains(t, meta.AudioTerm, "FLAC")
	assertSliceContains(t, meta.AudioTerm, "E-AC-3")
	assertSliceContains(t, meta.AudioTerm, "AAC")
	assertSliceContains(t, meta.AudioTerm, "Dual Audio")
	assertSliceContains(t, meta.AudioTerm, "Multi Audio")
}

func TestParseAudioCodecBeforeAtmosReleaseGroupSuffix(t *testing.T) {
	meta := Parse("機動戦士ガンダム.Mobile.Suit.Gundam.SEED.Freedom.2024.Movie.BDRip.1080p+4K.2160p.HDR10.x265.FLAC/Atmos-CoolFansSub&Sakurato&Comicat")

	assertEqual(t, meta.Title, "機動戦士ガンダム Mobile Suit Gundam SEED Freedom")
	assertSliceContains(t, meta.AnimeType, "Movie")
	assertSliceContains(t, meta.Source, "BDRip")
	assertSliceContains(t, meta.VideoTerm, "x265")
	assertSliceContains(t, meta.AudioTerm, "FLAC")
	assertEqual(t, meta.ReleaseGroup, "CoolFansSub&Sakurato&Comicat")
}

func TestParseStandaloneDualAudioMarker(t *testing.T) {
	meta := Parse("ALL YOU NEED IS KILL 2025 1080p AMZN WEB-DL DUAL DDP5.1 H.264-SCOPE")

	assertEqual(t, meta.Title, "ALL YOU NEED IS KILL")
	assertSliceContains(t, meta.Source, "AMZN")
	assertSliceContains(t, meta.Source, "WEB-DL")
	assertSliceContains(t, meta.AudioTerm, "Dual Audio")
	assertSliceContains(t, meta.AudioTerm, "DDP")
	assertSliceContains(t, meta.VideoTerm, "H.264")
	assertEqual(t, meta.VideoResolution, "1080p")
	assertEqual(t, meta.ReleaseGroup, "SCOPE")
}

func TestParseStandaloneDualSubsDoesNotBecomeDualAudio(t *testing.T) {
	meta := Parse("[Saiko] Kinnikuman Nisei BATCH 01 (Episodes 1-4) DUAL SUBS: ENGLISH + ESPAÑOL")

	assertEqual(t, meta.ReleaseGroup, "Saiko")
	assertEqual(t, meta.Title, "Kinnikuman Nisei")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"1", "4"})
	assertSliceNotContains(t, meta.AudioTerm, "Dual Audio")
}

func TestParsePipeAlternateSeasonEpisodeAsEpisodeAlt(t *testing.T) {
	meta := Parse("[yaneura] Re:Zero Break Time S4 - 03 (HDTV 1080p) | ReZero S00E73")

	assertEqual(t, meta.ReleaseGroup, "yaneura")
	assertEqual(t, meta.Title, "Re:Zero Break Time")
	assertSliceEqual(t, meta.SeasonNumber, []string{"4"})
	assertSliceEqual(t, meta.EpisodeNumber, []string{"3"})
	assertSliceEqual(t, meta.EpisodeNumberAlt, []string{"73"})
}

func TestParseSeparatedEnglishSubtitleTerms(t *testing.T) {
	engSub := Parse("[Ommex] Doraemon (2005) Episode 913 [ENG SUB][1080p x265 AAC]")
	assertEqual(t, engSub.ReleaseGroup, "Ommex")
	assertEqual(t, engSub.Title, "Doraemon")
	assertSliceEqual(t, engSub.EpisodeNumber, []string{"913"})
	assertSliceContains(t, engSub.Subtitles, "Eng-Sub")
	assertSliceContains(t, engSub.Language, "Eng")
	assertSliceContains(t, engSub.SubtitleLanguage, "Eng")

	engHyphenSub := Parse("[KawaSubs] Journal with Witch - S01E05 [WEB 1080p AVC EAC3] [Eng-Sub] | Ikoku Nikki")
	assertEqual(t, engHyphenSub.ReleaseGroup, "KawaSubs")
	assertEqual(t, engHyphenSub.Title, "Journal with Witch")
	assertSliceContains(t, engHyphenSub.Subtitles, "Eng-Sub")
	assertSliceContains(t, engHyphenSub.Language, "Eng")
	assertSliceContains(t, engHyphenSub.SubtitleLanguage, "Eng")
}

func TestParseConsecutiveLeadingReleaseGroups(t *testing.T) {
	meta := Parse("[R-Archive] [AB-Corner] ASTRO BOY TETSUWAN ATOM Special The Secret of Atom's Birth 360p Web-DL Eng Sub")

	assertEqual(t, meta.ReleaseGroup, "R-Archive & AB-Corner")
	assertEqual(t, meta.Title, "ASTRO BOY TETSUWAN ATOM Special The Secret of Atom's Birth")
	assertEqual(t, meta.VideoResolution, "360p")
	assertSliceContains(t, meta.Source, "WEB-DL")
	assertSliceContains(t, meta.Subtitles, "Eng-Sub")
	assertSliceContains(t, meta.SubtitleLanguage, "Eng")
}

func TestParseLooseMovieYearBeforeTechnicalStack(t *testing.T) {
	meta := Parse("maboroshi 2023 MULTi AD 1080p NF WEB-DL DDP5.1 DV HDR10 H.265-Tsundere-Raws (VF, FRENCH, SUBFRENCH, VOSTFR, Alice and Therese's Illusion Factory)")

	assertEqual(t, meta.Title, "maboroshi")
	assertEqual(t, meta.Year, "2023")
	assertEqual(t, meta.FormattedTitle, "maboroshi (2023)")
	assertSliceContains(t, meta.Source, "NF")
	assertSliceContains(t, meta.Source, "WEB-DL")
	assertSliceContains(t, meta.AudioTerm, "DDP")
	assertSliceContains(t, meta.VideoTerm, "HDR10")
	assertSliceContains(t, meta.VideoTerm, "H.265")
	assertEqual(t, meta.VideoResolution, "1080p")
	assertEqual(t, meta.ReleaseGroup, "Tsundere-Raws")
}

func TestParseTitleYearBeforeSeasonEpisodeStaysTitle(t *testing.T) {
	meta := Parse("[ToonsHub] Way of Choices 2026 S01E16 1080p iQ WEB-DL AAC2.0 H.264 (Multi-Subs)")

	assertEqual(t, meta.ReleaseGroup, "ToonsHub")
	assertEqual(t, meta.Title, "Way of Choices 2026")
	assertEqual(t, meta.Year, "")
	assertSliceEqual(t, meta.SeasonNumber, []string{"1"})
	assertSliceEqual(t, meta.EpisodeNumber, []string{"16"})
}

func TestParseRepackSuffixAfterEpisodeAsReleaseMetadata(t *testing.T) {
	meta := Parse("[Freehold] Chained Soldier S02E08 REPACK2 [ADN WEB-DL 1080p x264 AAC EAC3 Dual-Audio Uncensored] | Mato Seihei no Slave Season 2")

	assertEqual(t, meta.ReleaseGroup, "Freehold")
	assertEqual(t, meta.Title, "Chained Soldier")
	assertSliceEqual(t, meta.SeasonNumber, []string{"2"})
	assertSliceEqual(t, meta.EpisodeNumber, []string{"8"})
	assertEqual(t, meta.EpisodeTitle, "")
	assertSliceContains(t, meta.ReleaseInformation, "REPACK")
	assertSliceContains(t, meta.ReleaseVersion, "2")
	assertSliceContains(t, meta.ReleaseInformation, "Uncensored")
	assertSliceContains(t, meta.Source, "ADN")
	assertSliceContains(t, meta.Source, "WEB-DL")
}

func TestParseDashSuffixReleaseGroupAfterStandaloneModeMarker(t *testing.T) {
	meta := Parse("One.Piece.Vol002.DVDRip.480p.x265.10bit.AAC.Multi-uP")

	assertEqual(t, meta.ReleaseGroup, "uP")
	assertEqual(t, meta.Title, "One Piece")
	assertSliceEqual(t, meta.VolumeNumber, []string{"2"})
	assertSliceContains(t, meta.Source, "DVD-Rip")
	assertEqual(t, meta.VideoResolution, "480p")
	assertSliceContains(t, meta.VideoTerm, "x265")
	assertSliceContains(t, meta.VideoTerm, "10bit")
	assertSliceContains(t, meta.AudioTerm, "AAC")
}

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func assertSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %#v, want %#v", got, want)
		}
	}
}

func assertSliceContains(t *testing.T, got []string, want string) {
	t.Helper()
	for _, value := range got {
		if value == want {
			return
		}
	}
	t.Fatalf("got %#v, want it to contain %q", got, want)
}

func assertSliceNotContains(t *testing.T, got []string, unwanted string) {
	t.Helper()
	for _, value := range got {
		if value == unwanted {
			t.Fatalf("got %#v, want it to omit %q", got, unwanted)
		}
	}
}

func parserForDateTest(filename string) *parser {
	p := newParser(filename, Options{})
	p.tokens = tokenize(p.baseName)
	p.groups = analyzeBrackets(p.tokens)
	return p
}

func tokenIndexForDateTest(t *testing.T, tokens []*Token, value string) int {
	t.Helper()
	for i, token := range tokens {
		if token.Value == value {
			return i
		}
	}
	t.Fatalf("token %q not found", value)
	return -1
}

type strictTitleGroundTruth struct {
	Name     string
	FileName string
	Expected map[string]any
}

func TestTitlesCorpusHumanGroundTruthStrict(t *testing.T) {
	cases := []strictTitleGroundTruth{
		{
			Name:     "modern toons hub sxe with alternate title",
			FileName: "[ToonsHub] Go For It Nakamura-kun S01E07 1080p CR WEB-DL MULTi AAC2.0 H.264 (Ganbare! Nakamura-kun!!, Multi-Audio, Multi-Subs)",
			Expected: strictExpectedFields(
				"release_group", "ToonsHub",
				"title", "Go For It Nakamura-kun",
				"season_number", []string{"1"},
				"episode_number", []string{"7"},
				"video_resolution", "1080p",
				"source", []string{"CR", "WEB-DL"},
				"audio_term", []string{"AAC", "Multi Audio"},
				"video_term", []string{"H.264"},
				"subtitles", []string{"Multi-Subs"},
			),
		},
		{
			Name:     "compact bracket technical stack",
			FileName: "[ASW] Nigetsuri - 06 [1080p HEVC x265 10Bit][AAC]",
			Expected: strictExpectedFields(
				"release_group", "ASW",
				"title", "Nigetsuri",
				"episode_number", []string{"6"},
				"video_resolution", "1080p",
				"audio_term", []string{"AAC"},
				"video_term", []string{"HEVC", "x265", "10bit"},
			),
		},
		{
			Name:     "episode title before technical stack and suffix group",
			FileName: "Chained Soldier S02E09 A Commanders Resolve 1080p HIDI WEB-DL DUAL AAC2.0 H 264-VARYG (Mato Seihei no Slave 2, Dual-Audio)",
			Expected: strictExpectedFields(
				"release_group", "VARYG",
				"title", "Chained Soldier",
				"season_number", []string{"2"},
				"episode_number", []string{"9"},
				"episode_title", "A Commanders Resolve",
				"video_resolution", "1080p",
				"source", []string{"HIDI", "WEB-DL"},
				"audio_term", []string{"AAC", "Dual Audio"},
				"video_term", []string{"H.264"},
			),
		},
		{
			Name:     "dub language checksum release",
			FileName: "[Yameii] Go For It, Nakamura-kun!! - S01E07 [English Dub] [CR WEB-DL 1080p H264 AAC] [D4A7DF3A] (Ganbare! Nakamura-kun!!)",
			Expected: strictExpectedFields(
				"release_group", "Yameii",
				"title", "Go For It, Nakamura-kun!!",
				"season_number", []string{"1"},
				"episode_number", []string{"7"},
				"language", []string{"English"},
				"subtitles", []string{},
				"audio_term", []string{"Dub", "AAC"},
				"video_resolution", "1080p",
				"source", []string{"CR", "WEB-DL"},
				"video_term", []string{"H.264"},
				"file_checksum", "D4A7DF3A",
			),
		},
		{
			Name:     "ani slash multilingual title",
			FileName: "[ANi] Candy Caries /  CANDY CARIES 蛀在糖糖裡 - 04 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Expected: strictExpectedFields(
				"release_group", "ANi",
				"title", "Candy Caries / CANDY CARIES 蛀在糖糖裡",
				"episode_number", []string{"4"},
				"video_resolution", "1080P",
				"source", []string{"Baha", "WEB-DL"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"AVC"},
				"language", []string{"CHT"},
			),
		},
		{
			Name:     "pipe separated alternate title after metadata",
			FileName: "[Unfucked] Gals Can't Be Kind to Otaku!? - S01E05 (1080p CR WEB-DL AVC AAC 2.0) | Otaku ni Yasashii Gal wa Inai!?",
			Expected: strictExpectedFields(
				"release_group", "Unfucked",
				"title", "Gals Can't Be Kind to Otaku!?",
				"season_number", []string{"1"},
				"episode_number", []string{"5"},
				"video_resolution", "1080p",
				"source", []string{"CR", "WEB-DL"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"AVC"},
			),
		},
		{
			Name:     "pipe separated multilingual title with total episode",
			FileName: "Needy Girl Overdose | Зависимая девушка: Передозировка [TV-1] [2026] [05 of 12] [WEBRip] [1080p] [RUS + JAP]",
			Expected: strictExpectedFields(
				"title", "Needy Girl Overdose",
				"year", "2026",
				"episode_number", []string{"5"},
				"other_episode_number", []string{"12"},
				"episode_title", "",
				"source", []string{"WEBRip"},
				"video_resolution", "1080p",
				"language", []string{"RUS", "Jap"},
			),
		},
		{
			Name:     "repack episode title and suffix group",
			FileName: "One Piece S01E1160 An Encounter on a Snowfield-Loki the Accursed Prince REPACK 1080p CR WEB-DL AAC2.0 H 264-VARYG (Multi-Subs)",
			Expected: strictExpectedFields(
				"release_group", "VARYG",
				"title", "One Piece",
				"season_number", []string{"1"},
				"episode_number", []string{"1160"},
				"episode_title", "An Encounter on a Snowfield-Loki the Accursed Prince",
				"release_information", []string{"REPACK"},
				"video_resolution", "1080p",
				"source", []string{"CR", "WEB-DL"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"H.264"},
				"subtitles", []string{"Multi-Subs"},
			),
		},
		{
			Name:     "movie title number is not episode",
			FileName: "Patlabor 2 - Le Film (1993) CUSTOM REMASTER MULTi-FRENCH/VOSTFR 2160p (4K) HDR 10bits BluRay x265-GundamGuy (VF) (Mobile Police Patlabor 2 The Movie)",
			Expected: strictExpectedFields(
				"release_group", "GundamGuy",
				"title", "Patlabor 2 - Le Film",
				"year", "1993",
				"episode_number", []string{},
				"release_information", []string{"CUSTOM", "Remaster"},
				"video_resolution", "2160p",
				"source", []string{"Blu-Ray"},
				"video_term", []string{"HDR", "10bits", "x265"},
				"language", []string{"VOSTFR", "VF"},
			),
		},
		{
			Name:     "rus sub bracket language",
			FileName: "[Pokefans] Pocket Monsters (2023) 136 [RUS SUB, 1080p].mkv",
			Expected: strictExpectedFields(
				"release_group", "Pokefans",
				"title", "Pocket Monsters",
				"year", "2023",
				"episode_number", []string{"136"},
				"language", []string{"RUS"},
				"subtitles", []string{"Sub"},
				"video_resolution", "1080p",
				"file_extension", "mkv",
			),
		},
		{
			Name:     "dual raw title with suffix release group",
			FileName: "機動戦士ガンダム.Mobile.Suit.Gundam.SEED.Freedom.2024.Movie.BDRip.1080p+4K.2160p.HDR10.x265.FLAC/Atmos-CoolFansSub&Sakurato&Comicat",
			Expected: strictExpectedFields(
				"release_group", "CoolFansSub&Sakurato&Comicat",
				"title", "機動戦士ガンダム Mobile Suit Gundam SEED Freedom",
				"year", "2024",
				"anime_type", []string{"Movie"},
				"video_resolution", "2160p",
				"source", []string{"BDRip"},
				"audio_term", []string{"FLAC", "Atmos"},
				"video_term", []string{"HDR10", "x265"},
			),
		},
		{
			Name:     "lazyleido episode plus parenthesized sxe",
			FileName: "[Lazyleido] DIGIMON BEATBREAK - 29 (S01E29) - (WEB 1080p HEVC x265 10-bit AAC 2.0) [8FD512A4]",
			Expected: strictExpectedFields(
				"release_group", "Lazyleido",
				"title", "DIGIMON BEATBREAK",
				"season_number", []string{"1"},
				"episode_number", []string{"29"},
				"video_resolution", "1080p",
				"source", []string{"WEB"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"HEVC", "x265", "10bit"},
				"file_checksum", "8FD512A4",
			),
		},
		{
			Name:     "digits inside leading release group are not episodes",
			FileName: "[7³ACG & 343-Labs] Yofukashi no Uta よふかしのうた Season 2 [BDRip 1080p AV1 OPUS] (Call of the Night 2)",
			Expected: strictExpectedFields(
				"release_group", "7³ACG & 343-Labs",
				"title", "Yofukashi no Uta よふかしのうた",
				"season_number", []string{"2"},
				"episode_number", []string{},
				"episode_title", "",
				"source", []string{"BDRip"},
				"audio_term", []string{"Opus"},
				"video_term", []string{"AV1"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "lowercase as remains title word not source",
			FileName: "[ToonsHub] That Time I Got Reincarnated as a Slime S04E03 1080p CR WEB-DL DUAL AAC2.0 H.264 (Tensei Shitara Slime Datta Ken, Dual-Audio, Multi-Subs)",
			Expected: strictExpectedFields(
				"release_group", "ToonsHub",
				"title", "That Time I Got Reincarnated as a Slime",
				"season_number", []string{"4"},
				"episode_number", []string{"3"},
				"source", []string{"CR", "WEB-DL"},
				"audio_term", []string{"AAC", "Dual Audio"},
				"subtitles", []string{"Multi-Subs"},
				"video_term", []string{"H.264"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "dot delimited multi marker is not episode title",
			FileName: "Rooster.Fighter.S01E08.MULTi.1080p.WEBRiP.x265-KAF",
			Expected: strictExpectedFields(
				"release_group", "KAF",
				"title", "Rooster Fighter",
				"season_number", []string{"1"},
				"episode_number", []string{"8"},
				"episode_title", "",
				"source", []string{"WEBRip"},
				"video_term", []string{"x265"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "compact ep prefix parses episode",
			FileName: "[ToonsHub] One Piece EP1160 1080p CR WEB-DL AAC2.0 H.264 (English-Sub)",
			Expected: strictExpectedFields(
				"release_group", "ToonsHub",
				"title", "One Piece",
				"episode_number", []string{"1160"},
				"source", []string{"CR", "WEB-DL"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"H.264"},
				"subtitles", []string{"English-Sub"},
				"language", []string{"English"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "dot delimited movie year before multi marker",
			FileName: "Boku.No.Hero.Academia.You.re.Next.2024.MULTi.1080p.WEBRiP.x265-KAF",
			Expected: strictExpectedFields(
				"release_group", "KAF",
				"title", "Boku No Hero Academia You re Next",
				"year", "2024",
				"episode_number", []string{},
				"source", []string{"WEBRip"},
				"video_term", []string{"x265"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "numeric sequel before dash episode is not episode range",
			FileName: "[Erai-raws] Otonari no Tenshi-sama ni Itsunomanika Dame Ningen ni Sareteita Ken 2 - 05 [1080p CR WEBRip HEVC AAC][MultiSub][B5EC786B]",
			Expected: strictExpectedFields(
				"release_group", "Erai-raws",
				"title", "Otonari no Tenshi-sama ni Itsunomanika Dame Ningen ni Sareteita Ken 2",
				"episode_number", []string{"5"},
				"source", []string{"CR", "WEBRip"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"HEVC"},
				"subtitles", []string{"MultiSub"},
				"video_resolution", "1080p",
				"file_checksum", "B5EC786B",
			),
		},
		{
			Name:     "special in title phrase is not anime type",
			FileName: "Special Value-2 Full Episodes Of Fruits Basket (Promo DVD) DVDRemux",
			Expected: strictExpectedFields(
				"title", "Special Value",
				"episode_number", []string{"2"},
				"episode_title", "Full Episodes Of Fruits Basket",
				"anime_type", []string{},
				"source", []string{"DVD"},
			),
		},
		{
			Name:     "movie inside title phrase is not anime type",
			FileName: "CHAINSAW MAN THE MOVIE REZE ARC 2025 1080p CR WEB-DL MULTi AAC2.0 H 264-VARYG (Chainsaw Man: Reze-hen, Multi-Audio, Multi-Subs)",
			Expected: strictExpectedFields(
				"release_group", "VARYG",
				"title", "CHAINSAW MAN THE MOVIE REZE ARC",
				"year", "2025",
				"anime_type", []string{},
				"source", []string{"CR", "WEB-DL"},
				"audio_term", []string{"AAC", "Multi Audio"},
				"subtitles", []string{"Multi-Subs"},
				"video_term", []string{"H.264"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "adjacent bracket episode has no phantom closing bracket title",
			FileName: "[桜都字幕组] 冰之城墙 / Koori no Jouheki [05][1080P][简繁内封]",
			Expected: strictExpectedFields(
				"release_group", "桜都字幕组",
				"title", "冰之城墙 / Koori no Jouheki",
				"episode_number", []string{"5"},
				"episode_title", "",
				"video_resolution", "1080P",
			),
		},
		{
			Name:     "multi bracket title after release group",
			FileName: "[SweetSub][小書痴的下剋上 領主的養女][Honzuki no Gekokujou S04][02][WebRip][1080P][AVC 8bit][繁日雙語]（第四季）",
			Expected: strictExpectedFields(
				"release_group", "SweetSub",
				"title", "小書痴的下剋上 領主的養女",
				"season_number", []string{"4"},
				"episode_number", []string{"2"},
				"episode_title", "",
				"source", []string{"WEBRip"},
				"video_term", []string{"AVC", "8bit"},
				"video_resolution", "1080P",
			),
		},
		{
			Name:     "lowercase as in alternate title group is not source",
			FileName: "That Time I Got Reincarnated as a Slime S04E03 VOSTFR 1080p WEB x264 AAC -Tsundere-Raws (CR) (Tensei shitara Slime Datta Ken 4th Season,That Time I Got Reincarnated as a Slime Season 4)",
			Expected: strictExpectedFields(
				"release_group", "Tsundere-Raws",
				"title", "That Time I Got Reincarnated as a Slime",
				"season_number", []string{"4"},
				"episode_number", []string{"3"},
				"source", []string{"WEB", "CR"},
				"language", []string{"VOSTFR"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"x264"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "plain year before bracket technical stack is not episode",
			FileName: "[7³ACG & 343-Labs] Fushigi no Kuni de Alice to 不思議の国でアリスと -Dive in Wonderland- 2025 [BDRip 1080p AV1 OPUS]",
			Expected: strictExpectedFields(
				"release_group", "7³ACG & 343-Labs",
				"title", "Fushigi no Kuni de Alice to 不思議の国でアリスと -Dive in Wonderland",
				"year", "2025",
				"episode_number", []string{},
				"source", []string{"BDRip"},
				"audio_term", []string{"Opus"},
				"video_term", []string{"AV1"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "bd catalog code is not episode",
			FileName: "The Red Turtle (2016) レッドタートル ある島の物語 [12-bit 4:2:0 Decrypted MGVC BD ISO VWBS-8782]",
			Expected: strictExpectedFields(
				"title", "The Red Turtle",
				"year", "2016",
				"release_group", "",
				"episode_number", []string{},
				"episode_title", "",
				"source", []string{"BDISO"},
			),
		},
		{
			Name:     "bracket title before anime type metadata",
			FileName: "[9901RAW][Weiss_Kreuz_OVA][DVDRIP]",
			Expected: strictExpectedFields(
				"release_group", "9901RAW",
				"title", "Weiss Kreuz",
				"anime_type", []string{"OVA"},
				"source", []string{"DVD-Rip"},
			),
		},
		{
			Name:     "ordinal season before dash episode is structured metadata",
			FileName: "[Erai-raws] Re:Zero kara Hajimeru Isekai Seikatsu 4th Season - 05 [1080p CR WEBRip HEVC AAC][MultiSub][E023F814]",
			Expected: strictExpectedFields(
				"release_group", "Erai-raws",
				"title", "Re:Zero kara Hajimeru Isekai Seikatsu",
				"season_number", []string{"4"},
				"episode_number", []string{"5"},
				"source", []string{"CR", "WEBRip"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"HEVC"},
				"subtitles", []string{"MultiSub"},
				"video_resolution", "1080p",
				"file_checksum", "E023F814",
			),
		},
		{
			Name:     "ordinal season with colon subtitle keeps subtitle in title",
			FileName: "[Erai-raws] Youkoso Jitsuryoku Shijou Shugi no Kyoushitsu e 4th Season: 2-nensei-hen 1 Gakki - 09 [1080p CR WEBRip HEVC AAC][MultiSub][2E96AA1F]",
			Expected: strictExpectedFields(
				"release_group", "Erai-raws",
				"title", "Youkoso Jitsuryoku Shijou Shugi no Kyoushitsu e: 2-nensei-hen 1 Gakki",
				"season_number", []string{"4"},
				"episode_number", []string{"9"},
				"source", []string{"CR", "WEBRip"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"HEVC"},
				"subtitles", []string{"MultiSub"},
				"video_resolution", "1080p",
				"file_checksum", "2E96AA1F",
			),
		},
		{
			Name:     "short streaming service before webdl is source",
			FileName: "[ToonsHub] One Piece EP1160 1080p iQ WEB-DL AAC2.0 H.264 (Multi-Subs)",
			Expected: strictExpectedFields(
				"release_group", "ToonsHub",
				"title", "One Piece",
				"episode_number", []string{"1160"},
				"source", []string{"iQ", "WEB-DL"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"H.264"},
				"subtitles", []string{"Multi-Subs"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "broadcast channel in parenthesized stack is source",
			FileName: "[shincaps] Marika-chan no Koukando wa Bukkowareteiru - 04 (ANIPLUS 1920x1080 H264 AAC).ts",
			Expected: strictExpectedFields(
				"release_group", "shincaps",
				"title", "Marika-chan no Koukando wa Bukkowareteiru",
				"episode_number", []string{"4"},
				"source", []string{"ANIPLUS"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"H.264"},
				"video_resolution", "1920x1080",
				"file_extension", "ts",
			),
		},
		{
			Name:     "hyphenated broadcast channel before resolution is source",
			FileName: "[shincaps] Kami no Niwatsuki Kusunoki-tei - 05 (BS-EX 1440x1080 MPEG2 AAC).ts",
			Expected: strictExpectedFields(
				"release_group", "shincaps",
				"title", "Kami no Niwatsuki Kusunoki-tei",
				"episode_number", []string{"5"},
				"source", []string{"BS-EX"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"MPEG2"},
				"video_resolution", "1440x1080",
				"file_extension", "ts",
			),
		},
		{
			Name:     "alternate title season does not become second main season",
			FileName: "Re ZERO Starting Life in Another World S04E05 Stick Swinger 1080p CR WEB-DL MULTi AAC2.0 H 264-VARYG (Re:Zero kara Hajimeru Isekai Seikatsu 2nd Season Part 2, Multi-Audio, Multi-Subs)",
			Expected: strictExpectedFields(
				"release_group", "VARYG",
				"title", "Re ZERO Starting Life in Another World",
				"season_number", []string{"4"},
				"part_number", []string{},
				"episode_number", []string{"5"},
				"episode_title", "Stick Swinger",
				"source", []string{"CR", "WEB-DL"},
				"audio_term", []string{"AAC", "Multi Audio"},
				"video_term", []string{"H.264"},
				"subtitles", []string{"Multi-Subs"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "bracketed alternate title season does not override main sxe",
			FileName: "OSHI NO KO S03E06 Idols and Relationships 1080p CR WEB-DL MULTi AAC2.0 H 264-VARYG ([Oshi no Ko] 2nd Season, Multi-Audio, Multi-Subs)",
			Expected: strictExpectedFields(
				"release_group", "VARYG",
				"title", "OSHI NO KO",
				"season_number", []string{"3"},
				"episode_number", []string{"6"},
				"episode_title", "Idols and Relationships",
				"source", []string{"CR", "WEB-DL"},
				"audio_term", []string{"AAC", "Multi Audio"},
				"video_term", []string{"H.264"},
				"subtitles", []string{"Multi-Subs"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "bare ordinal before dash episode is season metadata",
			FileName: "[AsukaRaws] Re Zero kara Hajimeru Isekai Seikatsu 4th - 05 (71) (WEB-DL 1280x720 x264 AAC)",
			Expected: strictExpectedFields(
				"release_group", "AsukaRaws",
				"title", "Re Zero kara Hajimeru Isekai Seikatsu",
				"season_number", []string{"4"},
				"episode_number", []string{"5"},
				"episode_number_alt", []string{"71"},
				"source", []string{"WEB-DL"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"x264"},
				"video_resolution", "1280x720",
			),
		},
		{
			Name:     "episode word inside episode title is not second episode number",
			FileName: "[Gecko] LINK CLICK - S00E25 - Mini Link Click Episode 5 (时光小剧场; Shiguang Xiao Juchang; Shiguang Dailiren Mini Theater 2) [YTB.WEB-DL 1080P AVC, Opus][81B530F3]",
			Expected: strictExpectedFields(
				"release_group", "Gecko",
				"title", "LINK CLICK",
				"season_number", []string{"0"},
				"episode_number", []string{"25"},
				"episode_title", "Mini Link Click Episode 5",
				"source", []string{"YTB", "WEB-DL"},
				"audio_term", []string{"Opus"},
				"video_term", []string{"AVC"},
				"video_resolution", "1080P",
				"file_checksum", "81B530F3",
			),
		},
		{
			Name:     "bracketed broadcast year range is not episode or release group",
			FileName: "Dragon Ball Daima (ドラゴンボールＤＡＩＭＡ) Fuji TV Broadcast [2024-2025]",
			Expected: strictExpectedFields(
				"title", "Dragon Ball Daima (ドラゴンボールDAIMA) Fuji TV Broadcast",
				"year", "2024",
				"episode_number", []string{},
				"release_group", "",
			),
		},
		{
			Name:     "paired tilde title keeps closing tilde",
			FileName: "[shincaps] Shunkashuutou Daikousha ~Haru no Mai~ - 06 (BS11 1920x1080 MPEG2 AAC).ts",
			Expected: strictExpectedFields(
				"release_group", "shincaps",
				"title", "Shunkashuutou Daikousha ~Haru no Mai~",
				"episode_number", []string{"6"},
				"source", []string{"BS11"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"MPEG2"},
				"video_resolution", "1920x1080",
				"file_extension", "ts",
			),
		},
		{
			Name:     "bs tx broadcast channel before resolution is source",
			FileName: "[shincaps] LIAR GAME - 04 (BS-TX 1440x1080 MPEG2 AAC).ts",
			Expected: strictExpectedFields(
				"release_group", "shincaps",
				"title", "LIAR GAME",
				"episode_number", []string{"4"},
				"source", []string{"BS-TX"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"MPEG2"},
				"video_resolution", "1440x1080",
				"file_extension", "ts",
			),
		},
		{
			Name:     "cs ex2 broadcast channel before resolution is source",
			FileName: "[shincaps] Crayon Shin-chan - 340-349 (CS-EX2 1440x1080 MPEG2 AAC).ts",
			Expected: strictExpectedFields(
				"release_group", "shincaps",
				"title", "Crayon Shin-chan",
				"episode_number", []string{"340", "349"},
				"source", []string{"CS-EX2"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"MPEG2"},
				"video_resolution", "1440x1080",
				"file_extension", "ts",
			),
		},
		{
			Name:     "bs fuji broadcast channel before resolution is source",
			FileName: "[shincaps] Awajima Hyakkei - 01 (BS-FUJI 1440x1080 MPEG2 AAC).ts",
			Expected: strictExpectedFields(
				"release_group", "shincaps",
				"title", "Awajima Hyakkei",
				"episode_number", []string{"1"},
				"source", []string{"BS-FUJI"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"MPEG2"},
				"video_resolution", "1440x1080",
				"file_extension", "ts",
			),
		},
		{
			Name:     "pipe alternate sxe becomes alternate episode not second main episode",
			FileName: "[yaneura] Re:Zero Break Time S4 - 03 (HDTV 1080p) | ReZero S00E73",
			Expected: strictExpectedFields(
				"release_group", "yaneura",
				"title", "Re:Zero Break Time",
				"season_number", []string{"4"},
				"episode_number", []string{"3"},
				"episode_number_alt", []string{"73"},
				"source", []string{"HDTV"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "pipe episode prefix becomes alternate episode",
			FileName: "[Asakura] Tensei Shitara Slime Datta Ken 4th Season - 03 [1080p WEB AAC x264] [0796CBAD] | That Time I Got Reincarnated as a Slime Season 4 | Episode 75",
			Expected: strictExpectedFields(
				"release_group", "Asakura",
				"title", "Tensei Shitara Slime Datta Ken",
				"season_number", []string{"4"},
				"episode_number", []string{"3"},
				"episode_number_alt", []string{"75"},
				"source", []string{"WEB"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"x264"},
				"video_resolution", "1080p",
				"file_checksum", "0796CBAD",
			),
		},
		{
			Name:     "pipe alternate season descriptor does not add conflicting main season",
			FileName: "[GetItTwisted] Samurai Girls S00E14-E19 [BD 1080p AVC Opus Dual-Audio] | Hyakka Ryouran: Samurai Bride | Season 2 BD Specials",
			Expected: strictExpectedFields(
				"release_group", "GetItTwisted",
				"title", "Samurai Girls",
				"season_number", []string{"0"},
				"episode_number", []string{"14", "19"},
				"anime_type", []string{"Specials"},
				"source", []string{"BD"},
				"audio_term", []string{"Opus", "Dual Audio"},
				"video_term", []string{"AVC"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "broadcast date is not episode and hash number is episode",
			FileName: "涼宮ハルヒの憂鬱 #4「涼宮ハルヒの退屈」[再] (TOKYO MX1 2026-04-20).ts | Suzumiya Haruhi no Yuuutsu - 04 without tsunami map",
			Expected: strictExpectedFields(
				"title", "涼宮ハルヒの憂鬱",
				"year", "2026",
				"episode_number", []string{"4"},
				"episode_title", "涼宮ハルヒの退屈",
			),
		},
		{
			Name:     "season plus list does not become episode number",
			FileName: "Tadaima!/I'm Home! Chibi Godzilla Season 1+2",
			Expected: strictExpectedFields(
				"title", "Tadaima!/I'm Home! Chibi Godzilla",
				"season_number", []string{"1", "2"},
				"episode_number", []string{},
			),
		},
		{
			Name:     "tv hyphen season label is not title or episode",
			FileName: "Tsue to Tsurugi no Wistoria TV-2 | Меч и жезл Вистории [ТВ-2] | Wistoria: Wand and Sword 2 [2026] [04 of 12] [WEBRip] [1080p] [RUS + JAP]",
			Expected: strictExpectedFields(
				"title", "Tsue to Tsurugi no Wistoria",
				"year", "2026",
				"season_number", []string{"2"},
				"episode_number", []string{"4"},
				"other_episode_number", []string{"12"},
				"episode_title", "",
				"source", []string{"WEBRip"},
				"video_resolution", "1080p",
				"language", []string{"RUS", "Jap"},
			),
		},
		{
			Name:     "compact bd resolution bracket keeps release group suffix",
			FileName: "Eureka.Seven.Hi-Evolution.1.[BD1080p.x264-SHiNiGAMi].VOSTFR",
			Expected: strictExpectedFields(
				"release_group", "SHiNiGAMi",
				"title", "Eureka Seven Hi-Evolution",
				"episode_number", []string{"1"},
				"source", []string{"BD"},
				"video_resolution", "1080p",
				"video_term", []string{"x264"},
				"language", []string{"VOSTFR"},
			),
		},
		{
			Name:     "season range plus ova has no episode",
			FileName: "[DB] Yuukoku no Moriarty | Moriarty the Patriot (Season 1-2+OVA) [Dual Audio 10bit BD1080p][HEVC-x265]",
			Expected: strictExpectedFields(
				"release_group", "DB",
				"title", "Yuukoku no Moriarty",
				"season_number", []string{"1", "2"},
				"episode_number", []string{},
				"anime_type", []string{"OVA"},
				"audio_term", []string{"Dual Audio"},
				"source", []string{"BD"},
				"video_resolution", "1080p",
				"video_term", []string{"10bit", "HEVC", "x265"},
			),
		},
		{
			Name:     "season range plus movie collection keeps media metadata",
			FileName: "[anime4life.] Konosuba! - God Blessing on this Wonderful World! Season 1-3 + Movie LPCM_Dolby TrueHD BD_1080p Dual Audio",
			Expected: strictExpectedFields(
				"release_group", "anime4life.",
				"title", "Konosuba! - God Blessing on this Wonderful World!",
				"season_number", []string{"1", "3"},
				"episode_number", []string{},
				"anime_type", []string{"Movie"},
				"audio_term", []string{"LPCM", "Dolby", "TrueHD", "Dual Audio"},
				"source", []string{"BD"},
				"video_resolution", "1080p",
			),
		},
		{
			Name:     "audio channel layout is not episode",
			FileName: "攻殻機動隊.Ghost.in.the.Shell.1995.Movie.BDRip.4K.UHD.2160p.x265.DV.TrueHD.7.1.Atmos-CoolFansSub   (Koukaku Kidoutai)",
			Expected: strictExpectedFields(
				"release_group", "CoolFansSub",
				"title", "攻殻機動隊 Ghost in the Shell",
				"year", "1995",
				"episode_number", []string{},
				"anime_type", []string{"Movie"},
				"audio_term", []string{"TrueHD", "Atmos"},
				"source", []string{"BDRip"},
				"video_resolution", "2160p",
				"video_term", []string{"x265", "DV"},
			),
		},
		{
			Name:     "u17 title segment is not episode",
			FileName: "The.Prince.of.Tennis.II.U-17.World.Cup.Semifinal.E13.Final.WEBRip.x264-H3AsO3",
			Expected: strictExpectedFields(
				"release_group", "H3AsO3",
				"title", "The Prince of Tennis II.U-17 World Cup Semifinal",
				"episode_number", []string{"13"},
				"source", []string{"WEBRip"},
				"video_term", []string{"x264"},
			),
		},
		{
			Name:     "standalone movie year before language is year",
			FileName: "For Whom The Alchemist Exists 2019 SUBFRENCH 1080p CR WEB-DL.AAC2.0.x264-Tsundere-Raws (VOSTFR, Multi Subs, Ta ga Tame no Alchemist, Dare ga Tame no Alchemist, For Whom The Alchemist Exists The Movie)",
			Expected: strictExpectedFields(
				"release_group", "Tsundere-Raws",
				"title", "For Whom The Alchemist Exists",
				"year", "2019",
				"episode_number", []string{},
				"source", []string{"CR", "WEB-DL"},
				"audio_term", []string{"AAC"},
				"video_term", []string{"x264"},
				"video_resolution", "1080p",
				"language", []string{"SUBFRENCH", "VOSTFR"},
				"subtitles", []string{"Multi-Subs"},
			),
		},
		{
			Name:     "dual subs after episode range is not episode title",
			FileName: "[Saiko] Kinnikuman Nisei BATCH 01 (Episodes 1-4)  DUAL SUBS: ENGLISH + ESPAÑOL",
			Expected: strictExpectedFields(
				"release_group", "Saiko",
				"title", "Kinnikuman Nisei",
				"episode_number", []string{"1", "4"},
				"episode_title", "",
				"release_information", []string{"Batch"},
				"subtitles", []string{"Subs"},
				"language", []string{"English", "Spanish"},
			),
		},
		{
			Name:     "dot delimited title words and title number become spaces",
			FileName: "Chillin'.in.Another.World.With.Level.2.Super.Cheat.Powers.S01.1080p.Blu-ray.Remux.Dual-Audio.FLAC.2.0.H.264-LaCroiX | Lv2 Kara Cheat datta Moto Yuusha Kouho no Mattari Isekai Life",
			Expected: strictExpectedFields(
				"release_group", "LaCroiX",
				"title", "Chillin' in Another World With Level 2 Super Cheat Powers",
				"season_number", []string{"1"},
				"episode_number", []string{},
				"audio_term", []string{"Dual Audio", "FLAC"},
				"release_information", []string{"Remux"},
				"source", []string{"Blu-Ray"},
				"video_resolution", "1080p",
				"video_term", []string{"H.264"},
			),
		},
		{
			Name:     "double dash ova episode range keeps both endpoints",
			FileName: "Yankee Repputai ova 1--6",
			Expected: strictExpectedFields(
				"title", "Yankee Repputai",
				"anime_type", []string{"OVA"},
				"episode_number", []string{"1", "6"},
				"episode_title", "",
			),
		},
		{
			Name:     "bdmv marker is source not release group and quoted suffix remains title",
			FileName: "[BDMV][VWBS-7147] New Kabuki Production Nausicaä of the Valley of the Wind - 新作歌舞伎『風の谷のナウシカ』",
			Expected: strictExpectedFields(
				"release_group", "",
				"title", "New Kabuki Production Nausicaä of the Valley of the Wind - 新作歌舞伎『風の谷のナウシカ』",
				"source", []string{"BDMV"},
				"episode_number", []string{},
			),
		},
		{
			Name:     "hd quality marker after episode is resolution not episode title",
			FileName: "[APTX-Fansub] Detective Conan - 1197 FHD [8D939E16].mp4",
			Expected: strictExpectedFields(
				"release_group", "APTX-Fansub",
				"title", "Detective Conan",
				"episode_number", []string{"1197"},
				"episode_title", "",
				"video_resolution", "FHD",
				"file_checksum", "8D939E16",
				"file_extension", "mp4",
			),
		},
		{
			Name:     "opus channel suffix is audio not episode zero",
			FileName: "KissXSis.S00.1080p.BluRay.Opus2.0.x265-Headpatter",
			Expected: strictExpectedFields(
				"release_group", "Headpatter",
				"title", "KissXSis",
				"season_number", []string{"0"},
				"episode_number", []string{},
				"audio_term", []string{"Opus"},
				"source", []string{"Blu-Ray"},
				"video_resolution", "1080p",
				"video_term", []string{"x265"},
			),
		},
		{
			Name:     "dot delimited lowercase article remains title word",
			FileName: "KamiKatsu.Working.for.God.in.a.Godless.World.S01.1080p.BluRay.Remux.Dual.Audio.FLAC.2.0.H.264-Humble",
			Expected: strictExpectedFields(
				"release_group", "Humble",
				"title", "KamiKatsu Working for God in a Godless World",
				"season_number", []string{"1"},
				"episode_number", []string{},
				"audio_term", []string{"Dual Audio", "FLAC"},
				"release_information", []string{"Remux"},
				"source", []string{"Blu-Ray"},
				"video_resolution", "1080p",
				"video_term", []string{"H.264"},
			),
		},
		{
			Name:     "batch ordinal after release information is not episode",
			FileName: "Yu-Gi-Oh! Duel Monsters S02 - Batch 3 [DVDrip JP MULTI - VF NC & VOSTFR ADN]",
			Expected: strictExpectedFields(
				"title", "Yu-Gi-Oh! Duel Monsters",
				"season_number", []string{"2"},
				"episode_number", []string{},
				"release_information", []string{"Batch"},
				"source", []string{"DVD-Rip", "ADN"},
				"language", []string{"jp", "VF", "VOSTFR"},
			),
		},
		{
			Name:     "plain absolute episode before dash has episode title",
			FileName: "[HnY] Beyblade X 124 - True X Tower War (1080p) v2.mkv",
			Expected: strictExpectedFields(
				"release_group", "HnY",
				"title", "Beyblade X",
				"episode_number", []string{"124"},
				"episode_title", "True X Tower War",
				"release_version", []string{"2"},
				"video_resolution", "1080p",
				"file_extension", "mkv",
			),
		},
		{
			Name:     "parenthetical season episode notation after absolute episode is alternate",
			FileName: "[Naruto-Kun.Hu] Vigilante - Boku no Hero Academia Illegals 20 (2x07) [1080p].mkv",
			Expected: strictExpectedFields(
				"release_group", "Naruto-Kun.Hu",
				"title", "Vigilante - Boku no Hero Academia Illegals",
				"season_number", []string{"2"},
				"episode_number", []string{"20"},
				"episode_number_alt", []string{"7"},
				"video_resolution", "1080p",
				"file_extension", "mkv",
			),
		},
		{
			Name:     "leading bracket episode range wins over arc number",
			FileName: "[One Pace][397-398] Enies Lobby 10 [1080p][074A4A31].mkv",
			Expected: strictExpectedFields(
				"release_group", "One Pace",
				"title", "Enies Lobby",
				"episode_number", []string{"397", "398"},
				"video_resolution", "1080p",
				"file_checksum", "074A4A31",
				"file_extension", "mkv",
			),
		},
		{
			Name:     "ordinal season shorthand before underscore episode",
			FileName: "[EA]Re_Zero_kara_Hajimeru_Isekai_Seikatsu_4th_02_[1920x1080][HEVC][C7F327FC].mkv",
			Expected: strictExpectedFields(
				"release_group", "EA",
				"title", "Re Zero kara Hajimeru Isekai Seikatsu",
				"season_number", []string{"4"},
				"episode_number", []string{"2"},
				"video_resolution", "1920x1080",
				"video_term", []string{"HEVC"},
				"file_checksum", "C7F327FC",
				"file_extension", "mkv",
			),
		},
		{
			Name:     "bare season shorthand before dash episode",
			FileName: "[Ñ] ReːZero kara Hajimeru Isekai Seikatsu 4 - 02 [1080p] [04A19242].mkv",
			Expected: strictExpectedFields(
				"release_group", "Ñ",
				"title", "ReːZero kara Hajimeru Isekai Seikatsu",
				"season_number", []string{"4"},
				"episode_number", []string{"2"},
				"video_resolution", "1080p",
				"file_checksum", "04A19242",
				"file_extension", "mkv",
			),
		},
		{
			Name:     "dot delimited episode title trims terminal dot before metadata",
			FileName: "SHY.E17.Assalt.WEBRip.x264-H3AsO3",
			Expected: strictExpectedFields(
				"release_group", "H3AsO3",
				"title", "SHY",
				"episode_number", []string{"17"},
				"episode_title", "Assalt",
				"source", []string{"WEBRip"},
				"video_term", []string{"x264"},
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			meta := Parse(tc.FileName)
			if mismatches := compareStrictGroundTruth(meta, tc.Expected); len(mismatches) > 0 {
				t.Fatalf("strict mismatches for %q:\n%s\nmetadata: %#v", tc.FileName, strings.Join(mismatches, "\n"), meta)
			}
		})
	}
}

func strictExpectedFields(values ...any) map[string]any {
	if len(values)%2 != 0 {
		panic("strictExpectedFields requires key/value pairs")
	}
	result := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic("strictExpectedFields key must be string")
		}
		result[key] = values[i+1]
	}
	return result
}

func compareStrictGroundTruth(meta *Metadata, expected map[string]any) []string {
	mismatches := make([]string, 0)
	keys := make([]string, 0, len(expected))
	for key := range expected {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		got, ok := metadataField(meta, key)
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("  %s: unknown field", key))
			continue
		}
		if !reflect.DeepEqual(strictComparableValue(got), strictComparableValue(expected[key])) {
			mismatches = append(mismatches, fmt.Sprintf("  %s: want %#v, got %#v", key, expected[key], got))
		}
	}
	return mismatches
}

func metadataField(meta *Metadata, key string) (any, bool) {
	switch key {
	case "title":
		return meta.Title, true
	case "year":
		return meta.Year, true
	case "season", "season_number":
		return meta.SeasonNumber, true
	case "part", "part_number":
		return meta.PartNumber, true
	case "volume_number":
		return meta.VolumeNumber, true
	case "episode", "episode_number":
		return meta.EpisodeNumber, true
	case "episode_number_alt":
		return meta.EpisodeNumberAlt, true
	case "other_episode_number":
		return meta.OtherEpisodeNumber, true
	case "episode_title":
		return meta.EpisodeTitle, true
	case "type", "anime_type":
		return meta.AnimeType, true
	case "audio_term":
		return meta.AudioTerm, true
	case "audio_language":
		return meta.AudioLanguage, true
	case "device_compatibility":
		return meta.DeviceCompatibility, true
	case "file_checksum":
		return meta.FileChecksum, true
	case "file_extension":
		return meta.FileExtension, true
	case "file_index":
		return meta.FileIndex, true
	case "language":
		return meta.Language, true
	case "release_group":
		return meta.ReleaseGroup, true
	case "release_information":
		return meta.ReleaseInformation, true
	case "release_version":
		return meta.ReleaseVersion, true
	case "source":
		return meta.Source, true
	case "subtitles", "subs_term":
		return meta.Subtitles, true
	case "subtitle_language":
		return meta.SubtitleLanguage, true
	case "video_resolution":
		return meta.VideoResolution, true
	case "video_term":
		return meta.VideoTerm, true
	default:
		return nil, false
	}
}

func strictComparableValue(value any) any {
	if values, ok := value.([]string); ok && values == nil {
		return []string{}
	}
	return value
}

func TestTokenizerSplitJoinedCJKAndMetadata(t *testing.T) {
	meta1 := Parse("[LoliHouse] 入间同学入魔了！S4 - 05.mkv")
	assertEqual(t, meta1.Title, "入间同学入魔了!") // Normalized ！ -> !
	assertSliceEqual(t, meta1.SeasonNumber, []string{"4"})
	assertSliceEqual(t, meta1.EpisodeNumber, []string{"5"})

	meta2 := Parse("[7³ACG] Egao no Taenai Shokuba Desu. 笑顔のたえない職場です。S1 [BD].mkv")
	assertEqual(t, meta2.Title, "Egao no Taenai Shokuba Desu. 笑顔のたえない職場です。")
	assertSliceEqual(t, meta2.SeasonNumber, []string{"1"})
}

func TestDashedEpisodeTitleSpaces(t *testing.T) {
	meta := Parse("[No0bSubs] Yuusha-kei ni Shosu - Choubatsu Yuusha 9004-tai Keimu Kiroku (1080p).mkv")
	assertEqual(t, meta.Title, "Yuusha-kei ni Shosu - Choubatsu Yuusha 9004-tai Keimu Kiroku")
	assertSliceEqual(t, meta.EpisodeNumber, nil)
}

func TestParseAndroidTitleIsNotStripped(t *testing.T) {
	meta1 := Parse("[!!Sully FANSUB] Android wa Keiken + ESPECIAL [WEB-DL 1080p x264 AAC SEM CENSURA] [PT-BR]")
	assertEqual(t, meta1.Title, "Android wa Keiken + ESPECIAL")

	meta2 := Parse("[FênixFansub] Android wa Keiken Ninzuu ni Hairimasu ka + Especial [PT-BR]")
	assertEqual(t, meta2.Title, "Android wa Keiken Ninzuu ni Hairimasu ka + Especial")
}

func TestParseTrailingEpisodeWithReleaseInformation(t *testing.T) {
	meta := Parse("One Piece 0122 Remaster (BDRip 1080p x264 AC3 Multi) - Ryūjin 竜神")
	assertEqual(t, meta.Title, "One Piece")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"122"})
}

func TestParseCombinedVolumeTokens(t *testing.T) {
	meta1 := Parse("One.Piece.Vol002.DVDRip.480p.x265.10bit.AAC.Multi-uP")
	assertEqual(t, meta1.Title, "One Piece")
	assertSliceEqual(t, meta1.VolumeNumber, []string{"2"})

	meta2 := Parse("One.Piece.Vol001.DVDRip.480p.x265.10bit.AAC.Multi-uP")
	assertEqual(t, meta2.Title, "One Piece")
	assertSliceEqual(t, meta2.VolumeNumber, []string{"1"})
}

func TestAudioChannelBaseOpus(t *testing.T) {
	meta := Parse("The.100.Girlfriends.Who.Really.Really.Really.Really.REALLY.Love.You.S01.v2.1080p.BluRay.Dual-Audio.Opus.2.0.x265-YURASUKA")
	assertEqual(t, meta.Title, "The 100 Girlfriends Who Really Really Really Really REALLY Love You")
	assertSliceEqual(t, meta.SeasonNumber, []string{"1"})
	assertSliceEqual(t, meta.EpisodeNumber, nil)
	assertEqual(t, meta.EpisodeTitle, "")
	assertEqual(t, meta.ReleaseGroup, "YURASUKA")
}

func TestTitleSuffixBeforeDashEpisode(t *testing.T) {
	meta := Parse("[FSP] Douluo Dalu II - Soul Land 2 - 151 [1080p]")
	assertEqual(t, meta.Title, "Douluo Dalu II - Soul Land 2")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"151"})
	assertEqual(t, meta.ReleaseGroup, "FSP")
}

func TestTitleSuffixBeforeDashEpisodeMultiDigit(t *testing.T) {
	meta := Parse("[Naruto-Kun.Hu] Eyeshield 21 - 001-008 [1080p]")
	assertEqual(t, meta.Title, "Eyeshield 21")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"1", "8"})
	assertEqual(t, meta.ReleaseGroup, "Naruto-Kun.Hu")
}

func TestEpisodeNumberWithFormatSuffix(t *testing.T) {
	meta := Parse("[DBD-Raws][海贼王 第20季 和之国篇/One Piece 20th Season Wano Arc][1006-1088TV+总集篇][1080P]")
	assertEqual(t, meta.Title, "海贼王")
	assertSliceEqual(t, meta.SeasonNumber, []string{"20"})
	assertSliceEqual(t, meta.EpisodeNumber, []string{"1006", "1088"})
	assertEqual(t, meta.ReleaseGroup, "DBD-Raws")
}

func TestEpisodeTitleWithMetadataBracket(t *testing.T) {
	meta := Parse("[Brazh] Monster Henshū 10 (REPACK) - La villa des roses - 1080p.MULTI.x264.mkv")
	assertEqual(t, meta.Title, "Monster Henshū")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"10"})
	assertEqual(t, meta.EpisodeTitle, "La villa des roses")
	assertEqual(t, meta.ReleaseGroup, "Brazh")
	assertSliceEqual(t, meta.ReleaseInformation, []string{"REPACK"})
}

func TestRawAsTitleBoundary(t *testing.T) {
	meta := Parse("[Anime Land] One Piece 1157 (WEBRip 1080p AV1 AAC) RAW [A16FC776].mkv")
	assertEqual(t, meta.Title, "One Piece")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"1157"})
	assertEqual(t, meta.EpisodeTitle, "")
	assertSliceContains(t, meta.Language, "RAW")
}

func TestFileExtensionAsTitleBoundary(t *testing.T) {
	meta := Parse("[shincaps] Suzumiya Haruhi no Yuuutsu - 04 (BS11 1920x1080 MPEG2 AAC).ts (w/ tsunami alert map)")
	assertEqual(t, meta.Title, "Suzumiya Haruhi no Yuuutsu")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"4"})
	assertEqual(t, meta.EpisodeTitle, "")
}

func TestYearRangeTVCollision(t *testing.T) {
	meta := Parse("魔法のプリンセス.ミンキーモモ.(夢を抱きしめて).Mahou.no.Princess.Minky.Momo.Yume.wo.Dakishimete.1991-1994.TV.S2+2OVA.BDRip.720p.x265.FLAC-CoolFansSub")
	assertEqual(t, meta.Title, "魔法のプリンセス ミンキーモモ. (夢を抱きしめて).Mahou no Princess Minky Momo Yume wo Dakishimete")
	assertEqual(t, meta.Year, "1991")
	assertSliceEqual(t, meta.SeasonNumber, []string{"2"})
	assertSliceEqual(t, meta.EpisodeNumber, nil)
	assertEqual(t, meta.EpisodeTitle, "")
	assertSliceContains(t, meta.Source, "TV")
}

func TestThaiTitleAlternativeSeparator(t *testing.T) {
	meta := Parse("Detective Conan Thai 875-1182 / ยอดนักสืบจิ๋วโคนัน (TrueVisionsNOW TH WEB-DL) [eva]")
	assertEqual(t, meta.Title, "Detective Conan Thai")
	assertSliceEqual(t, meta.EpisodeNumber, []string{"875", "1182"})
	assertEqual(t, meta.EpisodeTitle, "ยอดนักสืบจิ๋วโคนัน")
	assertEqual(t, meta.ReleaseGroup, "eva")
}
