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
	os.WriteFile(tmplPath, []byte(tmplContent), 0o644)
	f := &CustomFormat{}
	if err := f.LoadTemplate(tmplPath); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestCustom_LoadTemplate_Valid(t *testing.T) {
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}")
	if f.tmpl == nil {
		t.Error("template should be loaded")
	}
}

func TestCustom_LoadTemplate_InvalidPath(t *testing.T) {
	if err := (&CustomFormat{}).LoadTemplate("/nonexistent/template.tmpl"); err == nil {
		t.Error("expected error")
	}
}

func TestCustom_LoadTemplate_InvalidTemplate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.tmpl")
	os.WriteFile(p, []byte("{{.Invalid"), 0o644)
	if err := (&CustomFormat{}).LoadTemplate(p); err == nil {
		t.Error("expected error for invalid syntax")
	}
}

func TestCustom_Generate_Line(t *testing.T) {
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}")
	got := f.Generate(lineSlash, "2026", "Acme Corp", "MIT")
	if got != "// Copyright 2026 Acme Corp\n// License: MIT" {
		t.Errorf("got:\n%s", got)
	}
}

func TestCustom_Generate_NoTemplate(t *testing.T) {
	if got := (&CustomFormat{}).Generate(lineSlash, "2026", "Acme", "MIT"); got != "" {
		t.Error("should return empty without template")
	}
}

func TestCustom_HasValid(t *testing.T) {
	dir := t.TempDir()
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}")
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "MIT")
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(hdr+"\n\npackage main\n"), 0o644)
	valid, _ := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if !valid {
		t.Error("expected valid")
	}
}

func TestCustom_HasValid_DifferentYear(t *testing.T) {
	dir := t.TempDir()
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}")
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// Copyright 2024 Acme Corp\n// License: MIT\n\npackage main\n"), 0o644)
	valid, _ := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if !valid {
		t.Error("should accept any year")
	}
}

func TestCustom_HasValid_NoTemplate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("package main\n"), 0o644)
	if _, err := (&CustomFormat{}).HasValid(p, lineSlash, "Acme", "MIT"); err == nil {
		t.Error("expected error without template")
	}
}

func TestCustom_StripExisting(t *testing.T) {
	f := setupCustomFormat(t, "{{.Comment}} Copyright {{.Year}} {{.Holder}}\n{{.Comment}} License: {{.License}}")
	got := string(f.StripExisting([]byte("// Copyright 2026 Acme Corp\n// License: MIT\n\npackage main\n"), lineSlash))
	if strings.Contains(got, "Copyright 2026") {
		t.Error("should be stripped")
	}
}

func TestCustom_StripExisting_NoTemplate(t *testing.T) {
	src := "package main\n"
	if got := string((&CustomFormat{}).StripExisting([]byte(src), lineSlash)); got != src {
		t.Error("should be unchanged without template")
	}
}
