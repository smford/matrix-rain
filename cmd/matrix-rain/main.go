package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/smford/matrix-rain/pkg/matrix"
	"golang.org/x/term"
)

var (
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	fps := flag.Int("fps", 30, "Target frame rate per second (10-120)")
	density := flag.Int("density", 50, "Rain drop spawn density percentage (1-100)")
	speed := flag.Float64("speed", 1.0, "Rain speed multiplier (0.2 - 3.0)")
	themeName := flag.String("color", "green", "Color scheme: green, cyan, amber, red, white, rainbow")
	charSetName := flag.String("charset", "matrix", "Character set: matrix, ascii, binary, hex")
	boldHead := flag.Bool("bold", true, "Highlight leading rain character in bold")
	showVersion := flag.Bool("version", false, "Display version and build information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Matrix Rain - Digital Rain Terminal Simulator\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nControls:\n")
		fmt.Fprintf(os.Stderr, "  q, Esc, Ctrl+C : Quit\n")
		fmt.Fprintf(os.Stderr, "  Space          : Pause / Resume\n")
		fmt.Fprintf(os.Stderr, "  c              : Cycle color schemes\n")
		fmt.Fprintf(os.Stderr, "  + / -          : Increase / Decrease density\n")
		fmt.Fprintf(os.Stderr, "  r              : Reset rain streams\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("matrix-rain v%s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Validate terminal standard output
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Fprintln(os.Stderr, "Error: matrix-rain must be run inside an interactive terminal.")
		os.Exit(1)
	}

	cfg := matrix.Config{
		FPS:         *fps,
		Density:     *density,
		SpeedScale:  *speed,
		ThemeName:   *themeName,
		CharSetName: *charSetName,
		BoldHead:    *boldHead,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	engine := matrix.NewEngine(cfg, os.Stdout)

	// Ensure cleanup on panic
	defer func() {
		if r := recover(); r != nil {
			engine.Cleanup()
			panic(r)
		}
	}()

	if err := engine.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Engine runtime error: %v\n", err)
		os.Exit(1)
	}
}
