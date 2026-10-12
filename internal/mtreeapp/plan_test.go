package mtreeapp

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestMirrorPath(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		path    string
		wantRel string
		want    string
	}{
		{name: "source root", target: "/mirror", path: "/data", wantRel: ".", want: "/mirror"},
		{name: "nested file", target: "/mirror", path: "/data/sub/a.nc", wantRel: "sub/a.nc", want: "/mirror/sub/a.nc"},
		{name: "child sharing source prefix", target: "/mirror", path: "/data/database", wantRel: "database", want: "/mirror/database"},
		{name: "source name repeated in path", target: "/mirror", path: "/data/2024/data", wantRel: "2024/data", want: "/mirror/2024/data"},
		{name: "relative target", target: "mirror", path: "/data/sub", wantRel: "sub", want: "mirror/sub"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, got, err := mirrorPath("/data", tt.target, tt.path)
			if err != nil {
				t.Fatalf("mirrorPath returned an error: %v", err)
			}
			if rel != tt.wantRel || got != tt.want {
				t.Errorf("mirrorPath(%q, %q, %q) = (%q, %q), want (%q, %q)",
					"/data", tt.target, tt.path, rel, got, tt.wantRel, tt.want)
			}
		})
	}
}

func TestBuildMirrorPlanSortsDirsAndFiles(t *testing.T) {
	cfg := Config{Source: "/data", Target: "/mirror", Filter: "*"}
	sourceFiles := map[string][]string{
		"/data/sub": {"/data/sub/b.txt", "/data/sub/a.txt"},
		"/data":     {"/data/z.nc"},
	}

	got, err := buildMirrorPlan(cfg, sourceFiles)
	if err != nil {
		t.Fatalf("buildMirrorPlan returned an error: %v", err)
	}

	want := mirrorPlan{
		source: "/data",
		target: "/mirror",
		filter: "*",
		dirs: []mirrorDir{
			{
				source: "/data",
				target: "/mirror",
				files:  []mirrorFile{{source: "/data/z.nc", target: "/mirror/z.nc", rel: "z.nc"}},
			},
			{
				source: "/data/sub",
				target: "/mirror/sub",
				files: []mirrorFile{
					{source: "/data/sub/a.txt", target: "/mirror/sub/a.txt", rel: "sub/a.txt"},
					{source: "/data/sub/b.txt", target: "/mirror/sub/b.txt", rel: "sub/b.txt"},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("plan:\n%+v\nwant:\n%+v", got, want)
	}

	if sourceFiles["/data/sub"][0] != "/data/sub/b.txt" {
		t.Errorf("buildMirrorPlan reordered input files: %v", sourceFiles["/data/sub"])
	}
}

func TestBuildMirrorPlanEmpty(t *testing.T) {
	got, err := buildMirrorPlan(Config{Source: "/data", Target: "/mirror"}, nil)
	if err != nil {
		t.Fatalf("buildMirrorPlan returned an error: %v", err)
	}
	if len(got.dirs) != 0 {
		t.Errorf("dirs = %+v, want none", got.dirs)
	}
}

// testPlan builds a plan mirroring /data into mirror from source-relative file
// paths, grouping them by directory as find.FindFilesWithDirs does.
func testPlan(t *testing.T, rels ...string) mirrorPlan {
	t.Helper()

	sourceFiles := map[string][]string{}
	for _, rel := range rels {
		file := filepath.Join("/data", rel)
		dir := filepath.Dir(file)
		sourceFiles[dir] = append(sourceFiles[dir], file)
	}

	plan, err := buildMirrorPlan(Config{Source: "/data", Target: "mirror", Filter: "*"}, sourceFiles)
	if err != nil {
		t.Fatalf("buildMirrorPlan returned an error: %v", err)
	}
	return plan
}
