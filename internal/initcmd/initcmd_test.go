// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package initcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectProjectType_Go(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/foo"), 0o644)
	if got := DetectProjectType(dir); got != ProjectGo {
		t.Errorf("got %q, want %q", got, ProjectGo)
	}
}

func TestDetectProjectType_Node(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644)
	if got := DetectProjectType(dir); got != ProjectNode {
		t.Errorf("got %q, want %q", got, ProjectNode)
	}
}

func TestDetectProjectType_Python(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(""), 0o644)
	if got := DetectProjectType(dir); got != ProjectPython {
		t.Errorf("got %q, want %q", got, ProjectPython)
	}
}

func TestDetectProjectType_Rust(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(""), 0o644)
	if got := DetectProjectType(dir); got != ProjectRust {
		t.Errorf("got %q, want %q", got, ProjectRust)
	}
}

func TestDetectProjectType_Generic(t *testing.T) {
	dir := t.TempDir()
	if got := DetectProjectType(dir); got != ProjectGeneric {
		t.Errorf("got %q, want %q", got, ProjectGeneric)
	}
}

func TestSuggestExcludes(t *testing.T) {
	if got := SuggestExcludes(ProjectGo); len(got) == 0 {
		t.Error("Go project should have suggested excludes")
	}
	if got := SuggestExcludes(ProjectNode); len(got) == 0 {
		t.Error("Node project should have suggested excludes")
	}
	if got := SuggestExcludes(ProjectPython); len(got) == 0 {
		t.Error("Python project should have suggested excludes")
	}
	if got := SuggestExcludes(ProjectRust); len(got) == 0 {
		t.Error("Rust project should have suggested excludes")
	}
	if got := SuggestExcludes(ProjectGeneric); len(got) != 0 {
		t.Errorf("Generic project should have no extra excludes, got %v", got)
	}
}

func TestGenerateConfig_Basic(t *testing.T) {
	got := GenerateConfig("MIT", "Acme Corp", "spdx", nil)
	if !strings.Contains(got, "license: MIT") {
		t.Error("should contain license")
	}
	if !strings.Contains(got, `copyright-holder: "Acme Corp"`) {
		t.Error("should contain holder")
	}
	// spdx is default, should not be explicitly listed
	if strings.Contains(got, "format:") {
		t.Error("spdx format should not be listed (it's the default)")
	}
}

func TestGenerateConfig_NonDefaultFormat(t *testing.T) {
	got := GenerateConfig("MIT", "Acme Corp", "reuse", nil)
	if !strings.Contains(got, "format: reuse") {
		t.Error("non-default format should be listed")
	}
}

func TestGenerateConfig_WithExcludes(t *testing.T) {
	got := GenerateConfig("MIT", "", "spdx", []string{"dist/**", "build/**"})
	if !strings.Contains(got, "exclude:") {
		t.Error("should contain exclude section")
	}
	if !strings.Contains(got, "dist/**") {
		t.Error("should contain exclude pattern")
	}
}

func TestGenerateConfig_NoHolder(t *testing.T) {
	got := GenerateConfig("MIT", "", "spdx", nil)
	if strings.Contains(got, "copyright-holder") {
		t.Error("should not contain holder when empty")
	}
}
