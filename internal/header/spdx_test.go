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
		t.Errorf("block header wrong:\n%s", got)
	}
}

func TestSPDX_Generate_1Line(t *testing.T) {
	f := &SPDXFormat{}
	if got := f.Generate(lineSlash, "2026", "", "MIT"); got != "// SPDX-License-Identifier: MIT" {
		t.Errorf("got %q", got)
	}
}

func TestSPDX_Generate_1Line_Block(t *testing.T) {
	f := &SPDXFormat{}
	if got := f.Generate(blockC, "2026", "", "MIT"); got != "/* SPDX-License-Identifier: MIT */" {
		t.Errorf("got %q", got)
	}
}

func TestSPDX_HasValid_2Line(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// Copyright 2026 Acme Corp\n// SPDX-License-Identifier: Apache-2.0\n\npackage main\n"), 0o644)
	valid, err := (&SPDXFormat{}).HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid")
	}
}

func TestSPDX_HasValid_DifferentYear(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// Copyright 2024 Acme Corp\n// SPDX-License-Identifier: MIT\n\npackage main\n"), 0o644)
	valid, _ := (&SPDXFormat{}).HasValid(p, lineSlash, "Acme Corp", "MIT")
	if !valid {
		t.Error("HasValid should accept any year")
	}
}

func TestSPDX_HasValid_1Line(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// SPDX-License-Identifier: MIT\n\npackage main\n"), 0o644)
	valid, _ := (&SPDXFormat{}).HasValid(p, lineSlash, "", "MIT")
	if !valid {
		t.Error("expected valid for 1-line")
	}
}

func TestSPDX_HasValid_Missing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("package main\n"), 0o644)
	valid, _ := (&SPDXFormat{}).HasValid(p, lineSlash, "Acme Corp", "MIT")
	if valid {
		t.Error("expected invalid for missing header")
	}
}

func TestSPDX_HasValid_WrongLicense(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// Copyright 2026 Acme Corp\n// SPDX-License-Identifier: MIT\n\npackage main\n"), 0o644)
	valid, _ := (&SPDXFormat{}).HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if valid {
		t.Error("expected invalid for wrong license")
	}
}

func TestSPDX_HasValid_WrongHolder(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// Copyright 2026 Other\n// SPDX-License-Identifier: MIT\n\npackage main\n"), 0o644)
	valid, _ := (&SPDXFormat{}).HasValid(p, lineSlash, "Acme Corp", "MIT")
	if valid {
		t.Error("expected invalid for wrong holder")
	}
}

func TestSPDX_HasValid_Block(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte("/*\n Copyright 2026 Acme Corp\n SPDX-License-Identifier: MIT\n*/\n\n.body {}\n"), 0o644)
	valid, _ := (&SPDXFormat{}).HasValid(p, blockC, "Acme Corp", "MIT")
	if !valid {
		t.Error("expected valid for block")
	}
}

func TestSPDX_StripExisting_2Line(t *testing.T) {
	f := &SPDXFormat{}
	got := string(f.StripExisting([]byte("// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n"), lineSlash))
	if strings.Contains(got, "SPDX-License-Identifier") {
		t.Error("header should be stripped")
	}
	if !strings.Contains(got, "package main") {
		t.Error("code should be preserved")
	}
}

func TestSPDX_StripExisting_NoHeader(t *testing.T) {
	f := &SPDXFormat{}
	src := "package main\n"
	if got := string(f.StripExisting([]byte(src), lineSlash)); got != src {
		t.Error("unchanged file should stay unchanged")
	}
}

func TestSPDX_Idempotency(t *testing.T) {
	dir := t.TempDir()
	f := &SPDXFormat{}
	for name, style := range map[string]*language.Style{"slash": lineSlash, "hash": lineHash, "dash": lineDash} {
		hdr := f.Generate(style, "2026", "Acme Corp", "MIT")
		p := filepath.Join(dir, "test_"+name+".go")
		os.WriteFile(p, []byte(hdr+"\n\npackage main\n"), 0o644)
		valid, err := f.HasValid(p, style, "Acme Corp", "MIT")
		if err != nil {
			t.Errorf("style %s: %v", name, err)
		}
		if !valid {
			t.Errorf("style %s: Generate → HasValid should be true", name)
		}
	}
}

func TestSPDX_Idempotency_Block(t *testing.T) {
	dir := t.TempDir()
	f := &SPDXFormat{}
	hdr := f.Generate(blockC, "2026", "Acme Corp", "MIT")
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte(hdr+"\n\n.body {}\n"), 0o644)
	valid, _ := f.HasValid(p, blockC, "Acme Corp", "MIT")
	if !valid {
		t.Error("block idempotency failed")
	}
}
