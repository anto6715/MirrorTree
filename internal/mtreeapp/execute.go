package mtreeapp

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirPerm  = 0o755
	filePerm = 0o644
)

// executePlan creates the mirror: the missing parents of the target, the
// target itself, every directory and an empty file for every planned file.
// It stops at the first error, leaving in place whatever was already created,
// and returns how many directories and files it created below the target.
func executePlan(plan mirrorPlan) (dirs, files int, err error) {
	if err := os.MkdirAll(filepath.Dir(plan.target), dirPerm); err != nil {
		return 0, 0, fmt.Errorf("create parents of target: %w", err)
	}

	// Mkdir, not MkdirAll: fail if the target appeared after validation
	if err := os.Mkdir(plan.target, dirPerm); err != nil {
		return 0, 0, fmt.Errorf("create target: %w", err)
	}

	c := &creator{}
	err = c.createChildren(plan.target, buildTree(plan).root)
	return c.dirs, c.files, err
}

type creator struct {
	dirs  int
	files int
}

// createChildren creates the children of node inside path, depth first in
// name order. Everything is created exclusively, so nothing is overwritten.
func (c *creator) createChildren(path string, node *treeNode) error {
	for _, child := range node.sortedChildren() {
		childPath := filepath.Join(path, child.name)

		if !child.isDir {
			if err := createEmptyFile(childPath); err != nil {
				return err
			}
			c.files++
			continue
		}

		if err := os.Mkdir(childPath, dirPerm); err != nil {
			return err
		}
		c.dirs++

		if err := c.createChildren(childPath, child); err != nil {
			return err
		}
	}

	return nil
}

func createEmptyFile(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePerm)
	if err != nil {
		return err
	}
	return file.Close()
}
