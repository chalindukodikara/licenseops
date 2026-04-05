// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- ApacheLong Generate ---

func TestApacheLong_Generate_Line(t *testing.T) {
	f := &ApacheLongFormat{}
	got := f.Generate(lineSlash, "2026", "Acme Corp", "Apache-2.0")
	if !strings.Contains(got, "Copyright 2026 Acme Corp") {
		t.Error("missing copyright line")
	}
	if !strings.Contains(got, "Licensed under the Apache License") {
		t.Error("missing Apache anchor")
	}
	if !strings.Contains(got, "http://www.apache.org/licenses/LICENSE-2.0") {
		t.Error("missing license URL")
	}
	// All lines should start with //
	for _, line := range strings.Split(got, "\n") {
		if !strings.HasPrefix(line, "//") {
			t.Errorf("line should start with //: %q", line)
		}
	}
}

func TestApacheLong_Generate_Hash(t *testing.T) {
	f := &ApacheLongFormat{}
	got := f.Generate(lineHash, "2026", "Acme Corp", "Apache-2.0")
	for _, line := range strings.Split(got, "\n") {
		if !strings.HasPrefix(line, "#") {
			t.Errorf("line should start with #: %q", line)
		}
	}
}

func TestApacheLong_Generate_Block(t *testing.T) {
	f := &ApacheLongFormat{}
	got := f.Generate(blockC, "2026", "Acme Corp", "Apache-2.0")
	if !strings.HasPrefix(got, "/*") {
		t.Error("block should start with /*")
	}
	if !strings.HasSuffix(got, "*/") {
		t.Error("block should end with */")
	}
}

// --- ApacheLong HasValid ---

func TestApacheLong_HasValid(t *testing.T) {
	dir := t.TempDir()
	f := &ApacheLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "Apache-2.0")
	content := hdr + "\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected valid=true for generated Apache header")
	}
}

func TestApacheLong_HasValid_Missing(t *testing.T) {
	dir := t.TempDir()
	content := "package main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	f := &ApacheLongFormat{}
	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected valid=false for missing header")
	}
}

func TestApacheLong_HasValid_WrongHolder(t *testing.T) {
	dir := t.TempDir()
	f := &ApacheLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Other Corp", "Apache-2.0")
	content := hdr + "\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected valid=false for wrong holder")
	}
}

// --- ApacheLong StripExisting ---

func TestApacheLong_StripExisting(t *testing.T) {
	f := &ApacheLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "Apache-2.0")
	src := hdr + "\n\npackage main\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if strings.Contains(got, "Apache License") {
		t.Error("Apache header should be stripped")
	}
	if !strings.Contains(got, "package main") {
		t.Error("code should be preserved")
	}
}

func TestApacheLong_StripExisting_NoHeader(t *testing.T) {
	f := &ApacheLongFormat{}
	src := "package main\n"
	got := string(f.StripExisting([]byte(src), lineSlash))
	if got != src {
		t.Error("file with no header should be unchanged")
	}
}

// --- ApacheLong Idempotency ---

func TestApacheLong_Idempotency(t *testing.T) {
	dir := t.TempDir()
	f := &ApacheLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "Apache-2.0")
	content := hdr + "\n\npackage main\n"
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(content), 0o644)

	valid, err := f.HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("Generate → HasValid should be true")
	}
}
