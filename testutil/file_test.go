package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func prepareTree(t *testing.T) (string, *FileInfo) {
	fooMode := os.ModeSymlink | 0777
	fooTarget := "../bar/baz"
	barMode := os.ModeDir | 0700
	barBazMode := os.FileMode(0600)
	tree := &FileInfo{
		BaseName: "",
		Children: []*FileInfo{
			{
				BaseName: "foo",
				Mode:     &fooMode,
				Target:   &fooTarget,
			},
			{
				BaseName: "bar",
				Mode:     &barMode,
				Children: []*FileInfo{
					{
						BaseName: "baz",
						Mode:     &barBazMode,
						Content:  []byte("1\n"),
					},
				},
			},
		},
	}
	tempDir, err := ioutil.TempDir("", "testutil-tree")
	if err != nil {
		t.Fatal(err)
	}
	foo := filepath.Join(tempDir, "foo")
	bar := filepath.Join(tempDir, "bar")
	barBaz := filepath.Join(bar, "baz")
	if err := os.Mkdir(bar, 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(barBaz, []byte("1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	os.Chdir(tempDir)
	if err := os.Symlink("../bar/baz", foo); err != nil {
		t.Fatal(err)
	}
	os.Chdir(wd)
	return tempDir, tree
}

func TestAssertTree(t *testing.T) {
	d, tree := prepareTree(t)
	DumpDir(t, d)
	defer os.RemoveAll(d)
	AssertTree(t, d, tree)
}

func TestAssertTreeExtraFile(t *testing.T) {
	d, tree := prepareTree(t)
	if err := ioutil.WriteFile(filepath.Join(d, "extra"),
		[]byte("extra\n"), 0777); err != nil {
		t.Fatal(err)
	}
	DumpDir(t, d)
	defer os.RemoveAll(d)
	err := assertTree(d, tree)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "extra") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestAssertTreeMissingFile(t *testing.T) {
	d, tree := prepareTree(t)
	if err := os.Remove(filepath.Join(d, "foo")); err != nil {
		t.Fatal(err)
	}
	DumpDir(t, d)
	defer os.RemoveAll(d)
	err := assertTree(d, tree)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no such") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestAssertTreeWrongMode(t *testing.T) {
	d, tree := prepareTree(t)
	if err := os.Chmod(filepath.Join(d, "bar", "baz"), 0400); err != nil {
		t.Fatal(err)
	}
	DumpDir(t, d)
	defer os.RemoveAll(d)
	err := assertTree(d, tree)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "expected mode") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestAssertTreeWrongContent(t *testing.T) {
	d, tree := prepareTree(t)
	if err := ioutil.WriteFile(filepath.Join(d, "bar", "baz"),
		[]byte("2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	DumpDir(t, d)
	defer os.RemoveAll(d)
	err := assertTree(d, tree)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "expected content") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestAssertTreeWrongTarget(t *testing.T) {
	d, tree := prepareTree(t)
	foo := filepath.Join(d, "foo")
	if err := os.Remove(foo); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("nowhere", foo); err != nil {
		t.Fatal(err)
	}
	DumpDir(t, d)
	defer os.RemoveAll(d)
	err := assertTree(d, tree)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "expected target") {
		t.Fatalf("unexpected error %v", err)
	}
}
