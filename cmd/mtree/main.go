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
	})
	if err != nil {
		slog.Error("Failed to run", "error", err)
		os.Exit(1)
	}
}
