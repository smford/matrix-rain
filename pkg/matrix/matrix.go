package matrix

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
)

// Config encapsulates runtime parameters for the Matrix rain simulation.
type Config struct {
	FPS         int
	Density     int     // 1 - 100
	SpeedScale  float64 // Multiplier for drop speeds (e.g., 0.5x - 3.0x)
	ThemeName   string
	CharSetName string
	BoldHead    bool
}

// DefaultConfig returns production-ready default settings.
func DefaultConfig() Config {
	return Config{
		FPS:         30,
		Density:     50,
		SpeedScale:  1.0,
		ThemeName:   "green",
		CharSetName: string(CharSetMatrix),
		BoldHead:    true,
	}
}

// Cell represents one terminal character grid cell.
type Cell struct {
	Rune  rune
	Color string
}

// Engine drives the matrix rain simulation and rendering loop.
type Engine struct {
	cfg        Config
	mu         sync.Mutex
	width      int
	height     int
	columns    []*Column
	pool       []rune
	theme      Theme
	themeIndex int
	paused     bool
	writer     *bufio.Writer
	stdinFd    int
	stdoutFd   int
	termState  *term.State
}

// NewEngine constructs a matrix simulation engine.
func NewEngine(cfg Config, out io.Writer) *Engine {
	if cfg.FPS <= 0 {
		cfg.FPS = 30
	}
	if cfg.Density <= 0 || cfg.Density > 100 {
		cfg.Density = 50
	}
	if cfg.SpeedScale <= 0 {
		cfg.SpeedScale = 1.0
	}

	theme := GetThemeByName(cfg.ThemeName)
	themeIdx := 0
	for i, t := range AllThemes {
		if t.Name == theme.Name {
			themeIdx = i
			break
		}
	}

	return &Engine{
		cfg:        cfg,
		pool:       GetCharPool(CharSet(cfg.CharSetName)),
		theme:      theme,
		themeIndex: themeIdx,
		writer:     bufio.NewWriterSize(out, 64*1024), // 64KB render buffer
		stdinFd:    int(os.Stdin.Fd()),
		stdoutFd:   int(os.Stdout.Fd()),
	}
}

// Run starts the engine in full-screen terminal raw mode, blocking until exit signal or user quit.
func (e *Engine) Run(ctx context.Context) error {
	// Terminal initialization
	if term.IsTerminal(e.stdinFd) {
		oldState, err := term.MakeRaw(e.stdinFd)
		if err != nil {
			return fmt.Errorf("failed to enable terminal raw mode: %w", err)
		}
		e.termState = oldState
	}

	// Defensive cleanup guarantee: restore terminal state on return or panic
	defer e.Cleanup()

	// Switch to alternate screen buffer, hide cursor, clear screen
	_, _ = fmt.Fprint(e.writer, "\x1b[?1049h\x1b[?25l\x1b[2J")
	_ = e.writer.Flush()

	// Initial dimensions
	w, h, err := term.GetSize(e.stdoutFd)
	if err != nil || w <= 0 || h <= 0 {
		w, h = 80, 24
	}
	e.resize(w, h)

	// OS signal handling
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Window resize signal (SIGWINCH)
	winchChan := make(chan os.Signal, 2)
	notifyResize(winchChan)

	// Non-blocking keyboard input stream
	keyChan := make(chan byte, 16)
	go e.listenKeys(keyChan)

	// Simulation ticker
	frameDuration := time.Second / time.Duration(e.cfg.FPS)
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-sigChan:
			return nil

		case <-winchChan:
			nw, nh, err := term.GetSize(e.stdoutFd)
			if err == nil && nw > 0 && nh > 0 {
				e.mu.Lock()
				e.resize(nw, nh)
				e.mu.Unlock()
			}

		case key := <-keyChan:
			if e.handleInput(key) {
				return nil // Exit requested
			}

		case <-ticker.C:
			e.mu.Lock()
			if !e.paused {
				e.step()
			}
			e.render()
			e.mu.Unlock()
		}
	}
}

// resize updates internal dimensions and adapts column slices.
func (e *Engine) resize(newW, newH int) {
	e.width = newW
	e.height = newH

	currentCols := len(e.columns)
	if newW > currentCols {
		for i := currentCols; i < newW; i++ {
			e.columns = append(e.columns, NewColumn())
		}
	} else if newW < currentCols {
		e.columns = e.columns[:newW]
	}
}

