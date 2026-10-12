package mtreeapp

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Config struct {
	Source   string
	Target   string
	MaxDepth int
	Filter   string
	Color    bool
}

func normalizeConfig(cfg Config) (Config, error) {
	if cfg.Filter == "" {
		return Config{}, fmt.Errorf("filter cannot be empty")
	}

	if _, err := filepath.Match(cfg.Filter, ""); err != nil {
		return Config{}, fmt.Errorf("invalid filter %q: %w", cfg.Filter, err)
	}

	if cfg.MaxDepth < -1 {
		return Config{}, fmt.Errorf("max depth must be -1 or greater")
	}

	// Source validation
	source, err := canonicalDir(cfg.Source)
	if err != nil {
		return Config{}, fmt.Errorf("invalid source %q: %w", cfg.Source, err)
	}

	if cfg.Target == "" {
		return Config{}, fmt.Errorf("target cannot be empty")
	}

	// Raise error if target already exists or in case of other errors
	_, err = os.Lstat(cfg.Target)
	switch {
	case err == nil:
		return Config{}, fmt.Errorf("target %q already exists", cfg.Target)
	case !errors.Is(err, fs.ErrNotExist):
		return Config{}, fmt.Errorf("checking target %q: %w", cfg.Target, err)
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
