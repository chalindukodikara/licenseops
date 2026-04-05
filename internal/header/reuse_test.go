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

// --- Reuse Generate ---

func TestReuse_Generate_Line(t *testing.T) {
	f := &ReuseFormat{}
	got := f.Generate(lineSlash, "2026", "Acme Corp", "MIT")
	want := "// SPDX-FileCopyrightText: 2026 Acme Corp\n// SPDX-License-Identifier: MIT"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestReuse_Generate_Hash(t *testing.T) {
	f := &ReuseFormat{}
	got := f.Generate(lineHash, "2026", "Acme Corp", "Apache-2.0")
	if !strings.HasPrefix(got, "# SPDX-FileCopyrightText:") {
		t.Errorf("unexpected output: %s", got)
	}
}

func TestReuse_Generate_Block(t *testing.T) {
	f := &ReuseFormat{}
	got := f.Generate(blockC, "2026", "Acme Corp", "MIT")
	if !strings.HasPrefix(got, "/*") || !strings.HasSuffix(got, "*/") {
		t.Errorf("block should be wrapped in /* */:\n%s", got)
	}
	if !strings.Contains(got, "SPDX-FileCopyrightText:") {
		t.Error("missing SPDX-FileCopyrightText line")
	}
}

// --- Reuse HasValid ---

func TestReuse_HasValid_Line(t *testing.T) {
	dir := t.TempDir()
	content := "// SPDX-FileCopyrightText: 2026 Acme Corp\n// SPDX-License-Identifier: MIT\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &ReuseFormat{}
	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid=true")
	}
}

func TestReuse_HasValid_DifferentYear(t *testing.T) {
	dir := t.TempDir()
	content := "# SPDX-FileCopyrightText: 2024 Acme Corp\n# SPDX-License-Identifier: MIT\n\nimport os\n"
	p := filepath.Join(dir, "script.py")
	os.WriteFile(p, []byte(content), 0o644)

	f := &ReuseFormat{}
	valid, err := f.HasValid(p, lineHash, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("HasValid should accept any year")
	}
}

func TestReuse_HasValid_Missing(t *testing.T) {
	dir := t.TempDir()
	content := "package main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &ReuseFormat{}
	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected valid=false for missing header")
	}
}

func TestReuse_HasValid_Block(t *testing.T) {
	dir := t.TempDir()
	content := "/*\n SPDX-FileCopyrightText: 2026 Acme Corp\n SPDX-License-Identifier: MIT\n*/\n\n.body {}\n"
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte(content), 0o644)

	f := &ReuseFormat{}
	valid, err := f.HasValid(p, blockC, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid=true for block comment")
	}
}

// --- Reuse StripExisting ---

func TestReuse_StripExisting(t *testing.T) {
	f := &ReuseFormat{}
	src := "// SPDX-FileCopyrightText: 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if strings.Contains(got, "SPDX-FileCopyrightText") {
		t.Error("reuse header should be stripped")
	}
	if !strings.Contains(got, "package main") {
		t.Error("code should be preserved")
	}
}

func TestReuse_StripExisting_NoHeader(t *testing.T) {
	f := &ReuseFormat{}
	src := "package main\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if got != src {
		t.Error("file with no header should be unchanged")
	}
}

// --- Reuse Idempotency ---

func TestReuse_Idempotency(t *testing.T) {
	dir := t.TempDir()
	f := &ReuseFormat{}

	styles := map[string]*language.Style{"slash": lineSlash, "hash": lineHash, "dash": lineDash}
	for name, style := range styles {
		hdr := f.Generate(style, "2026", "Acme Corp", "MIT")
		content := hdr + "\n\npackage main\n"
		p := filepath.Join(dir, "test_reuse_"+name+".go")
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
