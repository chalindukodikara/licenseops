// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/.licenseops.yaml")
	if err != nil {
		t.Errorf("missing file should not error: %v", err)
	}
	if cfg.Format != "spdx" {
		t.Errorf("default format = %q", cfg.Format)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte("license: MIT\ncopyright-holder: \"Acme Corp\"\nformat: reuse\nyear: \"2025\"\n"), 0o644)
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.License != "MIT" || cfg.CopyrightHolder != "Acme Corp" || cfg.Format != "reuse" || cfg.Year != "2025" {
		t.Errorf("unexpected cfg: %+v", cfg)
	}
}

func TestLoad_ExcludesAppended(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte("license: MIT\nexclude:\n  - \"custom/**\"\n"), 0o644)
	cfg, _ := Load(p)
	hasVendor, hasCustom := false, false
	for _, e := range cfg.Exclude {
		if e == "vendor/**" {
			hasVendor = true
		}
		if e == "custom/**" {
			hasCustom = true
		}
	}
	if !hasVendor {
		t.Error("default 'vendor/**' should be preserved")
	}
	if !hasCustom {
		t.Error("custom exclude should be appended")
	}
}

func TestLoad_SkipGeneratedFalse(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte("license: MIT\nskip-generated: false\n"), 0o644)
	cfg, _ := Load(p)
	if cfg.ShouldSkipGenerated() {
		t.Error("skip-generated: false should be respected")
	}
}

func TestLoad_GitignoreFalse(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte("license: MIT\ngitignore: false\n"), 0o644)
	cfg, _ := Load(p)
	if cfg.ShouldUseGitignore() {
		t.Error("gitignore: false should be respected")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte("license: MIT\n\t\tinvalid:\n  [broken"), 0o644)
	if _, err := Load(p); err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestValidate_MissingLicense(t *testing.T) {
	cfg := Defaults()
	if _, err := cfg.Validate(); err == nil {
		t.Error("expected error when license is empty")
	}
}

func TestValidate_ValidSPDX(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	if _, err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_ReuseWithoutHolder(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "reuse"
	if _, err := cfg.Validate(); err == nil {
		t.Error("reuse without holder should error")
	}
}

func TestValidate_ApacheLongWrongLicense(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "apache-long"
	cfg.CopyrightHolder = "Acme"
	_, err := cfg.Validate()
	if err == nil {
		t.Error("apache-long + MIT should error")
	}
	if !strings.Contains(err.Error(), "Apache-2.0") {
		t.Errorf("error should mention Apache-2.0: %v", err)
	}
}

func TestValidate_GPLLongWrongLicense(t *testing.T) {
	cfg := Defaults()
	cfg.License = "Apache-2.0"
	cfg.Format = "gpl-long"
	cfg.CopyrightHolder = "Acme"
	if _, err := cfg.Validate(); err == nil {
		t.Error("gpl-long + Apache should error")
	}
}

func TestValidate_CustomWithoutTemplate(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "custom"
	if _, err := cfg.Validate(); err == nil {
		t.Error("custom without template should error")
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
	if _, err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_UnknownFormat(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "nonexistent"
	if _, err := cfg.Validate(); err == nil {
		t.Error("unknown format should error")
	}
}

func TestValidate_DeprecatedLicense(t *testing.T) {
	cfg := Defaults()
	cfg.License = "GPL-2.0"
	warnings, err := cfg.Validate()
	if err != nil {
		t.Errorf("deprecated should not error: %v", err)
	}
	if len(warnings) == 0 {
		t.Error("expected deprecation warning")
	}
}

func TestShouldSkipGenerated_Default(t *testing.T) {
	cfg := Defaults()
	if !cfg.ShouldSkipGenerated() {
		t.Error("default should skip generated")
	}
}

func TestShouldSkipGenerated_Nil(t *testing.T) {
	cfg := Config{}
	if !cfg.ShouldSkipGenerated() {
		t.Error("nil should default to true")
	}
}

func TestShouldUseGitignore_Default(t *testing.T) {
	cfg := Defaults()
	if !cfg.ShouldUseGitignore() {
		t.Error("default should use gitignore")
	}
}

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Format != "spdx" || cfg.Year == "" || len(cfg.Paths) != 1 || len(cfg.Exclude) == 0 {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}
