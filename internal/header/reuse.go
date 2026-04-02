package header

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/chalindu/licenser/internal/language"
)

// ReuseFormat implements the FSFE REUSE specification header format.
// Uses SPDX-FileCopyrightText instead of Copyright.
type ReuseFormat struct{}

func (f *ReuseFormat) Name() string { return "reuse" }

func (f *ReuseFormat) Generate(style *language.Style, year, holder, license string) string {
	copyrightLine := fmt.Sprintf("SPDX-FileCopyrightText: %s %s", year, holder)
	spdxLine := fmt.Sprintf("SPDX-License-Identifier: %s", license)

	if style.IsBlock() {
		return fmt.Sprintf("%s\n %s\n %s\n%s",
			style.BlockStart, copyrightLine, spdxLine, style.BlockEnd)
	}
	return fmt.Sprintf("%s %s\n%s %s",
		style.LinePrefix, copyrightLine, style.LinePrefix, spdxLine)
}

func (f *ReuseFormat) HasValid(path string, style *language.Style, holder, license string) (bool, error) {
	lines, err := ReadHeaderLines(path, 5)
	if err != nil {
		return false, err
	}

	if style.IsBlock() {
		return f.hasValidBlock(lines, style, holder, license), nil
	}
	return f.hasValidLine(lines, style, holder, license), nil
}

func (f *ReuseFormat) hasValidLine(lines []string, style *language.Style, holder, license string) bool {
	if len(lines) < 3 {
		return false
	}
	prefix := style.LinePrefix

	copyrightPat := fmt.Sprintf(`^%s SPDX-FileCopyrightText: \d{4} %s$`,
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

func (f *ReuseFormat) hasValidBlock(lines []string, style *language.Style, holder, license string) bool {
	if len(lines) < 4 {
		return false
	}

	if strings.TrimSpace(lines[0]) != style.BlockStart {
		return false
	}

	copyrightPat := fmt.Sprintf(`^\s*SPDX-FileCopyrightText: \d{4} %s$`, regexp.QuoteMeta(holder))
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

func (f *ReuseFormat) StripExisting(src []byte, style *language.Style) []byte {
	lines := strings.Split(string(src), "\n")
	i := skipPreamble(lines)
	headerStart := i

	if style.IsBlock() {
		i = stripBlockHeaderGeneric(lines, i, style)
	} else {
		i = stripReuseLineHeader(lines, i)
	}

	if i == headerStart {
		return src
	}
	return reconstructAfterStrip(lines, src, i)
}

func stripReuseLineHeader(lines []string, i int) int {
	if i < len(lines) && reReuseCopyrightLine.MatchString(lines[i]) {
		if i+1 < len(lines) && reSPDXLine.MatchString(lines[i+1]) {
			return i + 2
		}
		return i + 1
	}
	return i
}
