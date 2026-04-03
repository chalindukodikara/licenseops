// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/chalindukodikara/licenseops/internal/language"
)

// SPDXFormat implements the SPDX short-form header format.
// Supports both 2-line (Copyright + SPDX) and 1-line (SPDX only) modes.
// Mode is determined by whether holder is empty.
type SPDXFormat struct{}

func (f *SPDXFormat) Name() string { return "spdx" }

func (f *SPDXFormat) Generate(style *language.Style, year, holder, license string) string {
	spdxLine := fmt.Sprintf("SPDX-License-Identifier: %s", license)

	if holder == "" {
		// 1-line mode
		if style.IsBlock() {
			return fmt.Sprintf("%s %s %s", style.BlockStart, spdxLine, style.BlockEnd)
		}
		return fmt.Sprintf("%s %s", style.LinePrefix, spdxLine)
	}

	// 2-line mode
	copyrightLine := fmt.Sprintf("Copyright %s %s", year, holder)
	if style.IsBlock() {
		return fmt.Sprintf("%s\n %s\n %s\n%s",
			style.BlockStart, copyrightLine, spdxLine, style.BlockEnd)
	}
	return fmt.Sprintf("%s %s\n%s %s",
		style.LinePrefix, copyrightLine, style.LinePrefix, spdxLine)
}

func (f *SPDXFormat) HasValid(path string, style *language.Style, holder, license string) (bool, error) {
	lines, err := ReadHeaderLines(path, 5)
	if err != nil {
		return false, err
	}

	if holder == "" {
		return f.hasValid1Line(lines, style, license), nil
	}

	if style.IsBlock() {
		return f.hasValidBlock2Line(lines, style, holder, license), nil
	}
	return f.hasValidLine2Line(lines, style, holder, license), nil
}

func (f *SPDXFormat) hasValid1Line(lines []string, style *language.Style, license string) bool {
	if len(lines) < 2 {
		return false
	}

	if style.IsBlock() {
		// Single-line block: /* SPDX-License-Identifier: MIT */
		expected := fmt.Sprintf("%s SPDX-License-Identifier: %s %s",
			style.BlockStart, license, style.BlockEnd)
		if strings.TrimSpace(lines[0]) != expected {
			return false
		}
	} else {
		spdxPat := fmt.Sprintf(`^%s SPDX-License-Identifier: %s$`,
			regexp.QuoteMeta(style.LinePrefix), regexp.QuoteMeta(license))
		if matched, _ := regexp.MatchString(spdxPat, lines[0]); !matched {
			return false
		}
	}

	// Next line should be blank
	if strings.TrimSpace(lines[1]) != "" {
		return false
	}
	return true
}

func (f *SPDXFormat) hasValidLine2Line(lines []string, style *language.Style, holder, license string) bool {
	if len(lines) < 3 {
		return false
	}
	prefix := style.LinePrefix

	copyrightPat := fmt.Sprintf(`^%s Copyright \d{4} %s$`,
		regexp.QuoteMeta(prefix), regexp.QuoteMeta(holder))
	if matched, _ := regexp.MatchString(copyrightPat, lines[0]); !matched {
		return false
	}

	spdxPat := fmt.Sprintf(`^%s SPDX-License-Identifier: %s$`,
		regexp.QuoteMeta(prefix), regexp.QuoteMeta(license))
	if matched, _ := regexp.MatchString(spdxPat, lines[1]); !matched {
		return false
	}

	if strings.TrimSpace(lines[2]) != "" {
		return false
	}
	return true
}

func (f *SPDXFormat) hasValidBlock2Line(lines []string, style *language.Style, holder, license string) bool {
	if len(lines) < 4 {
		return false
	}

	if strings.TrimSpace(lines[0]) != style.BlockStart {
		return false
	}

	copyrightPat := fmt.Sprintf(`^\s*Copyright \d{4} %s$`, regexp.QuoteMeta(holder))
	if matched, _ := regexp.MatchString(copyrightPat, strings.TrimPrefix(lines[1], " ")); !matched {
		return false
	}

	spdxPat := fmt.Sprintf(`^\s*SPDX-License-Identifier: %s$`, regexp.QuoteMeta(license))
	if matched, _ := regexp.MatchString(spdxPat, strings.TrimPrefix(lines[2], " ")); !matched {
		return false
	}

	if strings.TrimSpace(lines[3]) != style.BlockEnd {
		return false
	}
	return true
}

func (f *SPDXFormat) StripExisting(src []byte, style *language.Style) []byte {
	lines := strings.Split(string(src), "\n")
	i := skipPreamble(lines)
	headerStart := i

	if style.IsBlock() {
		i = stripBlockHeaderGeneric(lines, i, style)
	} else {
		i = stripSPDXLineHeader(lines, i)
	}

	if i == headerStart {
		return src
	}
	return reconstructAfterStrip(lines, src, i)
}

func stripSPDXLineHeader(lines []string, i int) int {
	// Check for copyright line
	if i < len(lines) && reCopyrightLine.MatchString(lines[i]) {
		if i+1 < len(lines) && reSPDXLine.MatchString(lines[i+1]) {
			return i + 2
		}
		return i + 1
	}
	// Check for SPDX-only line
	if i < len(lines) && reSPDXLine.MatchString(lines[i]) {
		return i + 1
	}
	return i
}
