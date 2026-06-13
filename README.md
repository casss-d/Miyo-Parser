# Miyo

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/casss-d/miyo-parser.svg)](https://pkg.go.dev/github.com/casss-d/miyo-parser)

Miyo is a deterministic, high-performance anime filename parser for Go. 

Parsing anime media filenames—where naming conventions are notoriously chaotic and inconsistent—is incredibly difficult. Unlike traditional regex-based approaches that break under the weight of edge cases, Miyo utilizes a **heuristic scoring engine**. Confidence accumulates from multiple signals (position, context, keywords, and patterns) to handle the wild variety of fansub naming conventions gracefully.


## Features

* **Heuristic Scoring:** Tokens aren't guessed instantly; they accumulate possibilities and scores, resolving ambiguities perfectly (e.g., is "OP" a release group, an Opening Theme, or part of "ONE PIECE"?).
* **Context-Aware:** Intelligently differentiates between "English" as an audio language, a subtitle language, or a word in the title (like *English Teacher*).
* **Full-Width Rune Normalization:** Automatically handles Japanese full-width alphanumeric characters (e.g., `１０８０ｐ` → `1080p`).
* **Zero Dependencies:** Built entirely with the Go standard library. Fast, safe, and portable.
* **Debuggable:** Built-in debug tracing to output the exact token states, possibilities, and scores as JSON.

## Installation

```bash
go get github.com/casss-d/miyo-parser
```

## Quick Start

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/casss-d/miyo-parser"
)

func main() {
    filename := "[TaigaSubs]_Toradora!_(2008)_-_01v2_-_Tiger_and_Dragon_[1280x720_H.264_FLAC][1234ABCD].mkv"
    
    // Parse the filename into a structured Metadata object
    meta := miyo.Parse(filename)

    fmt.Println("Title:", meta.Title)                     // Toradora!
    fmt.Println("Year:", meta.Year)                       // 2008
    fmt.Println("Episode:", meta.EpisodeNumber[0])        // 1
    fmt.Println("Resolution:", meta.VideoResolution)      // 1280x720
    fmt.Println("Release Group:", meta.ReleaseGroup)      // TaigaSubs
    fmt.Println("File Checksum:", meta.FileChecksum)      // 1234ABCD
}
```

## Advanced Usage

### Custom Configuration

You can customize the parsing behavior by passing a `miyo.Options` struct. You can disable specific parsers to speed up execution or avoid false positives if you only care about specific fields.

```go
options := miyo.Options{
    DisableTitle:       false,
    DisableEpisode:     false,
    ParseFileChecksum:  true,
    YearMin:            1960,
    YearMax:            2030,
}

meta := miyo.ParseWithOptions("[Group] Title - 01 [1080p].mkv", options)
```

### Parsing with Filepaths

If you have a full directory path, you can use `ParsePath` to automatically extract context from the parent folders.

```go
meta := miyo.ParsePath("/anime/Toradora/[Group] Toradora! - 01.mkv", miyo.Options{})
```

### Debugging

If a filename isn't parsing the way you expect, Miyo includes a powerful debug mode that returns the exact scoring logic used under the hood.

```go
meta, debugInfo := miyo.ParseDebug("[Group] Title - 01.mkv", miyo.Options{})

// debugInfo contains Tokens, Matches, and the scoring map for each possibility
debugJSON, _ := json.MarshalIndent(debugInfo, "", "  ")
fmt.Println(string(debugJSON))
```

## Output Structure

Miyo returns a `*miyo.Metadata` struct containing flat fields for easy access, as well as a structured `Series` slice for complex media. 

| Field | Type | Example |
|---|---|---|
| `FileName` | `string` | `"[Group] Title - 01.mkv"` |
| `Title` | `string` | `"Title"` |
| `ReleaseGroup` | `string` | `"Group"` |
| `EpisodeNumber` | `[]string` | `["1"]` |
| `SeasonNumber` | `[]string` | `["2"]` |
| `EpisodeTitle` | `string` | `"The Beginning"` |
| `VideoResolution` | `string` | `"1080p"` |
| `VideoTerm` | `[]string` | `["H.264", "x265"]` |
| `AudioTerm` | `[]string` | `["FLAC", "Dual Audio"]` |
| `Subtitles` | `[]string` | `["Multi-Subs"]` |
| `Language` | `[]string` | `["English", "Jap"]` |
| `FileChecksum` | `string` | `"1234ABCD"` |
| `FileExtension` | `string` | `"mkv"` |
| `Series` | `[]miyo.SeriesInfo`| Structured nested data (for batches/ranges) |

*Note: Slices are used for fields that can have multiple values (e.g., dual audio tracks, episode ranges).*

## How does it work?

Anime filenames are notoriously inconsistent. Regex-based parsers fail because they cannot cover the combinatorial explosion of naming conventions. Miyo processes filenames through a 6-stage pipeline:

1. **Tokenize:** Splits input into tokens, normalizing full-width runes, and detecting brackets, delimiters, and text boundaries.
2. **Analyze Groups:** Maps bracket pairs `()`, `[]`, `「」` to establish isolated metadata zones.
3. **Lexicon Matching:** Matches tokens against known technical keywords (e.g., "HEVC", "WEB-DL"), assigning base probabilities.
4. **Context Scanning:** Pattern-based rules scan for dynamic data (checksums, dates, sequence numbers, Japanese episode markers).
5. **Score & Resolve:** Context-aware rules adjust confidence based on position and neighbors. Ties are broken using a strict rank hierarchy. The highest-scoring possibility wins.
6. **Compose:** Assembles the resolved tokens into the final, clean `Metadata` struct.

## License

Miyo is licensed under the **Mozilla Public License 2.0 (MPL 2.0)**.

**What does this mean practically?**
* **You CAN** use Miyo in any project—open-source, closed-source, commercial, or personal. You do not have to open-source your application.
* **You MUST** share any modifications you make to Miyo itself. If you fix a bug, improve the scoring logic, or add new lexicon terms to Miyo's source files, you are legally required to publish those changes back to the community under the same MPL 2.0 license. 

For the exact legal terms, see the `LICENSE` file.
