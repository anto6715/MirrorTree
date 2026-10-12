package mtreeapp

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"testing"
)

func TestExecutePlanCreatesMirror(t *testing.T) {
	// parents out/2026 do not exist yet
	target := filepath.Join(t.TempDir(), "out", "2026", "mirror")
	plan := testPlanAt(t, target, "a.nc", "database/x.nc", "sub/b.txt", "sub/deep/c.nc")

	dirs, files, err := executePlan(plan)
	if err != nil {
		t.Fatalf("executePlan returned an error: %v", err)
	}
	if dirs != 3 || files != 4 {
		t.Errorf("created %d dirs, %d files, want 3 dirs, 4 files", dirs, files)
	}

	want := map[string]bool{ // rel path -> is directory
		"a.nc":          false,
		"database":      true,
		"database/x.nc": false,
		"sub":           true,
		"sub/b.txt":     false,
		"sub/deep":      true,
		"sub/deep/c.nc": false,
	}
	got := map[string]bool{}
	err = filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == target {
			return nil
		}

		rel, err := filepath.Rel(target, path)
		if err != nil {
			return err
		}
		got[rel] = d.IsDir()

		info, err := d.Info()
		if err != nil {
			return err
		}
		if !d.IsDir() && (!info.Mode().IsRegular() || info.Size() != 0) {
			t.Errorf("%s: mode %v, size %d, want empty regular file", rel, info.Mode(), info.Size())
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk mirror: %v", err)
	}

	if !maps.Equal(got, want) {
		t.Errorf("mirror content:\n%v\nwant:\n%v", got, want)
	}
}

func TestExecutePlanRefusesExistingTarget(t *testing.T) {
	target := makeTestDir(t, filepath.Join(t.TempDir(), "mirror"))
	kept := filepath.Join(target, "a.nc")
	if err := os.WriteFile(kept, []byte("data"), 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	dirs, files, err := executePlan(testPlanAt(t, target, "a.nc"))
	if !errors.Is(err, fs.ErrExist) {
		t.Fatalf("executePlan error = %v, want %v", err, fs.ErrExist)
	}
	if dirs != 0 || files != 0 {
		t.Errorf("created %d dirs, %d files, want none", dirs, files)
	}

	content, err := os.ReadFile(kept)
	if err != nil || string(content) != "data" {
		t.Errorf("existing file changed: content %q, error %v", content, err)
	}
}

func TestExecutePlanStopsAtFirstError(t *testing.T) {
	target := filepath.Join(t.TempDir(), "mirror")
	// a NUL byte is invalid in file names, so the second file cannot be created
	plan := testPlanAt(t, target, "a.nc", "b\x00.nc", "c.nc")

	dirs, files, err := executePlan(plan)
	if err == nil {
		t.Fatal("executePlan returned nil error")
	}
	if dirs != 0 || files != 1 {
		t.Errorf("created %d dirs, %d files, want 0 dirs, 1 file", dirs, files)
	}

	if _, err := os.Lstat(filepath.Join(target, "a.nc")); err != nil {
		t.Errorf("file created before the failure was removed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target, "c.nc")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("file after the failure: Lstat error = %v, want not exist", err)
	}
}
