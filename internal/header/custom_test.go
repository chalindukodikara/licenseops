// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupCustomFormat(t *testing.T, tmplContent string) *CustomFormat {
	t.Helper()
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "header.tmpl")
	if err := os.WriteFile(tmplPath, []byte(tmplContent), 0o644); err != nil {
		t.Fatal(err)
	}
	f := &CustomFormat{}
	if err := f.LoadTemplate(tmplPath); err != nil {
		t.Fatal(err)
	}
	return f
}

// --- Custom LoadTemplate ---

func TestCustom_LoadTemplate_Valid(t *testing.T) {
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}")
	if f.tmpl == nil {
		t.Error("template should be loaded")
	}
}

func TestCustom_LoadTemplate_InvalidPath(t *testing.T) {
	f := &CustomFormat{}
	err := f.LoadTemplate("/nonexistent/path/template.tmpl")
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestCustom_LoadTemplate_InvalidTemplate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.tmpl")
	os.WriteFile(p, []byte("{{.Invalid"), 0o644)
	f := &CustomFormat{}
	err := f.LoadTemplate(p)
	if err == nil {
		t.Error("expected error for invalid template syntax")
	}
}

// --- Custom Generate ---

func TestCustom_Generate_Line(t *testing.T) {
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}")
	got := f.Generate(lineSlash, "2026", "Acme Corp", "MIT")
	want := "// Copyright 2026 Acme Corp\n// License: MIT"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestCustom_Generate_Block(t *testing.T) {
	tmpl := "{{if .Comment}}{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}{{else}}{{.BlockStart}}\n Copyright {{.Year}} {{.Holder}}\n License: {{.License}}\n{{.BlockEnd}}{{end}}"
	f := setupCustomFormat(t, tmpl)
	got := f.Generate(blockC, "2026", "Acme Corp", "MIT")
	if !strings.Contains(got, "/*") {
		t.Errorf("block generate should include block start:\n%s", got)
	}
	if !strings.Contains(got, "Copyright 2026 Acme Corp") {
		t.Error("missing copyright")
	}
}

func TestCustom_Generate_NoTemplate(t *testing.T) {
	f := &CustomFormat{}
	got := f.Generate(lineSlash, "2026", "Acme Corp", "MIT")
	if got != "" {
		t.Error("Generate without template should return empty string")
	}
}

// --- Custom HasValid ---

func TestCustom_HasValid(t *testing.T) {
	dir := t.TempDir()
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}")

	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "MIT")
	content := hdr + "\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid=true")
	}
}

func TestCustom_HasValid_DifferentYear(t *testing.T) {
	dir := t.TempDir()
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}")

	content := "// Copyright 2024 Acme Corp\n// License: MIT\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("HasValid should accept any year")
	}
}

func TestCustom_HasValid_Missing(t *testing.T) {
	dir := t.TempDir()
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}")

	content := "package main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected valid=false for missing header")
	}
}

func TestCustom_HasValid_NoTemplate(t *testing.T) {
	dir := t.TempDir()
	f := &CustomFormat{}

	content := "package main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	_, err := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if err == nil {
		t.Error("expected error when template not loaded")
	}
}

// --- Custom StripExisting ---

func TestCustom_StripExisting(t *testing.T) {
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}")
	src := "// Copyright 2026 Acme Corp\n// License: MIT\n\npackage main\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if strings.Contains(got, "Copyright 2026") {
		t.Error("custom header should be stripped")
	}
	if !strings.Contains(got, "package main") {
		t.Error("code should be preserved")
	}
}

func TestCustom_StripExisting_NoTemplate(t *testing.T) {
	f := &CustomFormat{}
	src := "package main\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if got != src {
		t.Error("should return src unchanged when no template loaded")
	}
}
