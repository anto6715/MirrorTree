package mtreeapp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeConfigResolvesSourceSymlink(t *testing.T) {
	root := t.TempDir()
	source := makeTestDir(t, filepath.Join(root, "source"))
	sourceLink := filepath.Join(root, "source-link")
	if err := os.Symlink(source, sourceLink); err != nil {
		t.Fatalf("create source symlink: %v", err)
	}
	wantSource, err := filepath.EvalSymlinks(source)
	if err != nil {
		t.Fatalf("resolve source: %v", err)
	}

	cfg := Config{
		Source:   sourceLink,
		Target:   filepath.Join(root, "mirror"),
		MaxDepth: -1,
		Filter:   "*",
	}

	got, err := normalizeConfig(cfg)
	if err != nil {
		t.Fatalf("normalizeConfig returned an error: %v", err)
	}
	if got.Source != wantSource {
		t.Errorf("Source = %q, want %q", got.Source, wantSource)
	}
	if got.Target != cfg.Target {
		t.Errorf("Target = %q, want %q", got.Target, cfg.Target)
	}
}

func TestNormalizeConfigCleansTarget(t *testing.T) {
	root := t.TempDir()
	source := makeTestDir(t, filepath.Join(root, "source"))
	target := filepath.Join(root, "mirror")

	got, err := normalizeConfig(Config{Source: source, Target: target + "/", MaxDepth: -1, Filter: "*"})
	if err != nil {
		t.Fatalf("normalizeConfig returned an error: %v", err)
	}
	if got.Target != target {
		t.Errorf("Target = %q, want %q", got.Target, target)
	}
}

func TestNormalizeConfigRejectsInvalidValues(t *testing.T) {
	root := t.TempDir()
	source := makeTestDir(t, filepath.Join(root, "source"))
	file := writeTestFile(t, filepath.Join(root, "file.txt"))
	existing := makeTestDir(t, filepath.Join(root, "existing"))
	dangling := filepath.Join(root, "dangling")
	if err := os.Symlink(filepath.Join(root, "missing"), dangling); err != nil {
		t.Fatalf("create dangling symlink: %v", err)
	}
	target := filepath.Join(root, "mirror")

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "empty filter",
			cfg:  Config{Source: source, Target: target, MaxDepth: -1},
		},
		{
			name: "malformed filter",
			cfg:  Config{Source: source, Target: target, MaxDepth: -1, Filter: "["},
		},
		{
			name: "max depth below unlimited sentinel",
			cfg:  Config{Source: source, Target: target, MaxDepth: -2, Filter: "*"},
		},
		{
			name: "empty source",
			cfg:  Config{Target: target, MaxDepth: -1, Filter: "*"},
		},
		{
			name: "missing source",
			cfg:  Config{Source: filepath.Join(root, "missing"), Target: target, MaxDepth: -1, Filter: "*"},
		},
		{
			name: "source is a file",
			cfg:  Config{Source: file, Target: target, MaxDepth: -1, Filter: "*"},
		},
		{
			name: "empty target",
			cfg:  Config{Source: source, MaxDepth: -1, Filter: "*"},
		},
		{
			name: "target already exists",
			cfg:  Config{Source: source, Target: existing, MaxDepth: -1, Filter: "*"},
		},
		{
			name: "target is a dangling symlink",
			cfg:  Config{Source: source, Target: dangling, MaxDepth: -1, Filter: "*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := normalizeConfig(tt.cfg); err == nil {
				t.Fatal("normalizeConfig returned nil error")
			}
		})
	}
}

func makeTestDir(t *testing.T, path string) string {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("create directory %q: %v", path, err)
	}
	return path
}

func writeTestFile(t *testing.T, path string) string {
	t.Helper()

	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatalf("create file %q: %v", path, err)
	}
	return path
}
