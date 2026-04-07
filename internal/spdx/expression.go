// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package spdx

import (
	"fmt"
	"sort"
	"strings"
)

// ValidLicenses is the set of recognized SPDX identifiers.
var ValidLicenses = map[string]bool{
	"Apache-2.0":        true,
	"MIT":               true,
	"GPL-2.0-only":      true,
	"GPL-2.0-or-later":  true,
	"GPL-3.0-only":      true,
	"GPL-3.0-or-later":  true,
	"LGPL-2.1-only":     true,
	"LGPL-2.1-or-later": true,
	"LGPL-3.0-only":     true,
	"LGPL-3.0-or-later": true,
	"BSD-2-Clause":      true,
	"BSD-3-Clause":      true,
	"MPL-2.0":           true,
	"AGPL-3.0-only":     true,
	"AGPL-3.0-or-later": true,
	"ISC":               true,
	"Unlicense":         true,
	"BSL-1.0":           true,
	"0BSD":              true,
	"CC0-1.0":           true,
	"EPL-1.0":           true,
	"EPL-2.0":           true,
	"EUPL-1.2":          true,
	"Artistic-2.0":      true,
	"Zlib":              true,
	"PostgreSQL":        true,
	"AFL-3.0":           true,
	"WTFPL":             true,
	"MulanPSL-2.0":      true,
}

// DeprecatedGNU maps bare deprecated GNU IDs to their suggested replacements.
var DeprecatedGNU = map[string]string{
	"GPL-2.0":  "GPL-2.0-only or GPL-2.0-or-later",
	"GPL-3.0":  "GPL-3.0-only or GPL-3.0-or-later",
	"LGPL-2.1": "LGPL-2.1-only or LGPL-2.1-or-later",
	"LGPL-3.0": "LGPL-3.0-only or LGPL-3.0-or-later",
	"AGPL-3.0": "AGPL-3.0-only or AGPL-3.0-or-later",
}

// operators are SPDX expression keywords.
var operators = map[string]bool{
	"AND":  true,
	"OR":   true,
	"WITH": true,
}

// IsExpression returns true if the license string contains SPDX expression operators.
func IsExpression(license string) bool {
	for _, token := range tokenize(license) {
		if operators[token] || token == "(" || token == ")" {
			return true
		}
	}
	return false
}

// ValidateExpression validates an SPDX license expression.
// It returns a list of warnings (e.g. deprecated IDs) and an error if invalid.
func ValidateExpression(expr string) (warnings []string, err error) {
	tokens := tokenize(expr)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty license expression")
	}

	for _, token := range tokens {
		// Skip operators and parens
		if operators[token] || token == "(" || token == ")" {
			continue
		}

		// Handle the + suffix (e.g. EPL-1.0+)
		id := strings.TrimSuffix(token, "+")

		// Check for deprecated GNU IDs (case-insensitive)
		if suggestion, ok := findDeprecated(id); ok {
			warnings = append(warnings,
				fmt.Sprintf("deprecated SPDX ID %q — use %s instead", id, suggestion))
			continue // still valid, just deprecated
		}

		if !ValidLicenses[id] {
			return warnings, unknownLicenseError(id)
		}
	}

	return warnings, nil
}

// ValidateLicense validates a single license ID or expression.
func ValidateLicense(license string) (warnings []string, err error) {
	if IsExpression(license) {
		return ValidateExpression(license)
	}

	// Single ID
	id := strings.TrimSuffix(license, "+")

	// Check for deprecated GNU IDs (case-insensitive)
	if suggestion, ok := findDeprecated(id); ok {
		return []string{fmt.Sprintf("deprecated SPDX ID %q — use %s instead", id, suggestion)}, nil
	}

	if !ValidLicenses[id] {
		return nil, unknownLicenseError(id)
	}

	return nil, nil
}

// IsGPLFamily returns true if the license ID is a GPL/LGPL/AGPL variant.
func IsGPLFamily(license string) bool {
	id := license
	if IsExpression(license) {
		return false // expressions aren't valid for gpl-long format
	}
	return strings.HasPrefix(id, "GPL-") ||
		strings.HasPrefix(id, "LGPL-") ||
		strings.HasPrefix(id, "AGPL-")
}

// ValidLicenseList returns a sorted list of all valid SPDX identifiers.
func ValidLicenseList() []string {
	ids := make([]string, 0, len(ValidLicenses))
	for id := range ValidLicenses {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// findDeprecated does a case-insensitive lookup in DeprecatedGNU.
func findDeprecated(id string) (suggestion string, ok bool) {
	if s, found := DeprecatedGNU[id]; found {
		return s, true
	}
	lower := strings.ToLower(id)
	for dep, s := range DeprecatedGNU {
		if strings.ToLower(dep) == lower {
			return s, true
		}
	}
	return "", false
}

// SuggestLicense returns valid license IDs similar to the input,
// using case-insensitive exact, prefix, and substring matching.
func SuggestLicense(input string) []string {
	lower := strings.ToLower(input)
	var exact, prefix, contains []string

	for id := range ValidLicenses {
		idLower := strings.ToLower(id)
		switch {
		case idLower == lower:
			exact = append(exact, id)
		case strings.HasPrefix(idLower, lower):
			prefix = append(prefix, id)
		case strings.Contains(idLower, lower):
			contains = append(contains, id)
		}
	}

	if len(exact) > 0 {
		sort.Strings(exact)
		return exact
	}
	if len(prefix) > 0 {
		sort.Strings(prefix)
		return prefix
	}
	if len(contains) > 0 {
		sort.Strings(contains)
		return contains
	}
	return nil
}

// unknownLicenseError builds an error for an unrecognized license ID,
// showing "did you mean?" suggestions when possible.
func unknownLicenseError(id string) error {
	if suggestions := SuggestLicense(id); len(suggestions) > 0 {
		return fmt.Errorf("unknown SPDX license identifier: %q\n  did you mean: %s", id, strings.Join(suggestions, ", "))
	}
	return fmt.Errorf("unknown SPDX license identifier: %q\n  valid IDs: %s", id, formatIDList(ValidLicenseList()))
}

// formatIDList formats a list of IDs with line wrapping for readability.
func formatIDList(ids []string) string {
	const maxLen = 68
	var b strings.Builder
	lineLen := 0
	for i, id := range ids {
		if i > 0 {
			if lineLen+2+len(id) > maxLen {
				b.WriteString(",\n    ")
				lineLen = 4
			} else {
				b.WriteString(", ")
				lineLen += 2
			}
		}
		b.WriteString(id)
		lineLen += len(id)
	}
	return b.String()
}

// tokenize splits an SPDX expression into tokens.
func tokenize(expr string) []string {
	// Replace parens with space-padded versions for splitting
	r := strings.NewReplacer("(", " ( ", ")", " ) ")
	padded := r.Replace(expr)
	fields := strings.Fields(padded)
	return fields
}
