// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	files, err := New(dir, nil, false).Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3, got %d: %v", len(files), files)
	}
}

func TestScan_SkipsUnsupported(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "data.json", `{}`)
	createFile(t, dir, "readme.txt", "hello")
	files, _ := New(dir, nil, false).Scan([]string{dir})
	if len(files) != 1 {
		t.Errorf("expected 1, got %d", len(files))
	}
}

func TestScan_SkipsBinaryExtensions(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "image.png", "fake")
	createFile(t, dir, "app.exe", "fake")
	files, _ := New(dir, nil, false).Scan([]string{dir})
	if len(files) != 1 {
		t.Errorf("expected 1, got %d", len(files))
	}
}

func TestScan_ExcludePatterns(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "vendor/lib/lib.go", "package lib")
	createFile(t, dir, "node_modules/pkg/index.js", "module.exports = {}")
	files, _ := New(dir, []string{"vendor/**", "node_modules/**"}, false).Scan([]string{dir})
	if len(files) != 1 {
		t.Errorf("expected 1, got %d", len(files))
	}
}

func TestScan_NestedDirectories(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "a/b/c/deep.go", "package deep")
	createFile(t, dir, "a/top.go", "package a")
	files, _ := New(dir, nil, false).Scan([]string{dir})
	if len(files) != 2 {
		t.Errorf("expected 2, got %d", len(files))
	}
}

func TestScan_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	files, _ := New(dir, nil, false).Scan([]string{dir})
	if len(files) != 0 {
		t.Errorf("expected 0, got %d", len(files))
	}
}

func TestScan_SingleFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	files, _ := New(dir, nil, false).Scan([]string{filepath.Join(dir, "main.go")})
	if len(files) != 1 {
		t.Errorf("expected 1, got %d", len(files))
	}
}

func TestScan_RespectsGitignore(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".gitignore", "build/\n")
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "build/output.go", "package build")
	files, _ := New(dir, nil, true).Scan([]string{dir})
	for _, f := range files {
		if filepath.Base(f) == "output.go" {
			t.Error("build/ should be gitignored")
		}
	}
}

func TestScan_GitignoreDisabled(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".gitignore", "*.go\n")
	createFile(t, dir, "main.go", "package main")
	files, _ := New(dir, nil, false).Scan([]string{dir})
	if len(files) != 1 {
		t.Errorf("with gitignore disabled, expected 1, got %d", len(files))
	}
}

func TestScan_FilenameStyles(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "Dockerfile", "FROM golang")
	createFile(t, dir, "Makefile", ".PHONY: all")
	files, _ := New(dir, nil, false).Scan([]string{dir})
	if len(files) != 2 {
		t.Errorf("expected 2, got %d", len(files))
	}
}

func TestScan_DirectoryExclude_SkipsVendor(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "vendor/lib.go", "package vendor")
	files, _ := New(dir, []string{"vendor/**"}, false).Scan([]string{dir})
	for _, f := range files {
		if strings.Contains(f, "vendor") {
			t.Error("vendor should be excluded")
		}
	}
	if len(files) != 1 {
		t.Errorf("expected 1, got %d", len(files))
	}
}

func TestScan_NonExistentPath(t *testing.T) {
	_, err := New("/nonexistent", nil, false).Scan([]string{"/totally/missing/path"})
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestScan_SingleFile_Excluded(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	files, _ := New(dir, []string{"main.go"}, false).Scan([]string{filepath.Join(dir, "main.go")})
	if len(files) != 0 {
		t.Errorf("excluded single file should not be returned, got %d", len(files))
	}
}

func TestScan_SingleFile_Unsupported(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "data.json", "{}")
	files, _ := New(dir, nil, false).Scan([]string{filepath.Join(dir, "data.json")})
	if len(files) != 0 {
		t.Errorf("unsupported single file should not be returned, got %d", len(files))
	}
}

func TestScan_FilenamePatternMatch(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "a.go", "package a")
	createFile(t, dir, "b.go", "package b")
	createFile(t, dir, "nested/a.go", "package a")
	// "a.go" pattern without slash should match by basename anywhere
	files, _ := New(dir, []string{"a.go"}, false).Scan([]string{dir})
	for _, f := range files {
		if filepath.Base(f) == "a.go" {
			t.Errorf("a.go should be excluded by basename pattern, got %s", f)
		}
	}
}

func TestScan_DoublestarPattern(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "main.go", "package main")
	createFile(t, dir, "src/util.go", "package util")
	createFile(t, dir, "src/sub/deep.go", "package sub")
	files, _ := New(dir, []string{"src/**"}, false).Scan([]string{dir})
	if len(files) != 1 {
		t.Errorf("expected 1 (only main.go), got %d: %v", len(files), files)
	}
}
