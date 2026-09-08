package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListFiles(t *testing.T) {
	dir := t.TempDir()
	// Create files and a directory
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "c.jpg"), []byte("img"), 0644)
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	files, err := ListFiles(dir)
	if err != nil {
		t.Fatalf("ListFiles err %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("expected 3 files got %d %v", len(files), files)
	}
	// Should be sorted
	if files[0] != "a.txt" || files[1] != "b.txt" || files[2] != "c.jpg" {
		t.Fatalf("sorting failed %v", files)
	}
}

func TestListFilesEmpty(t *testing.T) {
	dir := t.TempDir()
	files, err := ListFiles(dir)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 got %v", files)
	}
}

func TestResolveFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	os.WriteFile(path, []byte("hi"), 0644)

	abs, err := ResolveFile(path)
	if err != nil {
		t.Fatalf("ResolveFile %v", err)
	}
	if abs == "" {
		t.Fatalf("abs empty")
	}

	_, err = ResolveFile(filepath.Join(dir, "nonexistent.txt"))
	if err == nil {
		t.Fatalf("expected error for nonexistent")
	}

	// Directory should error
	subdir := filepath.Join(dir, "sub")
	os.Mkdir(subdir, 0755)
	_, err = ResolveFile(subdir)
	if err == nil {
		t.Fatalf("expected error for directory")
	}
}
