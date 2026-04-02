package header

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/chalindu/licenser/internal/language"
)

// reCopyrightLine matches a copyright line in any single-line comment style.
var reCopyrightLine = regexp.MustCompile(`^(//|#|--)\s*Copyright\s+\d{4}\s+.+$`)

// reSPDXLine matches an SPDX-License-Identifier line in any single-line comment style.
var reSPDXLine = regexp.MustCompile(`^(//|#|--)\s*SPDX-License-Identifier:\s+.+$`)

// reBlockCopyrightLine matches a copyright line inside a block comment.
var reBlockCopyrightLine = regexp.MustCompile(`^\s*\*?\s*Copyright\s+\d{4}\s+.+$`)

// reBlockSPDXLine matches an SPDX line inside a block comment.
var reBlockSPDXLine = regexp.MustCompile(`^\s*\*?\s*SPDX-License-Identifier:\s+.+$`)

// reGeneratedMarker detects generated files.
var reGeneratedMarker = regexp.MustCompile(`(?i)(code generated.*do not edit|@generated)`)

// reShebang matches a shebang line.
var reShebang = regexp.MustCompile(`^#!.+`)

// rePythonEncoding matches Python encoding declarations.
var rePythonEncoding = regexp.MustCompile(`^#.*(-\*-\s*coding[:=]|coding[:=])\s*\S+`)

// Generate produces the license header text for the given style.
func Generate(style *language.Style, year, holder, license string) string {
	copyright := fmt.Sprintf("Copyright %s %s", year, holder)
	spdx := fmt.Sprintf("SPDX-License-Identifier: %s", license)

	if style.IsBlock() {
		return fmt.Sprintf("%s\n %s\n %s\n%s",
			style.BlockStart, copyright, spdx, style.BlockEnd)
	}
	return fmt.Sprintf("%s %s\n%s %s",
		style.LinePrefix, copyright, style.LinePrefix, spdx)
}

