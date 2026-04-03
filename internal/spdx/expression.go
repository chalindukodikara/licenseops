// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package spdx

import (
	"fmt"
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

		// Check for deprecated GNU IDs
		if suggestion, ok := DeprecatedGNU[id]; ok {
			warnings = append(warnings,
				fmt.Sprintf("deprecated SPDX ID %q — use %s instead", id, suggestion))
			continue // still valid, just deprecated
		}

		if !ValidLicenses[id] {
			return warnings, fmt.Errorf("unknown SPDX license identifier: %q", id)
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

	if suggestion, ok := DeprecatedGNU[id]; ok {
		return []string{fmt.Sprintf("deprecated SPDX ID %q — use %s instead", id, suggestion)}, nil
	}

	if !ValidLicenses[id] {
		return nil, fmt.Errorf("unknown SPDX license identifier: %q", id)
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

// tokenize splits an SPDX expression into tokens.
func tokenize(expr string) []string {
	// Replace parens with space-padded versions for splitting
	r := strings.NewReplacer("(", " ( ", ")", " ) ")
	padded := r.Replace(expr)
	fields := strings.Fields(padded)
	return fields
}
