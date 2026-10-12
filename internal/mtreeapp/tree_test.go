package mtreeapp

import "testing"

func TestBuildTreeCreatesIntermediateDirs(t *testing.T) {
	tree := buildTree(testPlan(t, "a/b/c.nc"))

	if tree.root.name != "mirror" || !tree.root.isDir {
		t.Fatalf("root = %+v, want directory named %q", tree.root, "mirror")
	}

	a := tree.root.children["a"]
	if a == nil || !a.isDir {
		t.Fatalf("missing intermediate directory a: %+v", tree.root.children)
	}
	b := a.children["b"]
	if b == nil || !b.isDir {
		t.Fatalf("missing intermediate directory a/b: %+v", a.children)
	}
	file := b.children["c.nc"]
	if file == nil || file.isDir {
		t.Fatalf("missing file a/b/c.nc: %+v", b.children)
	}

	if tree.dirs != 2 || tree.files != 1 {
		t.Errorf("counts = %d dirs, %d files, want 2 dirs, 1 file", tree.dirs, tree.files)
	}
}

func TestBuildTreeCounts(t *testing.T) {
	tests := []struct {
		name      string
		rels      []string
		wantDirs  int
		wantFiles int
	}{
		{name: "empty plan", rels: nil, wantDirs: 0, wantFiles: 0},
		{name: "files in root only", rels: []string{"a.nc", "b.nc"}, wantDirs: 0, wantFiles: 2},
		{
			name:      "shared parents counted once",
			rels:      []string{"a.nc", "database/x.nc", "sub/b.txt", "sub/deep/c.nc"},
			wantDirs:  3,
			wantFiles: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(testPlan(t, tt.rels...))
			if tree.dirs != tt.wantDirs || tree.files != tt.wantFiles {
				t.Errorf("counts = %d dirs, %d files, want %d dirs, %d files",
					tree.dirs, tree.files, tt.wantDirs, tt.wantFiles)
			}
		})
	}
}
