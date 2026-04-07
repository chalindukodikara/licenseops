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
	if got := DetectProjectType(t.TempDir()); got != ProjectGeneric {
		t.Errorf("got %q, want %q", got, ProjectGeneric)
	}
}

func TestSuggestExcludes(t *testing.T) {
	if len(SuggestExcludes(ProjectGo)) == 0 {
		t.Error("Go should have excludes")
	}
	if len(SuggestExcludes(ProjectGeneric)) != 0 {
		t.Error("Generic should have no extra excludes")
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
	if strings.Contains(got, "format:") {
		t.Error("spdx format should not be listed")
	}
}

func TestGenerateConfig_NonDefaultFormat(t *testing.T) {
	got := GenerateConfig("MIT", "Acme", "reuse", nil)
	if !strings.Contains(got, "format: reuse") {
		t.Error("non-default format should be listed")
	}
}

func TestGenerateConfig_WithExcludes(t *testing.T) {
	got := GenerateConfig("MIT", "", "spdx", []string{"dist/**"})
	if !strings.Contains(got, "exclude:") || !strings.Contains(got, "dist/**") {
		t.Error("should contain excludes")
	}
}

func TestGenerateConfig_NoHolder(t *testing.T) {
	got := GenerateConfig("MIT", "", "spdx", nil)
	if strings.Contains(got, "copyright-holder") {
		t.Error("should not contain holder when empty")
	}
}