// IsGenerated checks if a file contains generated-code markers.
// Only scans the first 30 lines.
func IsGenerated(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 30 && scanner.Scan(); i++ {
		if reGeneratedMarker.MatchString(scanner.Text()) {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// HasValid checks if a file has a valid license header matching the expected
// holder and license. It is fuzzy on year — any year is accepted.
func HasValid(path string, style *language.Style, holder, license string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string

	// Collect first meaningful lines, skipping leading blanks and shebangs
	for scanner.Scan() {
		line := scanner.Text()

		if len(lines) == 0 {
			// Skip leading blank lines
			if strings.TrimSpace(line) == "" {
				continue
			}
			// Skip shebang
			if reShebang.MatchString(line) {
				continue
			}
			// Skip Python encoding declarations
			if rePythonEncoding.MatchString(line) {
				continue
			}
		}

		lines = append(lines, line)
		if len(lines) >= 4 {
			break
		}
	}

	if style.IsBlock() {
		return hasValidBlockHeader(lines, style, holder, license), nil
	}
	return hasValidLineHeader(lines, style, holder, license), nil
}

func hasValidLineHeader(lines []string, style *language.Style, holder, license string) bool {
	if len(lines) < 3 {
		return false
	}

	prefix := style.LinePrefix

	// Line 0: copyright
	copyrightPat := fmt.Sprintf(`^%s Copyright \d{4} %s$`,
		regexp.QuoteMeta(prefix), regexp.QuoteMeta(holder))
	if matched, _ := regexp.MatchString(copyrightPat, lines[0]); !matched {
		return false
	}

	// Line 1: SPDX
	spdxPat := fmt.Sprintf(`^%s SPDX-License-Identifier: %s$`,
		regexp.QuoteMeta(prefix), regexp.QuoteMeta(license))
	if matched, _ := regexp.MatchString(spdxPat, lines[1]); !matched {
		return false
	}

	// Line 2: should be blank
	if strings.TrimSpace(lines[2]) != "" {
		return false
	}

	return true
}

func hasValidBlockHeader(lines []string, style *language.Style, holder, license string) bool {
	if len(lines) < 4 {
		return false
	}

	// Line 0: block start
	if strings.TrimSpace(lines[0]) != style.BlockStart {
		return false
	}

	// Line 1: copyright
	copyrightPat := fmt.Sprintf(`^\s*Copyright \d{4} %s$`, regexp.QuoteMeta(holder))
	if matched, _ := regexp.MatchString(copyrightPat, strings.TrimPrefix(lines[1], " ")); !matched {
		return false
	}

	// Line 2: SPDX
	spdxPat := fmt.Sprintf(`^\s*SPDX-License-Identifier: %s$`, regexp.QuoteMeta(license))
	if matched, _ := regexp.MatchString(spdxPat, strings.TrimPrefix(lines[2], " ")); !matched {
		return false
	}

	// Line 3: block end
	if strings.TrimSpace(lines[3]) != style.BlockEnd {
		return false
	}

	return true
}

// StripExisting removes any existing copyright/SPDX header from the beginning
// of the file content so a correct header can be prepended without duplication.
func StripExisting(src []byte, style *language.Style) []byte {
	lines := strings.Split(string(src), "\n")

	i := 0
	// Skip leading blank lines
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	// Preserve shebang
	shebangIdx := -1
	if i < len(lines) && reShebang.MatchString(lines[i]) {
		shebangIdx = i
		i++
		// Skip blank line after shebang
		if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
	}

	// Preserve Python encoding
	encodingLines := []string{}
	for i < len(lines) && rePythonEncoding.MatchString(lines[i]) {
		encodingLines = append(encodingLines, lines[i])
		i++
	}
	// Skip blank line after encoding
	if len(encodingLines) > 0 && i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	headerStart := i

	if style.IsBlock() {
		i = stripBlockHeader(lines, i, style)
	} else {
		i = stripLineHeader(lines, i, style)
	}

	if i == headerStart {
		return src
	}

	// Skip one trailing blank line after the header
	if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	// Reconstruct
	var parts []string
	if shebangIdx >= 0 {
		parts = append(parts, lines[shebangIdx])
	}
	parts = append(parts, encodingLines...)
	if (shebangIdx >= 0 || len(encodingLines) > 0) && i < len(lines) {
		// Ensure blank line separator
		if len(parts) > 0 && strings.TrimSpace(lines[i]) != "" {
			// We'll let the prepend logic handle the blank line
		}
	}
	parts = append(parts, lines[i:]...)

	return []byte(strings.Join(parts, "\n"))
}

func stripLineHeader(lines []string, i int, style *language.Style) int {
	// Check for copyright line
	if i < len(lines) && reCopyrightLine.MatchString(lines[i]) {
		// Copyright found, check for SPDX on next line
		if i+1 < len(lines) && reSPDXLine.MatchString(lines[i+1]) {
			return i + 2 // Both lines consumed
		}
		return i + 1 // Copyright only
	}

	// Check for SPDX-only line
	if i < len(lines) && reSPDXLine.MatchString(lines[i]) {
		return i + 1
	}

	return i
}

func stripBlockHeader(lines []string, i int, style *language.Style) int {
	if i >= len(lines) || strings.TrimSpace(lines[i]) != style.BlockStart {
		return i
	}

	// Look for block end within next few lines
	for j := i + 1; j < len(lines) && j < i+10; j++ {
		if strings.TrimSpace(lines[j]) == style.BlockEnd {
			// Check if this block contains copyright/SPDX content
			hasHeader := false
			for k := i + 1; k < j; k++ {
				if reBlockCopyrightLine.MatchString(lines[k]) || reBlockSPDXLine.MatchString(lines[k]) {
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

// Prepend adds the header to the file, handling shebangs and encoding declarations.
func Prepend(path string, style *language.Style, headerText string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Strip any existing header first
	stripped := StripExisting(src, style)

	lines := strings.Split(string(stripped), "\n")
	i := 0

	// Collect prefix lines that must stay at the top
	var prefixLines []string

	// Shebang
	if i < len(lines) && reShebang.MatchString(lines[i]) {
		prefixLines = append(prefixLines, lines[i])
		i++
		// Skip blank line after shebang
		if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
	}

	// Python encoding declarations
	for i < len(lines) && rePythonEncoding.MatchString(lines[i]) {
		prefixLines = append(prefixLines, lines[i])
		i++
	}
	// Skip blank line after encoding
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
