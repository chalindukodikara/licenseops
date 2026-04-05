// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper to create a file in a temp dir.
func createFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScan_FindsSupportedFiles(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "script.py", "print('hi')")
	createFile(t, dir, "app.js", "console.log('hi')")

	s := New(dir, nil, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(files), files)
	}
}

func TestScan_SkipsUnsupported(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "data.json", `{"key": "value"}`)
	createFile(t, dir, "readme.txt", "hello")

	s := New(dir, nil, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	// Only main.go should be found (.json is skipped, .txt is unsupported)
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(files), files)
	}
}

func TestScan_SkipsBinaryExtensions(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "image.png", "fake png")
	createFile(t, dir, "font.woff2", "fake font")
	createFile(t, dir, "app.exe", "fake exe")

	s := New(dir, nil, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file (main.go), got %d: %v", len(files), files)
	}
}

func TestScan_ExcludePatterns(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "vendor/lib/lib.go", "package lib")
	createFile(t, dir, "node_modules/pkg/index.js", "module.exports = {}")

	s := New(dir, []string{"vendor/**", "node_modules/**"}, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(files), files)
	}
}

func TestScan_ExcludeSimpleGlob(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "gen/foo.pb.go", "package gen")

	s := New(dir, []string{"*.pb.go"}, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if filepath.Base(f) == "foo.pb.go" {
			t.Error("*.pb.go should be excluded")
		}
	}
}

func TestScan_NestedDirectories(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "a/b/c/deep.go", "package deep")
	createFile(t, dir, "a/top.go", "package a")

	s := New(dir, nil, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestScan_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	s := New(dir, nil, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestScan_SingleFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")

	s := New(dir, nil, false)
	files, err := s.Scan([]string{filepath.Join(dir, "main.go")})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestScan_DirectoryExclude_SkipsVendor(t *testing.T) {
	// Exclude patterns work when scanning directories (not single files).
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "vendor/lib.go", "package vendor")

	s := New(dir, []string{"vendor/**"}, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.Contains(f, "vendor") {
			t.Error("vendor directory should be excluded during directory scan")
		}
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(files), files)
	}
}

func TestScan_RespectsGitignore(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".gitignore", "build/\n")
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "build/output.go", "package build")

	s := New(dir, nil, true)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if filepath.Base(f) == "output.go" {
			t.Error("build/ files should be ignored via .gitignore")
		}
	}
}

func TestScan_GitignoreDisabled(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".gitignore", "*.go\n")
	createFile(t, dir, "main.go", "package main")

	s := New(dir, nil, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("with gitignore disabled, expected 1 file, got %d", len(files))
	}
}

func TestScan_FilenameStyles(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "Dockerfile", "FROM golang")
	createFile(t, dir, "Makefile", ".PHONY: all")

	s := New(dir, nil, false)
	files, err := s.Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files (Dockerfile, Makefile), got %d: %v", len(files), files)
	}
}
