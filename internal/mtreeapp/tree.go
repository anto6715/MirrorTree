package mtreeapp

import (
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

type treeNode struct {
	name     string
	isDir    bool
	children map[string]*treeNode
}

type mirrorTree struct {
	root  *treeNode
	dirs  int // excludes the root, like the tree command
	files int
}

// buildTree arranges the plan files into a tree rooted at the target. The plan
// only lists directories that hold files, so intermediate directories are
// created here from each file's relative path.
func buildTree(plan mirrorPlan) mirrorTree {
	tree := mirrorTree{root: newDirNode(plan.target)}

	for _, dir := range plan.dirs {
		for _, file := range dir.files {
			parts := strings.Split(file.rel, string(filepath.Separator))
			name := parts[len(parts)-1]

			node := tree.root
			for _, part := range parts[:len(parts)-1] {
				child, ok := node.children[part]
				if !ok {
					child = newDirNode(part)
					node.children[part] = child
					tree.dirs++
				}
				node = child
			}

			node.children[name] = &treeNode{name: name}
			tree.files++
		}
	}

	return tree
}

func newDirNode(name string) *treeNode {
	return &treeNode{name: name, isDir: true, children: map[string]*treeNode{}}
}

// sortedChildren returns the children of n sorted by name.
func (n *treeNode) sortedChildren() []*treeNode {
	children := make([]*treeNode, 0, len(n.children))
	for _, name := range slices.Sorted(maps.Keys(n.children)) {
		children = append(children, n.children[name])
	}
	return children
}
