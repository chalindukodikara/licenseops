// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package spdx

import (
	"strings"
	"testing"
)

// --- ValidateLicense ---

func TestValidateLicense_SimpleID(t *testing.T) {
	warnings, err := ValidateLicense("Apache-2.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(warnings) > 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestValidateLicense_MIT(t *testing.T) {
	_, err := ValidateLicense("MIT")
	if err != nil {
		t.Errorf("MIT should be valid: %v", err)
	}
}

func TestValidateLicense_Empty(t *testing.T) {
	_, err := ValidateLicense("")
	if err == nil {
		t.Error("expected error for empty license")
	}
}

func TestValidateLicense_Unknown(t *testing.T) {
	_, err := ValidateLicense("INVALID-1.0")
	if err == nil {
		t.Error("expected error for unknown license")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention 'unknown': %v", err)
	}
}

func TestValidateLicense_DeprecatedGNU(t *testing.T) {
	tests := []struct {
		id         string
		wantSuffix string
	}{
		{"GPL-2.0", "GPL-2.0-only or GPL-2.0-or-later"},
		{"GPL-3.0", "GPL-3.0-only or GPL-3.0-or-later"},
		{"LGPL-2.1", "LGPL-2.1-only or LGPL-2.1-or-later"},
		{"LGPL-3.0", "LGPL-3.0-only or LGPL-3.0-or-later"},
		{"AGPL-3.0", "AGPL-3.0-only or AGPL-3.0-or-later"},
	}
	for _, tt := range tests {
		warnings, err := ValidateLicense(tt.id)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.id, err)
		}
		if len(warnings) == 0 {
			t.Errorf("%s: expected deprecation warning", tt.id)
			continue
		}
		if !strings.Contains(warnings[0], tt.wantSuffix) {
			t.Errorf("%s: warning %q should mention %q", tt.id, warnings[0], tt.wantSuffix)
		}
	}
}

func TestValidateLicense_PlusSuffix(t *testing.T) {
	// EPL-1.0+ should be valid (+ means "or later")
	_, err := ValidateLicense("EPL-1.0+")
	if err != nil {
		t.Errorf("EPL-1.0+ should be valid: %v", err)
	}
}

// --- ValidateExpression ---

func TestValidateExpression_SimpleOR(t *testing.T) {
	warnings, err := ValidateExpression("Apache-2.0 OR MIT")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(warnings) > 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestValidateExpression_AND(t *testing.T) {
	_, err := ValidateExpression("Apache-2.0 AND MIT")
	if err != nil {
		t.Errorf("AND expression should be valid: %v", err)
	}
}

func TestValidateExpression_WITH(t *testing.T) {
	// WITH isn't validated for exception ID, just the license part
	_, err := ValidateExpression("GPL-3.0-only WITH Classpath-exception-2.0")
	// The exception ID won't be in ValidLicenses — behavior depends on implementation
	// Just check no panic
	_ = err
}

func TestValidateExpression_Complex(t *testing.T) {
	_, err := ValidateExpression("(Apache-2.0 AND MIT) OR BSD-3-Clause")
	if err != nil {
		t.Errorf("complex expression should be valid: %v", err)
	}
}

func TestValidateExpression_InvalidID(t *testing.T) {
	_, err := ValidateExpression("Apache-2.0 OR INVALID-1.0")
	if err == nil {
		t.Error("expected error for invalid ID in expression")
	}
}

func TestValidateExpression_Empty(t *testing.T) {
	_, err := ValidateExpression("")
	if err == nil {
		t.Error("expected error for empty expression")
	}
}

func TestValidateExpression_DeprecatedInExpression(t *testing.T) {
	warnings, err := ValidateExpression("GPL-2.0 OR MIT")
	if err != nil {
		t.Errorf("deprecated ID in expression should not error: %v", err)
	}
	if len(warnings) == 0 {
		t.Error("expected deprecation warning")
	}
}

// --- IsExpression ---

func TestIsExpression(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"MIT", false},
		{"Apache-2.0", false},
		{"Apache-2.0 OR MIT", true},
		{"Apache-2.0 AND MIT", true},
		{"GPL-3.0-only WITH Classpath-exception-2.0", true},
		{"(Apache-2.0 OR MIT)", true},
	}
	for _, tt := range tests {
		got := IsExpression(tt.input)
		if got != tt.want {
			t.Errorf("IsExpression(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// --- IsGPLFamily ---

func TestIsGPLFamily(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"GPL-3.0-only", true},
		{"GPL-2.0-or-later", true},
		{"LGPL-2.1-only", true},
		{"LGPL-3.0-or-later", true},
		{"AGPL-3.0-only", true},
		{"MIT", false},
		{"Apache-2.0", false},
		{"Apache-2.0 OR MIT", false}, // expressions are not GPL family
	}
	for _, tt := range tests {
		got := IsGPLFamily(tt.input)
		if got != tt.want {
			t.Errorf("IsGPLFamily(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// --- ValidLicenseList ---

func TestValidLicenseList_Sorted(t *testing.T) {
	list := ValidLicenseList()
	if len(list) == 0 {
		t.Fatal("ValidLicenseList() returned empty")
	}
	for i := 1; i < len(list); i++ {
		if list[i] < list[i-1] {
			t.Errorf("list not sorted: %q before %q", list[i-1], list[i])
		}
	}
}

func TestValidLicenseList_ContainsCommonLicenses(t *testing.T) {
	list := ValidLicenseList()
	m := make(map[string]bool)
	for _, id := range list {
		m[id] = true
	}
	for _, id := range []string{"MIT", "Apache-2.0", "GPL-3.0-only", "BSD-3-Clause"} {
		if !m[id] {
			t.Errorf("missing common license: %s", id)
		}
	}
}

// --- SuggestLicense ---

func TestSuggestLicense_ExactCaseInsensitive(t *testing.T) {
	suggestions := SuggestLicense("mit")
	found := false
	for _, s := range suggestions {
		if s == "MIT" {
			found = true
		}
	}
	if !found {
		t.Errorf("SuggestLicense('mit') should suggest MIT, got %v", suggestions)
	}
}

func TestSuggestLicense_Prefix(t *testing.T) {
	suggestions := SuggestLicense("GPL")
	if len(suggestions) == 0 {
		t.Error("SuggestLicense('GPL') should return suggestions")
	}
	for _, s := range suggestions {
		if !strings.HasPrefix(strings.ToLower(s), "gpl") {
			t.Errorf("unexpected suggestion: %s", s)
		}
	}
}

func TestSuggestLicense_NoMatch(t *testing.T) {
	suggestions := SuggestLicense("zzzznotareallicense")
	if len(suggestions) != 0 {
		t.Errorf("expected no suggestions, got %v", suggestions)
	}
}

// --- tokenize ---

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
