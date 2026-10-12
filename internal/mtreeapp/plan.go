package mtreeapp

import (
	"maps"
	"path/filepath"
	"slices"
)

// buildMirrorPlan maps every directory and file found under cfg.Source to its
// mirror under cfg.Target. Directories and files are sorted so the plan is
// deterministic regardless of map iteration order.
func buildMirrorPlan(cfg Config, sourceFiles map[string][]string) (mirrorPlan, error) {
	plan := mirrorPlan{source: cfg.Source, target: cfg.Target, filter: cfg.Filter}

	for _, dir := range slices.Sorted(maps.Keys(sourceFiles)) {
		_, targetDir, err := mirrorPath(cfg.Source, cfg.Target, dir)
		if err != nil {
			return mirrorPlan{}, err
		}
		entry := mirrorDir{source: dir, target: targetDir}

		// do not reorder the caller's slice
		files := slices.Clone(sourceFiles[dir])
		slices.Sort(files)
		for _, file := range files {
			rel, targetFile, err := mirrorPath(cfg.Source, cfg.Target, file)
			if err != nil {
				return mirrorPlan{}, err
			}
			entry.files = append(entry.files, mirrorFile{source: file, target: targetFile, rel: rel})
		}

		plan.dirs = append(plan.dirs, entry)
	}

	return plan, nil
}

// mirrorPath returns path relative to source, and its location under target
// keeping that same relative position.
func mirrorPath(source, target, path string) (rel string, mirrored string, err error) {
	rel, err = filepath.Rel(source, path)
	if err != nil {
		return "", "", err
	}
	return rel, filepath.Join(target, rel), nil
}
