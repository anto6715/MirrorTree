package mtreeapp

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

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
		mirrorDir, err := getMirrorDir(cfg.Source, cfg.Target, dir)
		if err != nil {
			return err
		}
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
	if _, err := filepath.Match(cfg.Filter, ""); err != nil {
		return Config{}, fmt.Errorf("invalid filter %q: %w", cfg.Filter, err)
	}

	if cfg.Filter == "" {
		return Config{}, fmt.Errorf("filter cannot be empty")
	}

	if cfg.MaxDepth < -1 {
		return Config{}, fmt.Errorf("max depth must be -1 or greater")
	}

	// Source validation
	source, err := canonicalDir(cfg.Source)
	if err != nil {
		return Config{}, fmt.Errorf("invalid source %q: %w", cfg.Source, err)
	}

	// Raise error if target already exists or in case of other errors
	_, err = os.Lstat(cfg.Target)
	switch {
	case err == nil:
		return Config{}, fmt.Errorf("target %q already exists", cfg.Target)
	case !errors.Is(err, fs.ErrNotExist):
		return Config{}, fmt.Errorf("checking target %q: %w", cfg.Target, err)
	}

	if cfg.Target == "" {
		return Config{}, fmt.Errorf("target cannot be empty")
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

func getMirrorDir(source, target, toMirror string) (string, error) {
	rel, err := filepath.Rel(source, toMirror)
	if err != nil {
		return "", err
	}
	return filepath.Join(target, rel), nil
}
