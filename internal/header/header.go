// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/licenseops/licenseops/internal/language"
)

// Shared regex patterns used by multiple formats.
var (
	reCopyrightLine     = regexp.MustCompile(`^(//|#|--)\s*Copyright\s+\d{4}\s+.+$`)
	reSPDXLine          = regexp.MustCompile(`^(//|#|--)\s*SPDX-License-Identifier:\s+.+$`)
	reReuseCopyrightLine = regexp.MustCompile(`^(//|#|--)\s*SPDX-FileCopyrightText:\s+.+$`)
	reBlockCopyrightLine = regexp.MustCompile(`^\s*\*?\s*Copyright\s+\d{4}\s+.+$`)
	reBlockSPDXLine      = regexp.MustCompile(`^\s*\*?\s*SPDX-License-Identifier:\s+.+$`)
	reBlockReuseLine     = regexp.MustCompile(`^\s*\*?\s*SPDX-FileCopyrightText:\s+.+$`)
	reGeneratedMarker    = regexp.MustCompile(`(?i)(code generated.*do not ` + `edit|@` + `generated)`)
	reShebang            = regexp.MustCompile(`^#!.+`)
	rePythonEncoding     = regexp.MustCompile(`^#.*(-\*-\s*coding[:=]|coding[:=])\s*\S+`)
	reApacheAnchor       = regexp.MustCompile(`(?i)Licensed under the Apache License`)
	reGPLAnchor          = regexp.MustCompile(`(?i)GNU (General|Lesser|Affero).*Public License`)
)

// IsGenerated checks if a file contains generated-code markers.
func IsGenerated(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 30 && scanner.Scan(); i++ {
		if reGeneratedMarker.MatchString(scanner.Text()) {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// ReadHeaderLines reads the first N meaningful lines from a file,
// skipping leading blank lines, shebangs, and Python encoding declarations.
func ReadHeaderLines(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()

		if len(lines) == 0 {
			if strings.TrimSpace(line) == "" {
				continue
			}
			if reShebang.MatchString(line) {
				continue
			}
			if rePythonEncoding.MatchString(line) {
				continue
			}
		}

		lines = append(lines, line)
		if len(lines) >= n {
			break
		}
	}
	return lines, scanner.Err()
}

// Prepend adds a header to a file, handling shebangs and encoding declarations.
// It uses the given format to strip any existing header first.
func Prepend(path string, style *language.Style, headerText string, fmt Format) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Try stripping with ALL formats and pick the result that removed the most.
	// This handles cross-format migration (e.g. GPL-long → SPDX).
	best := src
	for _, f := range AllFormats() {
		candidate := f.StripExisting(src, style)
		if len(candidate) < len(best) {
			best = candidate
		}
	}
	// Also try the target format
	candidate := fmt.StripExisting(src, style)
	if len(candidate) < len(best) {
		best = candidate
	}
	stripped := best

	lines := strings.Split(string(stripped), "\n")
	i := 0

	var prefixLines []string

	// Shebang
	if i < len(lines) && reShebang.MatchString(lines[i]) {
		prefixLines = append(prefixLines, lines[i])
		i++
		if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
	}

	// Python encoding declarations
	for i < len(lines) && rePythonEncoding.MatchString(lines[i]) {
		prefixLines = append(prefixLines, lines[i])
		i++
	}
	if len(prefixLines) > 0 && i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	rest := strings.Join(lines[i:], "\n")

	var result string
	if len(prefixLines) > 0 {
		result = strings.Join(prefixLines, "\n") + "\n\n" + headerText + "\n\n" + rest
	} else {
		result = headerText + "\n\n" + rest
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(result), info.Mode())
}

// skipPreamble advances past leading blank lines, shebangs, and encoding declarations.
// Returns the index of the first content line that could be a header.
func skipPreamble(lines []string) int {
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i < len(lines) && reShebang.MatchString(lines[i]) {
		i++
		if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
	}
	for i < len(lines) && rePythonEncoding.MatchString(lines[i]) {
		i++
	}
	if i > 0 && i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		// Check if we skipped encoding lines
		if i >= 2 || (i >= 1 && rePythonEncoding.MatchString(lines[i-1])) {
			i++
		}
	}
	return i
}

// reconstructAfterStrip builds the result after header stripping.
func reconstructAfterStrip(lines []string, _ []byte, headerEnd int) []byte {
	// Skip one trailing blank line after the header
	if headerEnd < len(lines) && strings.TrimSpace(lines[headerEnd]) == "" {
		headerEnd++
	}

	// Collect preamble (shebang, encoding)
	var preamble []string
	j := 0
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}
	if j < len(lines) && reShebang.MatchString(lines[j]) {
		preamble = append(preamble, lines[j])
		j++
		if j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
	}
	for j < len(lines) && rePythonEncoding.MatchString(lines[j]) {
		preamble = append(preamble, lines[j])
		j++
	}

	var parts []string
	parts = append(parts, preamble...)
	parts = append(parts, lines[headerEnd:]...)

	return []byte(strings.Join(parts, "\n"))
}

// stripBlockHeaderGeneric strips a block comment header that contains copyright or SPDX lines.
func stripBlockHeaderGeneric(lines []string, i int, style *language.Style) int {
	if i >= len(lines) || strings.TrimSpace(lines[i]) != style.BlockStart {
		// Also check for single-line block comment: /* SPDX-License-Identifier: ... */
		if i < len(lines) {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, style.BlockStart) && strings.HasSuffix(line, style.BlockEnd) {
				if reBlockSPDXLine.MatchString(
					strings.TrimSuffix(strings.TrimPrefix(line, style.BlockStart), style.BlockEnd)) ||
					strings.Contains(line, "SPDX-License-Identifier:") {
					return i + 1
				}
			}
		}
		return i
	}

	for j := i + 1; j < len(lines) && j < i+25; j++ {
		if strings.TrimSpace(lines[j]) == style.BlockEnd {
			hasHeader := false
			for k := i + 1; k < j; k++ {
				if reBlockCopyrightLine.MatchString(lines[k]) ||
					reBlockSPDXLine.MatchString(lines[k]) ||
					reBlockReuseLine.MatchString(lines[k]) ||
					reApacheAnchor.MatchString(lines[k]) ||
					reGPLAnchor.MatchString(lines[k]) {
					hasHeader = true
					break
				}
			}
			if hasHeader {
				return j + 1
			}
			break
		}
	}
	return i
}

// stripLongCommentHeader strips a multi-line comment header identified by an anchor regex.
// It finds consecutive comment lines starting from a copyright line and ending when
// comment lines stop.
func stripLongCommentHeader(lines []string, i int, style *language.Style, anchor *regexp.Regexp) int {
	if i >= len(lines) {
		return i
	}

	// For line-comment styles, look for copyright followed by the anchor
	if !style.IsBlock() {
		prefix := style.LinePrefix
		if !reCopyrightLine.MatchString(lines[i]) {
			return i
		}

		// Scan ahead to see if we find the anchor within the next 20 lines
		foundAnchor := false
		end := i
		for j := i; j < len(lines) && j < i+25; j++ {
			line := lines[j]
			trimmed := strings.TrimSpace(line)

			if trimmed == "" {
				// Blank lines within the header are ok (e.g. between sections)
				if !foundAnchor {
					end = j
					continue
				}
				end = j
				continue
			}

			if !strings.HasPrefix(trimmed, prefix) {
				break
			}
			if anchor.MatchString(line) {
				foundAnchor = true
			}
			end = j + 1
		}

		if foundAnchor {
			return end
		}
		return i
	}

	// Block comment style
	return stripBlockHeaderGeneric(lines, i, style)
}
