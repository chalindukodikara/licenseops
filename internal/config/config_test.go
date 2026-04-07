// Copyright 2026 The LicenseOps Authors
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

func TestLoad_AllFields(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".licenseops.yaml")
	yaml := `license: Apache-2.0
copyright-holder: "Acme Corp"
year: "2025"
format: reuse
header-template: /tmp/foo.tmpl
paths:
  - src/
  - lib/
exclude:
  - "build/**"
skip-generated: false
gitignore: false
`
	os.WriteFile(p, []byte(yaml), 0o644)
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.License != "Apache-2.0" {
		t.Errorf("license = %q", cfg.License)
	}
	if cfg.HeaderTemplate != "/tmp/foo.tmpl" {
		t.Errorf("template = %q", cfg.HeaderTemplate)
	}
	if len(cfg.Paths) != 2 {
		t.Errorf("paths = %v", cfg.Paths)
	}
	if cfg.ShouldSkipGenerated() {
		t.Error("skip-generated false should be respected")
	}
	if cfg.ShouldUseGitignore() {
		t.Error("gitignore false should be respected")
	}
}

func TestValidate_ApacheLongValid(t *testing.T) {
	cfg := Defaults()
	cfg.License = "Apache-2.0"
	cfg.Format = "apache-long"
	cfg.CopyrightHolder = "Acme"
	if _, err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_ApacheLongMissingHolder(t *testing.T) {
	cfg := Defaults()
	cfg.License = "Apache-2.0"
	cfg.Format = "apache-long"
	if _, err := cfg.Validate(); err == nil {
		t.Error("apache-long without holder should error")
	}
}

func TestValidate_GPLLongValid(t *testing.T) {
	for _, lic := range []string{"GPL-3.0-only", "LGPL-2.1-only", "AGPL-3.0-or-later"} {
		t.Run(lic, func(t *testing.T) {
			cfg := Defaults()
			cfg.License = lic
			cfg.Format = "gpl-long"
			cfg.CopyrightHolder = "Acme"
			if _, err := cfg.Validate(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidate_GPLLongMissingHolder(t *testing.T) {
	cfg := Defaults()
	cfg.License = "GPL-3.0-only"
	cfg.Format = "gpl-long"
	if _, err := cfg.Validate(); err == nil {
		t.Error("gpl-long without holder should error")
	}
}

func TestValidate_CustomMissingFile(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "custom"
	cfg.HeaderTemplate = "/nonexistent/template.tmpl"
	if _, err := cfg.Validate(); err == nil {
		t.Error("custom with missing template file should error")
	}
}

func TestValidate_ReuseValid(t *testing.T) {
	cfg := Defaults()
	cfg.License = "MIT"
	cfg.Format = "reuse"
	cfg.CopyrightHolder = "Acme"
	if _, err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_ExpressionLicense(t *testing.T) {
	cfg := Defaults()
	cfg.License = "Apache-2.0 OR MIT"
	if _, err := cfg.Validate(); err != nil {
		t.Errorf("expression should be valid: %v", err)
	}
}

func TestValidate_InvalidLicenseID(t *testing.T) {
	cfg := Defaults()
	cfg.License = "INVALID-1.0"
	if _, err := cfg.Validate(); err == nil {
		t.Error("invalid license should error")
	}
}

func TestShouldSkipGenerated_FalseValue(t *testing.T) {
	f := false
	cfg := Config{SkipGenerated: &f}
	if cfg.ShouldSkipGenerated() {
		t.Error("explicit false should be respected")
	}
}

func TestShouldUseGitignore_FalseValue(t *testing.T) {
	f := false
	cfg := Config{Gitignore: &f}
	if cfg.ShouldUseGitignore() {
		t.Error("explicit false should be respected")
	}
}

func TestLoad_PathsReplacedNotAppended(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".licenseops.yaml")
	os.WriteFile(p, []byte("license: MIT\npaths:\n  - src/\n  - test/\n"), 0o644)
	cfg, _ := Load(p)
	if len(cfg.Paths) != 2 || cfg.Paths[0] != "src/" || cfg.Paths[1] != "test/" {
		t.Errorf("paths should be replaced, got %v", cfg.Paths)
	}
}
