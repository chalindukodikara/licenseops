// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApacheLong_Generate_Line(t *testing.T) {
	got := (&ApacheLongFormat{}).Generate(lineSlash, "2026", "Acme Corp", "Apache-2.0")
	if !strings.Contains(got, "Copyright 2026 Acme Corp") || !strings.Contains(got, "Licensed under the Apache License") {
		t.Error("missing expected content")
	}
}

func TestApacheLong_Generate_Block(t *testing.T) {
	got := (&ApacheLongFormat{}).Generate(blockC, "2026", "Acme Corp", "Apache-2.0")
	if !strings.HasPrefix(got, "/*") || !strings.HasSuffix(got, "*/") {
		t.Errorf("block wrong:\n%s", got)
	}
}

func TestApacheLong_HasValid(t *testing.T) {
	dir := t.TempDir()
	f := &ApacheLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "Apache-2.0")
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(hdr+"\n\npackage main\n"), 0o644)
	valid, _ := f.HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if !valid {
		t.Error("expected valid")
	}
}

func TestApacheLong_HasValid_Missing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("package main\n"), 0o644)
	valid, _ := (&ApacheLongFormat{}).HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if valid {
		t.Error("expected invalid")
	}
}

func TestApacheLong_StripExisting(t *testing.T) {
	f := &ApacheLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "Apache-2.0")
	got := string(f.StripExisting([]byte(hdr+"\n\npackage main\n"), lineSlash))
	if strings.Contains(got, "Apache License") {
		t.Error("should be stripped")
	}
	if !strings.Contains(got, "package main") {
		t.Error("code should be preserved")
	}
}

func TestApacheLong_Idempotency(t *testing.T) {
	dir := t.TempDir()
	f := &ApacheLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Acme Corp", "Apache-2.0")
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(hdr+"\n\npackage main\n"), 0o644)
	valid, _ := f.HasValid(p, lineSlash, "Acme Corp", "Apache-2.0")
	if !valid {
		t.Error("idempotency failed")
	}
}

func TestApacheLong_Name(t *testing.T) {
	if (&ApacheLongFormat{}).Name() != "apache-long" {
		t.Error("Name should be 'apache-long'")
	}
}

func TestApacheLong_Generate_Hash(t *testing.T) {
	got := (&ApacheLongFormat{}).Generate(lineHash, "2026", "Acme", "Apache-2.0")
	if !strings.HasPrefix(got, "# Copyright 2026 Acme") {
		t.Errorf("hash style wrong:\n%s", got)
	}
	if !strings.Contains(got, "# Licensed under the Apache License") {
		t.Errorf("hash anchor missing:\n%s", got)
	}
}

func TestApacheLong_HasValid_TooShort(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte("// Copyright 2026 Acme\n// short\npackage main\n"), 0o644)
	valid, _ := (&ApacheLongFormat{}).HasValid(p, lineSlash, "Acme", "Apache-2.0")
	if valid {
		t.Error("short header should not be valid Apache")
	}
}

func TestApacheLong_HasValid_WrongHolder(t *testing.T) {
	dir := t.TempDir()
	f := &ApacheLongFormat{}
	hdr := f.Generate(lineSlash, "2026", "Other", "Apache-2.0")
	p := filepath.Join(dir, "main.go")
	os.WriteFile(p, []byte(hdr+"\n\npackage main\n"), 0o644)
	valid, _ := f.HasValid(p, lineSlash, "Acme", "Apache-2.0")
	if valid {
		t.Error("expected invalid for wrong holder")
	}
}

func TestApacheLong_StripExisting_Block(t *testing.T) {
	f := &ApacheLongFormat{}
	hdr := f.Generate(blockC, "2026", "Acme", "Apache-2.0")
	got := string(f.StripExisting([]byte(hdr+"\n\n.body {}\n"), blockC))
	if strings.Contains(got, "Apache License") {
		t.Errorf("block Apache header should be stripped, got:\n%s", got)
	}
}
