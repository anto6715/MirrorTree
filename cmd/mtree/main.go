package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/anto6715/MirrorTree/internal/logging"
	"github.com/anto6715/MirrorTree/internal/mtreeapp"
)

type cli struct {
	Path     string `arg:"" name:"path" help:"Root directory to mirror"`
	Dest     string `arg:"" name:"dest" help:"Destination directory"`
	MaxDepth int    `short:"L" name:"max-depth" default:"-1" help:"Maximum depth to mirror (-1 for unlimited)"`
	Filter   string `short:"f" name:"filter" default:"*" help:"Filter pattern to match files"`
	NoColor  bool   `name:"no-color" help:"Disable colored output"`
	Apply    bool   `name:"apply" help:"Create the mirror (without it, only print the plan)"`
}

func main() {
	// Initialize logger
	logger := logging.New()
	slog.SetDefault(logger)

	// CLI
	var args cli
	kong.Parse(
		&args,
		kong.Name("Mirror Tree"),
		kong.Description("Mirror a directory tree to a destination directory"),
		kong.UsageOnError(),
	)

	err := mtreeapp.Run(mtreeapp.Config{
		Source:   args.Path,
		Target:   args.Dest,
		MaxDepth: args.MaxDepth,
		Filter:   args.Filter,
		Color:    useColor(args.NoColor),
		Apply:    args.Apply,
	})
	if err != nil {
		slog.Error("Failed to run", "error", err)
		os.Exit(1)
	}
}

// useColor enables colors only when stdout is a terminal, and never when
// --no-color or NO_COLOR (https://no-color.org) is set.
func useColor(noColor bool) bool {
	if noColor || os.Getenv("NO_COLOR") != "" {
		return false
	}

	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
