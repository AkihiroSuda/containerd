package testutil

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FileInfo is a handy struct for asserting information related to file.
// TODO: use continuity? ( requires https://github.com/stevvooe/continuity/issues/48 )
type FileInfo struct {
	BaseName string       // For root, this MUST be empty.
	Mode     *os.FileMode // ModeType | ModePerm. For root, this SHOULD be nil. (especially for overlay)
	Content  []byte
	Target   *string     // for symlink
	Children []*FileInfo // for dir
	// TODO: (optional) timestamp and so on
}

func regular(mode os.FileMode) bool {
	irregular := (os.ModeDir | os.ModeSymlink | os.ModeDevice | os.ModeNamedPipe | os.ModeSocket)
	return mode&irregular == 0
}

func _assertTree(path string, tree *FileInfo) error {
	if tree == nil {
		return nil
	}
	if tree.BaseName != "" {
		if tree.BaseName != filepath.Base(path) {
			return fmt.Errorf("expected basename %q, got %q for %q(%+v)",
				tree.BaseName, filepath.Base(path), path, tree)
		}
	}
	stat, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if tree.Mode != nil && *tree.Mode != stat.Mode() {
		return fmt.Errorf("expected mode %v, got %v for %q (%+v)",
			tree.Mode, stat.Mode(), path, tree)
	}
	if regular(stat.Mode()) {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Compare(tree.Content, content) != 0 {
			return fmt.Errorf("expected content %v, got %v for %q (%+v)",
				tree.Content, content, path, tree)
		}
	}
	if stat.Mode()&os.ModeSymlink != 0 && tree.Target != nil {
		target, err := os.Readlink(path)
		if err != nil {
			return err
		}
		if target != *tree.Target {
			return fmt.Errorf("expected target %q, got %q for %q (%+v)",
				tree.Target, target, path, tree)
		}
	}

	if stat.IsDir() {
		children, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		for _, child := range children {
			name := child.Name()
			found := false
			for _, exChild := range tree.Children {
				if exChild.BaseName == name {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("extra file %q for %q (%+v)",
					name, path, tree)
			}
		}
	}

	for _, child := range tree.Children {
		if child.BaseName == "" ||
			strings.Contains(child.BaseName, "/") {
			return fmt.Errorf("wrong basename: %s", child.BaseName)
		}
		if err := _assertTree(
			filepath.Join(path, child.BaseName),
			child); err != nil {
			return err
		}
	}
	return nil
}

func assertTree(dir string, tree *FileInfo) error {
	if tree == nil || tree.BaseName != "" ||
		(tree.Mode != nil && *tree.Mode&os.ModeDir == 0) {
		return errors.New("wrong tree specified")
	}
	return _assertTree(dir, tree)
}

// AssertTree asserts that the structure of dir matches the expected tree exactly.
// // i.e. existing file which is not specified in tree will result in an error
func AssertTree(t *testing.T, dir string, tree *FileInfo) {
	if err := assertTree(dir, tree); err != nil {
		t.Fatal(err)
	}
}
