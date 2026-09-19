# matrix-rain

High-performance, authentic *The Matrix* digital rain simulator for the terminal, written in Go.

[![CI](https://github.com/smford/matrix-rain/actions/workflows/ci.yml/badge.svg)](https://github.com/smford/matrix-rain/actions/workflows/ci.yml)
[![Release](https://github.com/smford/matrix-rain/actions/workflows/release.yml/badge.svg)](https://github.com/smford/matrix-rain/releases)
[![Pages](https://github.com/smford/matrix-rain/actions/workflows/pages.yml/badge.svg)](https://smford.github.io/matrix-rain/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Live Demo & Documentation**: [https://smford.github.io/matrix-rain/](https://smford.github.io/matrix-rain/)

---

## Features

- **Authentic Matrix Aesthetics**:
  - Half-width Katakana glyphs (single column width for crisp alignment), numerals, Latin characters, and symbols.
  - Glowing head characters with customizable 24-bit truecolor fading gradients.
  - In-place glyph mutations as characters cascade down the screen.
- **SRE & Production Engineering Quality**:
  - **Zero Terminal Corruption Guarantee**: Strict cleanup hooks for alternate screen buffer restoration, cursor recovery, and raw mode reset on exit, cancellation, or unexpected panics.
  - **Low Overhead**: Batch rendering with ANSI escape deduplication; minimal CPU and memory allocations even at high framerates.
  - **Dynamic Terminal Resizing**: Real-time terminal geometry adaptation via POSIX `SIGWINCH`.
  - **Non-TTY Protection**: Detects pipes or redirected streams and prevents ANSI escape pollution.
- **Interactive Controls**:
  - Live color palette cycling (Classic Green, Matrix Reloaded Cyan, Amber CRT, Blood Red, Monochrome, Rainbow).
  - Real-time pause, resume, density adjustments, and stream reset.

---

## Installation

### Via Homebrew (Recommended)

Install directly using the [smford/homebrew-tap](https://github.com/smford/homebrew-tap):

```bash
brew tap smford/tap
brew install matrix-rain
```

Or as a single command:

```bash
brew install smford/tap/matrix-rain
```

### Via Go Install

```bash
go install github.com/smford/matrix-rain/cmd/matrix-rain@latest
```

### Pre-built Binaries

Download pre-compiled multi-architecture binaries (macOS Apple Silicon/Intel, Linux ARM64/AMD64, Windows) directly from the [GitHub Releases](https://github.com/smford/matrix-rain/releases) page.

### Build from source

```bash
git clone https://github.com/smford/matrix-rain.git
cd matrix-rain
make build
```

The executable will be located at `./bin/matrix-rain`.

---

## Usage

Run with defaults:

```bash
matrix-rain
```

### CLI Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `-color` | string | `green` | Palette: `green`, `cyan`, `amber`, `red`, `white`, `rainbow` |
| `-charset` | string | `matrix` | Character set: `matrix`, `ascii`, `binary`, `hex` |
| `-density` | int | `50` | Stream spawn density percentage (1–100) |
| `-fps` | int | `30` | Target frame rate (10–120) |
| `-speed` | float | `1.0` | Speed multiplier (0.2–3.0) |
| `-bold` | bool | `true` | Highlight leading characters in bold |
| `-version` | bool | `false` | Display version and build metadata |

#### Examples

```bash
# Classic Matrix with high density at 60 FPS
matrix-rain -density 75 -fps 60

# Amber CRT phosphor style
matrix-rain -color amber -charset matrix

# Binary rain in cyberpunk cyan
matrix-rain -color cyan -charset binary
```

---

## Keyboard Controls

| Key | Action |
|---|---|
| `q`, `Esc`, `Ctrl+C` | Gracefully quit and restore terminal |
| `Space` | Pause / Resume animation |
| `c` | Cycle color themes |
| `+` / `-` | Increase / Decrease spawn density |
| `r` | Reset active rain streams |

---

## CI/CD, Semantic Versioning & Release Pipeline

1. **Automated Semantic Versioning**:
   - Commits following [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `feat!:`) determine automatic semantic version bumps (`patch`, `minor`, `major`).
   - Manual releases can be triggered via `workflow_dispatch` or git tag push (`git tag vX.Y.Z`).
2. **Multi-Architecture Release Pipeline**:
   - GoReleaser builds and packages binaries for Darwin (ARM64 & AMD64), Linux (ARM64 & AMD64), and Windows.
   - Generates SHA256 checksums and automated release notes on GitHub Releases.
3. **Automated Homebrew Tap Publishing**:
   - The GoReleaser pipeline automatically pushes the updated formula and hashes to [smford/homebrew-tap](https://github.com/smford/homebrew-tap).
4. **GitHub Pages Deployment**:
   - Automatically builds and publishes the landing page at `https://smford.github.io/matrix-rain/` from `docs/`.

---

## Testing & Quality Assurance

```bash
# Run unit tests
make test

# Run tests with race detection
make test-race

# Run linter
make lint
```

## Architecture

```
matrix-rain/
├── .github/
│   └── workflows/
│       ├── ci.yml          # Tests, linting, race detection
│       ├── pages.yml       # GitHub Pages automated deployment
│       └── release.yml     # SemVer tagging, GoReleaser, Homebrew publish
├── cmd/
│   └── matrix-rain/
│       └── main.go         # CLI parsing, signal trapping, entrypoint
├── docs/
│   └── index.html          # Interactive landing page hosted on GitHub Pages
├── Formula/
│   └── matrix-rain.rb      # Reference Homebrew formula for smford/homebrew-tap
├── pkg/
│   └── matrix/
│       ├── chars.go        # Character sets & entropy generation
│       ├── chars_test.go   # Charset unit tests
│       ├── column.go       # Drop state, mutation, and column lifecycle
│       ├── column_test.go  # Column simulation unit tests
│       ├── matrix.go       # Engine loop, ANSI rendering, raw mode
│       ├── matrix_test.go  # Engine & resize unit tests
│       ├── signal_unix.go  # Unix SIGWINCH resize listener
│       ├── signal_windows.go # Windows resize fallback
│       ├── theme.go        # 24-bit Truecolor palettes and gradient calculation
│       └── theme_test.go   # Theme unit tests
├── .goreleaser.yaml        # GoReleaser v2 release and Homebrew formula config
├── Makefile                # Build, test, lint, install
└── README.md
```

## License

MIT