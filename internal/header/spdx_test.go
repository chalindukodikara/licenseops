// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chalindukodikara/licenseops/internal/language"
)

var (
	lineSlash = &language.Style{LinePrefix: "//"}
	lineHash  = &language.Style{LinePrefix: "#"}
	lineDash  = &language.Style{LinePrefix: "--"}
	blockC    = &language.Style{BlockStart: "/*", BlockEnd: "*/"}
)

// --- SPDX Generate ---

func TestSPDX_Generate_2Line_SlashSlash(t *testing.T) {
	f := &SPDXFormat{}
	got := f.Generate(lineSlash, "2026", "Acme Corp", "Apache-2.0")
	want := "// Copyright 2026 Acme Corp\n// SPDX-License-Identifier: Apache-2.0"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestSPDX_Generate_2Line_Hash(t *testing.T) {
	f := &SPDXFormat{}
	got := f.Generate(lineHash, "2026", "Acme Corp", "MIT")
	want := "# Copyright 2026 Acme Corp\n# SPDX-License-Identifier: MIT"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestSPDX_Generate_2Line_DashDash(t *testing.T) {
	f := &SPDXFormat{}
	got := f.Generate(lineDash, "2026", "Acme Corp", "MIT")
	want := "-- Copyright 2026 Acme Corp\n-- SPDX-License-Identifier: MIT"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestSPDX_Generate_2Line_Block(t *testing.T) {
	f := &SPDXFormat{}
	got := f.Generate(blockC, "2026", "Acme Corp", "MIT")
	if !strings.HasPrefix(got, "/*") || !strings.HasSuffix(got, "*/") {
		t.Errorf("block header should be wrapped in /* */:\n%s", got)
	}
	if !strings.Contains(got, "Copyright 2026 Acme Corp") {
		t.Error("missing copyright line")
	}
	if !strings.Contains(got, "SPDX-License-Identifier: MIT") {
		t.Error("missing SPDX line")
	}
}

func TestSPDX_Generate_1Line_SlashSlash(t *testing.T) {
	f := &SPDXFormat{}
	got := f.Generate(lineSlash, "2026", "", "MIT")
	want := "// SPDX-License-Identifier: MIT"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSPDX_Generate_1Line_Block(t *testing.T) {
	f := &SPDXFormat{}
	got := f.Generate(blockC, "2026", "", "MIT")
	want := "/* SPDX-License-Identifier: MIT */"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- SPDX HasValid ---

func TestSPDX_HasValid_2Line(t *testing.T) {
	dir := t.TempDir()
	content := "// Copyright 2026 Acme Corp\n// SPDX-License-Identifier: Apache-2.0\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &SPDXFormat{}
	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid=true")
	}
}

func TestSPDX_HasValid_2Line_DifferentYear(t *testing.T) {
	dir := t.TempDir()
	// File has 2024, config says 2026 — should still be valid
	content := "// Copyright 2024 Acme Corp\n// SPDX-License-Identifier: MIT\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &SPDXFormat{}
	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("HasValid should accept any year")
	}
}

func TestSPDX_HasValid_1Line(t *testing.T) {
	dir := t.TempDir()
	content := "// SPDX-License-Identifier: MIT\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &SPDXFormat{}
	valid, err := f.HasValid(p, lineSlash, "", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid=true for 1-line")
	}
}

func TestSPDX_HasValid_MissingHeader(t *testing.T) {
	dir := t.TempDir()
	content := "package main\n\nfunc main() {}\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &SPDXFormat{}
	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected valid=false for missing header")
	}
}

func TestSPDX_HasValid_WrongLicense(t *testing.T) {
	dir := t.TempDir()
	content := "// Copyright 2026 Acme Corp\n// SPDX-License-Identifier: MIT\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &SPDXFormat{}
	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected valid=false for wrong license")
	}
}

func TestSPDX_HasValid_WrongHolder(t *testing.T) {
	dir := t.TempDir()
	content := "// Copyright 2026 Other Corp\n// SPDX-License-Identifier: MIT\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &SPDXFormat{}
	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected valid=false for wrong holder")
	}
}

func TestSPDX_HasValid_Block(t *testing.T) {
	dir := t.TempDir()
	content := "/*\n Copyright 2026 Acme Corp\n SPDX-License-Identifier: MIT\n*/\n\n.body {}\n"
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte(content), 0o644)

	f := &SPDXFormat{}
	valid, err := f.HasValid(p, blockC, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid=true for block comment header")
	}
}

func TestSPDX_HasValid_1Line_Block(t *testing.T) {
	dir := t.TempDir()
	content := "/* SPDX-License-Identifier: MIT */\n\n.body {}\n"
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte(content), 0o644)

	f := &SPDXFormat{}
	valid, err := f.HasValid(p, blockC, "", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid=true for 1-line block header")
	}
}

// --- SPDX StripExisting ---

func TestSPDX_StripExisting_2Line(t *testing.T) {
	f := &SPDXFormat{}
	src := "// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if strings.Contains(got, "SPDX-License-Identifier") {
		t.Error("header should be stripped")
	}
	if !strings.Contains(got, "package main") {
		t.Error("code should be preserved")
	}
}

func TestSPDX_StripExisting_1Line(t *testing.T) {
	f := &SPDXFormat{}
	src := "// SPDX-License-Identifier: MIT\n\npackage main\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if strings.Contains(got, "SPDX-License-Identifier") {
		t.Error("header should be stripped")
	}
	if !strings.Contains(got, "package main") {
		t.Error("code should be preserved")
	}
}

func TestSPDX_StripExisting_NoHeader(t *testing.T) {
	f := &SPDXFormat{}
	src := "package main\n\nfunc main() {}\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if got != src {
		t.Errorf("file with no header should be unchanged\ngot:\n%s", got)
	}
}

// --- SPDX Generate+HasValid idempotency ---

func TestSPDX_Idempotency(t *testing.T) {
	dir := t.TempDir()
	f := &SPDXFormat{}

	styles := map[string]*language.Style{"slash": lineSlash, "hash": lineHash, "dash": lineDash}
	for name, style := range styles {
		hdr := f.Generate(style, "2026", "Acme Corp", "MIT")
		content := hdr + "\n\npackage main\n"
		p := filepath.Join(dir, "test_"+name+".go")
		os.WriteFile(p, []byte(content), 0o644)

		valid, err := f.HasValid(p, style, "Acme Corp", "MIT")
		if err != nil {
			t.Errorf("style %q: %v", style.LinePrefix, err)
		}
		if !valid {
			t.Errorf("style %q: Generate → HasValid should be true", style.LinePrefix)
		}
	}
}

func TestSPDX_Idempotency_Block(t *testing.T) {
	dir := t.TempDir()
	f := &SPDXFormat{}

	hdr := f.Generate(blockC, "2026", "Acme Corp", "MIT")
	content := hdr + "\n\n.body {}\n"
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte(content), 0o644)

	valid, err := f.HasValid(p, blockC, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("block Generate → HasValid should be true")
	}
}
