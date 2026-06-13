# Miyo Parser

[![Go Reference](https://pkg.go.dev/badge/github.com/casss-d/miyo.svg)](https://pkg.go.dev/github.com/casss-d/miyo)
[![License](https://img.shields.io/github/license/casss-d/miyo.svg)](LICENSE)

Miyo Parser is a Go library designed to extract structured metadata from media and anime filenames. It parses complex titles to retrieve series names, seasons, episodes, release groups, technical video/audio specifications, and subtitle languages.

The parser uses a tokenization pipeline combined with bracket group analysis and contextual scoring heuristics to resolve ambiguities (e.g., distinguishing whether a number is a season, an episode, or part of the title itself).

---

## Features

- **Title Extraction:** Trims release groups, bracketed metadata, and structural delimiters to recover the canonical series title.
- **Sequence Parsing:** Identifies seasons, episodes, parts, and volumes across diverse patterns (such as standard `S01E02`, absolute numbers, ordinal notations, and Japanese counters like `第2期` / `01話`).
- **Context-Aware Heuristics:** Uses a localized scoring engine to evaluate neighboring tokens, helping avoid false positives on numbers within titles (e.g., *Blue Submarine No.6*).
- **Technical Metadata Extraction:**
  - **Video:** Resolution (e.g., `1080p`, `2160p`, `4K`), codecs (e.g., `HEVC`, `x265`, `H.264`), color depth, and HDR.
  - **Audio:** Codecs and formats (e.g., `FLAC`, `AAC`, `DDP5.1`, `TrueHD`, `Atmos`).
  - **Release Information:** Sources (`Blu-ray`, `WEB-DL`, `HDTV`), release groups, versions, and validation hashes (checksums).
- **Language Detection:** Detects audio and subtitle languages (e.g., `VOSTFR`, `Dual Audio`, `Multi-Subs`).

---

## Installation

To integrate Miyo Parser into your Go project:

```bash
go get github.com/casss-d/miyo
```

---

## Quick Start

### Basic Usage

Use the `Parse` function to quickly analyze a filename with default options:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/casss-d/miyo"
)

func main() {
	filename := "[SubsPlease] Sousou no Frieren - 14 [1080p].mkv"
	metadata := miyo.Parse(filename)

	fmt.Printf("Title: %s\n", metadata.Title)
	fmt.Printf("Episode: %s\n", metadata.EpisodeNumber[0])
	fmt.Printf("Resolution: %s\n", metadata.VideoResolution)
	fmt.Printf("Release Group: %s\n", metadata.ReleaseGroup)
	fmt.Printf("Extension: %s\n", metadata.FileExtension)
}
```

### Advanced Parsing with Options

You can customize parsing behavior or restrict specific scans using the `Options` struct:

```go
package main

import (
	"fmt"
	"github.com/casss-d/miyo"
)

func main() {
	filename := "[TaigaSubs]_Toradora!_(2008)_-_01v2_[1280x720_H.264_FLAC][1234ABCD].mkv"

	// Configure parsing constraints
	options := miyo.Options{
		YearMin: 2000,
		YearMax: 2030,
		// Turn off specific features if they are not needed
		DisableFileChecksum: false, 
	}

	metadata := miyo.ParseWithOptions(filename, options)
	fmt.Printf("Formatted Title: %s\n", metadata.FormattedTitle)
	fmt.Printf("Checksum: %s\n", metadata.FileChecksum)
}
```

---

## JSON Metadata Representation

When serialized, the `Metadata` struct produces structured, consumer-friendly output:

```json
{
  "file_name": "[Anime Time] Sword Art Online (S01+S02) [BD][Dual Audio][1080p][Batch].mkv",
  "title": "Sword Art Online",
  "formatted_title": "Sword Art Online",
  "season_number": ["1", "2"],
  "anime_type": ["Batch"],
  "audio_term": ["Dual Audio"],
  "file_extension": "mkv",
  "release_group": "Anime Time",
  "release_information": ["Batch"],
  "source": ["BD"],
  "video_resolution": "1080p",
  "series": [
    {
      "title": "Sword Art Online",
      "season": [
        { "number": "1" },
        { "number": "2" }
      ]
    }
  ]
}
```

---

## Architecture Overview

Miyo Parser divides processing into distinct, predictable steps:

1. **Tokenization:** Splits the raw input string into classified tokens (e.g., text, plain delimiters, context delimiters, and brackets).
2. **Bracket Analysis:** Matches open/close pairs to build structural `BracketGroup` contexts.
3. **Metadata Scanning:** Evaluates patterns and checks lexical entries to assign candidate metadata tags (such as resolution, audio formats, language, and sequences).
4. **Contextual & Heuristic Resolution:** Scores candidates numerically based on neighboring elements and position to separate legitimate title text from technical metadata.
5. **Title Recovery:** Dynamically determines boundaries, cleans up spacing and punctuation, and extracts both the primary and episode titles.
6. **Metadata Assembly:** Aggregates findings into the final structured `Metadata` object.

---

## Integration with Other Languages (JSON Worker)

The project includes a command-line utility, `cmd/miyo-parse-worker`, designed to run as a persistent background process. It reads filenames from `stdin` and writes JSON results to `stdout`, providing a clean path for integration with scripts written in languages like Python, Node.js, or Rust.

To start the worker:

```bash
go run ./cmd/miyo-parse-worker
```

---

## Testing

The library includes an extensive suite of tests validating diverse filename conventions. To run the tests:

```bash
go test -v ./...
```

For large-scale heuristic quality control, the repository includes a Python-based auditor (`local_testing/audit_parser.py`) that analyzes parsed outputs against diagnostic rules to flag suspicious parses (such as unstripped file extensions or metadata leaking into titles).

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
