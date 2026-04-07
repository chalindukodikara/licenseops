// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGPLLong_Generate_GPL3(t *testing.T) {
	got := (&GPLLongFormat{}).Generate(lineSlash, "2026", "Acme Corp", "GPL-3.0-only")
	if !strings.Contains(got, "GNU General Public License") || !strings.Contains(got, "version 3") {
		t.Error("missing expected content")
	}
}

func TestGPLLong_Generate_LGPL(t *testing.T) {
	got := (&GPLLongFormat{}).Generate(lineSlash, "2026", "Acme Corp", "LGPL-2.1-only")
	if !strings.Contains(got, "GNU Lesser General Public License") {
		t.Error("should use 'Lesser' wording")
	}
}

func TestGPLLong_Generate_AGPL(t *testing.T) {
	got := (&GPLLongFormat{}).Generate(lineSlash, "2026", "Acme Corp", "AGPL-3.0-only")
	if !strings.Contains(got, "GNU Affero General Public License") {
		t.Error("should use 'Affero' wording")
	}
}

func TestGPLLong_HasValid(t *testing.T) {
	dir := t.TempDir()
	f := &GPLLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "GPL-3.0-only")
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(hdr+"\n\npackage main\n"), 0o644)
	valid, _ := f.HasValid(p, lineSlash, "Acme Corp", "GPL-3.0-only")
	if !valid {
		t.Error("expected valid")
	}
}

func TestGPLLong_StripExisting(t *testing.T) {
	f := &GPLLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "GPL-3.0-only")
	got := string(f.StripExisting([]byte(hdr+"\n\npackage main\n"), lineSlash))
	if strings.Contains(got, "GNU General Public License") {
		t.Error("should be stripped")
	}
}

func TestGPLLong_Idempotency(t *testing.T) {
	dir := t.TempDir()
	f := &GPLLongFormat{}
	for _, lic := range []string{"GPL-3.0-only", "LGPL-2.1-only", "AGPL-3.0-only"} {
		hdr := f.Generate(lineSlash, "2026", "Acme Corp", lic)
		p := filepath.Join(dir, "test_"+lic+".go")
		os.WriteFile(p, []byte(hdr+"\n\npackage main\n"), 0o644)
		valid, _ := f.HasValid(p, lineSlash, "Acme Corp", lic)
		if !valid {
			t.Errorf("%s: idempotency failed", lic)
		}
	}
}
