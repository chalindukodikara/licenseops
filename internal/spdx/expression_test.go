// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package spdx

import (
	"strings"
	"testing"
)

func TestValidateLicense_SimpleID(t *testing.T) {
	if _, err := ValidateLicense("Apache-2.0"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateLicense_Empty(t *testing.T) {
	if _, err := ValidateLicense(""); err == nil {
		t.Error("expected error for empty license")
	}
}

func TestValidateLicense_Unknown(t *testing.T) {
	_, err := ValidateLicense("INVALID-1.0")
	if err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention 'unknown': %v", err)
	}
}

func TestValidateLicense_DeprecatedGNU(t *testing.T) {
	for _, id := range []string{"GPL-2.0", "GPL-3.0", "LGPL-2.1", "LGPL-3.0", "AGPL-3.0"} {
		warnings, err := ValidateLicense(id)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", id, err)
		}
		if len(warnings) == 0 {
			t.Errorf("%s: expected deprecation warning", id)
		}
	}
}

func TestValidateLicense_PlusSuffix(t *testing.T) {
	if _, err := ValidateLicense("EPL-1.0+"); err != nil {
		t.Errorf("EPL-1.0+ should be valid: %v", err)
	}
}

func TestValidateExpression_OR(t *testing.T) {
	if _, err := ValidateExpression("Apache-2.0 OR MIT"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateExpression_Complex(t *testing.T) {
	if _, err := ValidateExpression("(Apache-2.0 AND MIT) OR BSD-3-Clause"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateExpression_InvalidID(t *testing.T) {
	if _, err := ValidateExpression("Apache-2.0 OR INVALID-1.0"); err == nil {
		t.Error("expected error")
	}
}

func TestValidateExpression_Empty(t *testing.T) {
	if _, err := ValidateExpression(""); err == nil {
		t.Error("expected error for empty")
	}
}

func TestIsExpression(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"MIT", false},
		{"Apache-2.0 OR MIT", true},
		{"Apache-2.0 AND MIT", true},
		{"(Apache-2.0 OR MIT)", true},
	}
	for _, tt := range tests {
		if got := IsExpression(tt.input); got != tt.want {
			t.Errorf("IsExpression(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsGPLFamily(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"GPL-3.0-only", true},
		{"LGPL-2.1-only", true},
		{"AGPL-3.0-only", true},
		{"MIT", false},
		{"Apache-2.0 OR MIT", false},
	}
	for _, tt := range tests {
		if got := IsGPLFamily(tt.input); got != tt.want {
			t.Errorf("IsGPLFamily(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestValidLicenseList_Sorted(t *testing.T) {
	list := ValidLicenseList()
	for i := 1; i < len(list); i++ {
		if list[i] < list[i-1] {
			t.Errorf("not sorted: %q before %q", list[i-1], list[i])
		}
	}
}

func TestSuggestLicense_CaseInsensitive(t *testing.T) {
	suggestions := SuggestLicense("mit")
	found := false
	for _, s := range suggestions {
		if s == "MIT" {
			found = true
		}
	}
	if !found {
		t.Errorf("should suggest MIT, got %v", suggestions)
	}
}

func TestSuggestLicense_NoMatch(t *testing.T) {
	if got := SuggestLicense("zzzznotareallicense"); len(got) != 0 {
		t.Errorf("expected no suggestions, got %v", got)
	}
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("(Apache-2.0 AND MIT) OR BSD-3-Clause")
	expected := []string{"(", "Apache-2.0", "AND", "MIT", ")", "OR", "BSD-3-Clause"}
	if len(tokens) != len(expected) {
		t.Fatalf("got %v, want %v", tokens, expected)
	}
	for i, tok := range tokens {
		if tok != expected[i] {
			t.Errorf("token[%d] = %q, want %q", i, tok, expected[i])
		}
	}
}
