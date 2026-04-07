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

func TestReuse_Generate_Line(t *testing.T) {
	got := (&ReuseFormat{}).Generate(lineSlash, "2026", "Acme Corp", "MIT")
	want := "// SPDX-FileCopyrightText: 2026 Acme Corp\n// SPDX-License-Identifier: MIT"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestReuse_Generate_Block(t *testing.T) {
	got := (&ReuseFormat{}).Generate(blockC, "2026", "Acme Corp", "MIT")
	if !strings.HasPrefix(got, "/*") || !strings.Contains(got, "SPDX-FileCopyrightText:") {
		t.Errorf("unexpected: %s", got)
	}
}

func TestReuse_HasValid_Line(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// SPDX-FileCopyrightText: 2026 Acme Corp\n// SPDX-License-Identifier: MIT\n\npackage main\n"), 0o644)
	valid, _ := (&ReuseFormat{}).HasValid(p, lineSlash, "Acme Corp", "MIT")
	if !valid {
		t.Error("expected valid")
	}
}

func TestReuse_HasValid_Missing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("package main\n"), 0o644)
	valid, _ := (&ReuseFormat{}).HasValid(p, lineSlash, "Acme Corp", "MIT")
	if valid {
		t.Error("expected invalid")
	}
}

func TestReuse_StripExisting(t *testing.T) {
	got := string((&ReuseFormat{}).StripExisting([]byte("// SPDX-FileCopyrightText: 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n"), lineSlash))
	if strings.Contains(got, "SPDX-FileCopyrightText") {
		t.Error("should be stripped")
	}
}

func TestReuse_Idempotency(t *testing.T) {
	dir := t.TempDir()
	f := &ReuseFormat{}
	for name, style := range map[string]*language.Style{"slash": lineSlash, "hash": lineHash, "dash": lineDash} {
		hdr := f.Generate(style, "2026", "Acme Corp", "MIT")
		p := filepath.Join(dir, "reuse_"+name+".go")
		os.WriteFile(p, []byte(hdr+"\n\npackage main\n"), 0o644)
		valid, _ := f.HasValid(p, style, "Acme Corp", "MIT")
		if !valid {
			t.Errorf("style %s: idempotency failed", name)
		}
	}
}
