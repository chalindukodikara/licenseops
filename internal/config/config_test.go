// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Load ---

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/.licenseops.yaml")
	if err != nil {
		t.Errorf("missing file should not error: %v", err)
	}
	// Should return defaults
	if cfg.Format != "spdx" {
		t.Errorf("default format = %q, want 'spdx'", cfg.Format)
	}
	if cfg.Year == "" {
		t.Error("default year should be set")
	}
	if len(cfg.Paths) != 1 || cfg.Paths[0] != "." {
		t.Errorf("default paths = %v, want ['.']", cfg.Paths)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	content := `license: MIT
copyright-holder: "Acme Corp"
format: reuse
year: "2025"
`
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte(content), 0o644)

	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.License != "MIT" {
		t.Errorf("license = %q, want 'MIT'", cfg.License)
	}
	if cfg.CopyrightHolder != "Acme Corp" {
		t.Errorf("holder = %q, want 'Acme Corp'", cfg.CopyrightHolder)
	}
	if cfg.Format != "reuse" {
		t.Errorf("format = %q, want 'reuse'", cfg.Format)
	}
	if cfg.Year != "2025" {
		t.Errorf("year = %q, want '2025'", cfg.Year)
	}
}

func TestLoad_ExcludesAppended(t *testing.T) {
	dir := t.TempDir()
	content := `license: MIT
exclude:
  - "custom/**"
`
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte(content), 0o644)

	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}

	// Should have both defaults and custom
	hasVendor := false
	hasCustom := false
	for _, e := range cfg.Exclude {
		if e == "vendor/**" {
			hasVendor = true
		}
		if e == "custom/**" {
			hasCustom = true
		}
	}
	if !hasVendor {
		t.Error("default exclude 'vendor/**' should be preserved")
	}
	if !hasCustom {
		t.Error("custom exclude should be appended")
	}
}

func TestLoad_SkipGeneratedFalse(t *testing.T) {
	dir := t.TempDir()
	content := `license: MIT
skip-generated: false
`
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte(content), 0o644)

	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ShouldSkipGenerated() {
		t.Error("skip-generated: false should be respected")
	}
}

func TestLoad_GitignoreFalse(t *testing.T) {
	dir := t.TempDir()
	content := `license: MIT
gitignore: false
`
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte(content), 0o644)

	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ShouldUseGitignore() {
		t.Error("gitignore: false should be respected")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".licenseops.yaml")
	// Use genuinely invalid YAML (tab in wrong place with map key)
	os.WriteFile(p, []byte("license: MIT\n\t\tinvalid:\n  [broken"), 0o644)

	_, err := Load(p)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// --- Validate ---

func TestValidate_MissingLicense(t *testing.T) {
	cfg := Defaults()
	_, err := cfg.Validate()
	if err == nil {
		t.Error("expected error when license is empty")
	}
}

func TestValidate_ValidSPDX(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	warnings, err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(warnings) > 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestValidate_ReuseWithoutHolder(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "reuse"
	_, err := cfg.Validate()
	if err == nil {
		t.Error("reuse format without holder should error")
	}
}

func TestValidate_ReuseWithHolder(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "reuse"
	cfg.CopyrightHolder = "Acme Corp"
	_, err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_ApacheLongWrongLicense(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "apache-long"
	cfg.CopyrightHolder = "Acme Corp"
	_, err := cfg.Validate()
	if err == nil {
		t.Error("apache-long + MIT should error")
	}
	if !strings.Contains(err.Error(), "Apache-2.0") {
		t.Errorf("error should mention Apache-2.0: %v", err)
	}
}

func TestValidate_ApacheLongValid(t *testing.T) {
	cfg := Defaults()
	cfg.License = "Apache-2.0"
	cfg.Format = "apache-long"
	cfg.CopyrightHolder = "Acme Corp"
	_, err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_GPLLongWrongLicense(t *testing.T) {
	cfg := Defaults()
	cfg.License = "Apache-2.0"
	cfg.Format = "gpl-long"
	cfg.CopyrightHolder = "Acme Corp"
	_, err := cfg.Validate()
	if err == nil {
		t.Error("gpl-long + Apache-2.0 should error")
	}
}

func TestValidate_GPLLongValid(t *testing.T) {
	cfg := Defaults()
	cfg.License = "GPL-3.0-only"
	cfg.Format = "gpl-long"
	cfg.CopyrightHolder = "Acme Corp"
	_, err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_GPLLongWithoutHolder(t *testing.T) {
	cfg := Defaults()
	cfg.License = "GPL-3.0-only"
	cfg.Format = "gpl-long"
	_, err := cfg.Validate()
	if err == nil {
		t.Error("gpl-long without holder should error")
	}
}

func TestValidate_CustomWithoutTemplate(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "custom"
	_, err := cfg.Validate()
	if err == nil {
		t.Error("custom without template should error")
	}
}

func TestValidate_CustomWithNonexistentTemplate(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "custom"
	cfg.HeaderTemplate = "/nonexistent/template.tmpl"
	_, err := cfg.Validate()
	if err == nil {
		t.Error("custom with nonexistent template should error")
	}
}

func TestValidate_CustomValid(t *testing.T) {
	dir := t.TempDir()
	tmpl := filepath.Join(dir, "header.tmpl")
	os.WriteFile(tmpl, []byte("{{.Comment}} License: {{.License}}"), 0o644)

	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "custom"
	cfg.HeaderTemplate = tmpl
	_, err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_UnknownFormat(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "nonexistent"
	_, err := cfg.Validate()
	if err == nil {
		t.Error("unknown format should error")
	}
}

func TestValidate_DeprecatedLicense(t *testing.T) {
	cfg := Defaults()
	cfg.License = "GPL-2.0"
	warnings, err := cfg.Validate()
	if err != nil {
		t.Errorf("deprecated license should not error: %v", err)
	}
	if len(warnings) == 0 {
		t.Error("expected deprecation warning")
	}
}

func TestValidate_UnknownLicense(t *testing.T) {
	cfg := Defaults()
	cfg.License = "FAKE-1.0"
	_, err := cfg.Validate()
	if err == nil {
		t.Error("unknown license should error")
	}
}

// --- Helper methods ---

func TestShouldSkipGenerated_Default(t *testing.T) {
	cfg := Defaults()
	if !cfg.ShouldSkipGenerated() {
		t.Error("default should skip generated files")
	}
}

func TestShouldSkipGenerated_Nil(t *testing.T) {
	cfg := Config{}
	if !cfg.ShouldSkipGenerated() {
		t.Error("nil SkipGenerated should default to true")
	}
}

func TestShouldUseGitignore_Default(t *testing.T) {
	cfg := Defaults()
	if !cfg.ShouldUseGitignore() {
		t.Error("default should use gitignore")
	}
}

func TestShouldUseGitignore_Nil(t *testing.T) {
	cfg := Config{}
	if !cfg.ShouldUseGitignore() {
		t.Error("nil Gitignore should default to true")
	}
}

// --- Defaults ---

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Format != "spdx" {
		t.Errorf("default format = %q", cfg.Format)
	}
	if cfg.Year == "" {
		t.Error("year should not be empty")
	}
	if len(cfg.Paths) != 1 || cfg.Paths[0] != "." {
		t.Errorf("default paths = %v", cfg.Paths)
	}
	if len(cfg.Exclude) == 0 {
		t.Error("default excludes should not be empty")
	}
}
