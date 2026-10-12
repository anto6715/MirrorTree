package mtreeapp

import (
	"fmt"
	"os"

	"github.com/anto6715/goat/find"
)

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

	// Build Mirror Plan
	plan, err := buildMirrorPlan(cfg, sourceFiles)
	if err != nil {
		return err
	}

	if err := writeMirrorPlan(os.Stdout, plan, formatOptions{color: cfg.Color}); err != nil {
		return fmt.Errorf("failed to print mirror plan: %w", err)
	}

	return nil
}
