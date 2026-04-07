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

func TestReuse_Name(t *testing.T) {
	if (&ReuseFormat{}).Name() != "reuse" {
		t.Error("Name should be 'reuse'")
	}
}

func TestReuse_HasValid_Block(t *testing.T) {
	dir := t.TempDir()
	f := &ReuseFormat{}
	hdr := f.Generate(blockC, "2026", "Acme Corp", "MIT")
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte(hdr+"\n\n.body {}\n"), 0o644)
	valid, err := f.HasValid(p, blockC, "Acme Corp", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid block REUSE header")
	}
}

func TestReuse_HasValid_Block_WrongStart(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte("notblock\n SPDX-FileCopyrightText: 2026 Acme\n SPDX-License-Identifier: MIT\n*/\n.body {}\n"), 0o644)
	valid, _ := (&ReuseFormat{}).HasValid(p, blockC, "Acme", "MIT")
	if valid {
		t.Error("expected invalid when block start is wrong")
	}
}

func TestReuse_HasValid_Block_WrongHolder(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte("/*\n SPDX-FileCopyrightText: 2026 Other\n SPDX-License-Identifier: MIT\n*/\n\n.body {}\n"), 0o644)
	valid, _ := (&ReuseFormat{}).HasValid(p, blockC, "Acme", "MIT")
	if valid {
		t.Error("expected invalid for wrong holder in block")
	}
}

func TestReuse_HasValid_Block_WrongLicense(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte("/*\n SPDX-FileCopyrightText: 2026 Acme\n SPDX-License-Identifier: GPL-3.0-only\n*/\n\n.body {}\n"), 0o644)
	valid, _ := (&ReuseFormat{}).HasValid(p, blockC, "Acme", "MIT")
	if valid {
		t.Error("expected invalid for wrong license in block")
	}
}

func TestReuse_HasValid_Block_WrongEnd(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "style.css")
	os.WriteFile(p, []byte("/*\n SPDX-FileCopyrightText: 2026 Acme\n SPDX-License-Identifier: MIT\nnotend\n.body {}\n"), 0o644)
	valid, _ := (&ReuseFormat{}).HasValid(p, blockC, "Acme", "MIT")
	if valid {
		t.Error("expected invalid when block end is wrong")
	}
}

func TestReuse_StripExisting_Block(t *testing.T) {
	f := &ReuseFormat{}
	hdr := f.Generate(blockC, "2026", "Acme", "MIT")
	got := string(f.StripExisting([]byte(hdr+"\n\n.body {}\n"), blockC))
	if strings.Contains(got, "SPDX-FileCopyrightText") {
		t.Errorf("block REUSE header should be stripped, got:\n%s", got)
	}
}

func TestReuse_HasValid_TooFewLines(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// SPDX-FileCopyrightText: 2026 Acme\n"), 0o644)
	valid, _ := (&ReuseFormat{}).HasValid(p, lineSlash, "Acme", "MIT")
	if valid {
		t.Error("expected invalid for too-short header")
	}
}

func TestReuse_HasValid_DoesNotMatchCopyrightForm(t *testing.T) {
	// REUSE format requires SPDX-FileCopyrightText, not Copyright
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n"), 0o644)
	valid, _ := (&ReuseFormat{}).HasValid(p, lineSlash, "Acme", "MIT")
	if valid {
		t.Error("REUSE should not accept standard Copyright line")
	}
}