// step advances all columns by one frame.
func (e *Engine) step() {
	for _, col := range e.columns {
		col.Update(e.height, e.pool, e.cfg.Density, e.cfg.SpeedScale)
	}
}

// render computes the frame buffer and flushes optimized ANSI output.
func (e *Engine) render() {
	if e.width <= 0 || e.height <= 0 {
		return
	}

	// Build sparse/dense grid for this frame
	// Allocating/reusing a grid slice
	grid := make([][]Cell, e.height)
	for r := range grid {
		grid[r] = make([]Cell, e.width)
	}

	// Populate drops into grid
	for x, col := range e.columns {
		if x >= e.width {
			continue
		}
		for _, drop := range col.Drops {
			headRow := int(drop.Y)
			for dist := 0; dist < drop.Length; dist++ {
				row := headRow - dist
				if row >= 0 && row < e.height {
					// Drop with brightest/closest head takes precedence
					if dist < len(drop.Glyphs) {
						color := e.theme.ColorAtPosition(dist, drop.Length, e.cfg.BoldHead)
						grid[row][x] = Cell{
							Rune:  drop.Glyphs[dist],
							Color: color,
						}
					}
				}
			}
		}
	}

	// Move cursor to top-left home position (1,1)
	_, _ = e.writer.WriteString("\x1b[H")

	// Render cells with ANSI color change deduplication for performance
	currentColor := ""
	for y := 0; y < e.height; y++ {
		for x := 0; x < e.width; x++ {
			cell := grid[y][x]
			if cell.Rune == 0 || cell.Rune == ' ' {
				if currentColor != "" {
					_, _ = e.writer.WriteString("\x1b[0m")
					currentColor = ""
				}
				_ = e.writer.WriteByte(' ')
			} else {
				if cell.Color != currentColor {
					_, _ = e.writer.WriteString(cell.Color)
					currentColor = cell.Color
				}
				_, _ = e.writer.WriteString(string(cell.Rune))
			}
		}
		// Reset color at the end of each row if needed
		if currentColor != "" && y < e.height-1 {
			_, _ = e.writer.WriteString("\x1b[0m")
			currentColor = ""
		}
		if y < e.height-1 {
			_, _ = e.writer.WriteString("\r\n")
		}
	}

	if currentColor != "" {
		_, _ = e.writer.WriteString("\x1b[0m")
	}

	_ = e.writer.Flush()
}

// handleInput processes interactive single-byte commands. Returns true to signal exit.
func (e *Engine) handleInput(key byte) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch key {
	case 'q', 'Q', 3, 27: // 'q', 'Q', Ctrl+C (0x03), ESC (0x1b)
		return true

	case ' ': // Pause / resume
		e.paused = !e.paused

	case 'c', 'C': // Cycle theme
		e.themeIndex = (e.themeIndex + 1) % len(AllThemes)
		e.theme = AllThemes[e.themeIndex]

	case '+', '=': // Increase speed & density
		if e.cfg.SpeedScale < 3.0 {
			e.cfg.SpeedScale += 0.2
		}
		if e.cfg.Density <= 90 {
			e.cfg.Density += 10
		}

	case '-', '_': // Decrease speed & density
		if e.cfg.SpeedScale > 0.3 {
			e.cfg.SpeedScale -= 0.2
		}
		if e.cfg.Density >= 20 {
			e.cfg.Density -= 10
		}

	case 'r', 'R': // Reset rain
		for i := range e.columns {
			e.columns[i] = NewColumn()
		}
	}

	return false
}

// listenKeys streams raw byte keystrokes from stdin.
func (e *Engine) listenKeys(ch chan<- byte) {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}
		ch <- buf[0]
	}
}

// Cleanup restores the terminal to standard mode, leaves alternate buffer, and unhides cursor.
func (e *Engine) Cleanup() {
	// Restore cursor, leave alternate screen buffer, clear attributes
	_, _ = fmt.Fprint(e.writer, "\x1b[0m\x1b[?25h\x1b[?1049l")
	_ = e.writer.Flush()

	if e.termState != nil {
		_ = term.Restore(e.stdinFd, e.termState)
		e.termState = nil
	}
}
