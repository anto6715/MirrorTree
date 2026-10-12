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

	formatOpts := formatOptions{color: cfg.Color, apply: cfg.Apply}
	if err := writeMirrorPlan(os.Stdout, plan, formatOpts); err != nil {
		return fmt.Errorf("failed to print mirror plan: %w", err)
	}

	// Dry run, or nothing matched: create nothing
	if !cfg.Apply || len(plan.dirs) == 0 {
		return nil
	}

	dirs, files, err := executePlan(plan)
	if err != nil {
		return fmt.Errorf("mirror %q incomplete, created %s and %s before failing: %w",
			plan.target,
			plural(dirs, "directory", "directories"),
			plural(files, "file", "files"),
			err,
		)
	}

	if err := writeMirrorCreated(os.Stdout, plan, formatOpts); err != nil {
		return fmt.Errorf("failed to print result: %w", err)
	}

	return nil
}
