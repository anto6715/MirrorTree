package mtreeapp

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/anto6715/goat/find"
)

type Config struct {
	Source   string
	Target   string
	MaxDepth int
	Filter   string
}

func Run(cfg Config) error {
	// Validation
	cfg, err := normalizeConfig(cfg)
	if err != nil {
		return err
	}

	// Apply options
	opts := find.DefaultOptions()
	opts.MaxDepth = cfg.MaxDepth
	opts.Filter = cfg.Filter

	// List source files
	sourceFiles, err := find.FindFilesWithDirs(cfg.Source, opts)
	if err != nil {
		return err
	}
	slog.Info("directories to mirror", "count", len(sourceFiles))

	// Build Mirror Plan
	fmt.Println()
	for dir, files := range sourceFiles {
		mirrorDir := strings.ReplaceAll(dir, cfg.Source, cfg.Target)
		fmt.Printf("Mirroring %q -> %q\n", dir, mirrorDir)

		for _, file := range files {
			mirrorFile := filepath.Join(mirrorDir, filepath.Base(file))
			fmt.Printf("    %q -> %q\n", file, mirrorFile)
		}
		fmt.Println()
	}

	return nil
}

func normalizeConfig(cfg Config) (Config, error) {
	if cfg.MaxDepth < -1 {
		return Config{}, fmt.Errorf("max depth must be -1 or greater")
	}

	if cfg.Filter == "" {
		return Config{}, fmt.Errorf("filter cannot be empty")
	}

	// Source validation
	source, err := canonicalDir(cfg.Source)
	if err != nil {
		return Config{}, fmt.Errorf("invalid target %q: %w", cfg.Source, err)
	}

	if pathExists(cfg.Target) {
		return Config{}, fmt.Errorf("target %q already exists", cfg.Target)
	}

	normalized := cfg
	normalized.Source = source

	return normalized, nil
}

func canonicalDir(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", path)
	}

	return resolved, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
