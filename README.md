# matrix-rain

High-performance, authentic *The Matrix* digital rain simulator for the terminal, written in Go.

![Matrix Rain](https://raw.githubusercontent.com/smford/matrix-rain/main/screenshot.png) <!-- optional preview -->

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

## Installation & Build

### Prerequisites
- Go 1.24+

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
./bin/matrix-rain
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
./bin/matrix-rain -density 75 -fps 60

# Amber CRT phosphor style
./bin/matrix-rain -color amber -charset matrix

# Binary rain in cyberpunk cyan
./bin/matrix-rain -color cyan -charset binary
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
├── cmd/
│   └── matrix-rain/
│       └── main.go         # CLI parsing, signal trapping, entrypoint
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
├── Makefile                # Build, test, lint, install
└── README.md
```

## License

MIT