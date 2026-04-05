// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- GPLLong Generate ---

func TestGPLLong_Generate_GPL3(t *testing.T) {
	f := &GPLLongFormat{}
	got := f.Generate(lineSlash, "2026", "Acme Corp", "GPL-3.0-only")
	if !strings.Contains(got, "Copyright 2026 Acme Corp") {
		t.Error("missing copyright line")
	}
	if !strings.Contains(got, "GNU General Public License") {
		t.Error("missing GPL anchor")
	}
	if !strings.Contains(got, "version 3") {
		t.Error("GPL-3.0 should reference version 3")
	}
}

func TestGPLLong_Generate_LGPL(t *testing.T) {
	f := &GPLLongFormat{}
	got := f.Generate(lineSlash, "2026", "Acme Corp", "LGPL-2.1-only")
	if !strings.Contains(got, "GNU Lesser General Public License") {
		t.Error("LGPL should use 'Lesser' wording")
	}
	if !strings.Contains(got, "version 2") {
		t.Error("LGPL-2.1 should reference version 2")
	}
}

func TestGPLLong_Generate_AGPL(t *testing.T) {
	f := &GPLLongFormat{}
	got := f.Generate(lineSlash, "2026", "Acme Corp", "AGPL-3.0-only")
	if !strings.Contains(got, "GNU Affero General Public License") {
		t.Error("AGPL should use 'Affero' wording")
	}
}

func TestGPLLong_Generate_Hash(t *testing.T) {
	f := &GPLLongFormat{}
	got := f.Generate(lineHash, "2026", "Acme Corp", "GPL-3.0-only")
	for _, line := range strings.Split(got, "\n") {
		if !strings.HasPrefix(line, "#") {
			t.Errorf("line should start with #: %q", line)
		}
	}
}

func TestGPLLong_Generate_Block(t *testing.T) {
	f := &GPLLongFormat{}
	got := f.Generate(blockC, "2026", "Acme Corp", "GPL-3.0-only")
	if !strings.HasPrefix(got, "/*") {
		t.Error("block should start with /*")
	}
	if !strings.HasSuffix(got, "*/") {
		t.Error("block should end with */")
	}
}

// --- GPLLong HasValid ---

func TestGPLLong_HasValid(t *testing.T) {
	dir := t.TempDir()
	f := &GPLLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "GPL-3.0-only")
	content := hdr + "\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "GPL-3.0-only")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid=true")
	}
}

func TestGPLLong_HasValid_Missing(t *testing.T) {
	dir := t.TempDir()
	content := "package main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &GPLLongFormat{}
	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "GPL-3.0-only")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected valid=false")
	}
}

// --- GPLLong StripExisting ---

func TestGPLLong_StripExisting(t *testing.T) {
	f := &GPLLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "GPL-3.0-only")
	src := hdr + "\n\npackage main\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if strings.Contains(got, "GNU General Public License") {
		t.Error("GPL header should be stripped")
	}
	if !strings.Contains(got, "package main") {
		t.Error("code should be preserved")
	}
}

// --- GPLLong Idempotency ---

func TestGPLLong_Idempotency(t *testing.T) {
	dir := t.TempDir()
	f := &GPLLongFormat{}

	licenses := []string{"GPL-3.0-only", "LGPL-2.1-only", "AGPL-3.0-only"}
	for _, lic := range licenses {
		hdr := f.Generate(lineSlash, "2026", "Acme Corp", lic)
		content := hdr + "\n\npackage main\n"
		p := filepath.Join(dir, "test_"+lic+".go")
		os.WriteFile(p, []byte(content), 0o644)

		valid, err := f.HasValid(p, lineSlash, "Acme Corp", lic)
		if err != nil {
			t.Errorf("%s: %v", lic, err)
		}
		if !valid {
			t.Errorf("%s: Generate → HasValid should be true", lic)
		}
	}
}
