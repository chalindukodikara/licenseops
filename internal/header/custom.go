// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/chalindukodikara/licenseops/internal/language"
)

// CustomFormat implements a user-defined header template format.
type CustomFormat struct {
	TemplatePath string
	tmpl         *template.Template
}

// TemplateData holds the variables available in custom templates.
type TemplateData struct {
	Year       string
	Holder     string
	License    string
	Comment    string
	BlockStart string
	BlockEnd   string
}

func (f *CustomFormat) Name() string { return "custom" }

// LoadTemplate parses the template file. Must be called before Generate/HasValid.
func (f *CustomFormat) LoadTemplate(path string) error {
	f.TemplatePath = path
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading template %s: %w", path, err)
	}
	t, err := template.New("header").Parse(string(data))
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", path, err)
	}
	f.tmpl = t
	return nil
}

func (f *CustomFormat) Generate(style *language.Style, year, holder, license string) string {
	if f.tmpl == nil {
		return ""
	}

	data := TemplateData{
		Year:    year,
		Holder:  holder,
		License: license,
	}
	if style.IsBlock() {
		data.BlockStart = style.BlockStart
		data.BlockEnd = style.BlockEnd
	} else {
		data.Comment = style.LinePrefix
	}

	var buf bytes.Buffer
	if err := f.tmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return strings.TrimRight(buf.String(), "\n")
}

func (f *CustomFormat) HasValid(path string, style *language.Style, holder, license string) (bool, error) {
	if f.tmpl == nil {
		return false, fmt.Errorf("custom template not loaded")
	}

	lines, err := ReadHeaderLines(path, 30)
	if err != nil {
		return false, err
	}
	if len(lines) == 0 {
		return false, nil
	}

	// Generate expected header with a year wildcard
	expected := f.Generate(style, "YEAR_PLACEHOLDER", holder, license)
	expectedLines := strings.Split(expected, "\n")

	if len(lines) < len(expectedLines)+1 {
		return false, nil
	}

	for i, eLine := range expectedLines {
		// Replace the year placeholder with a regex for any year
		pattern := regexp.QuoteMeta(eLine)
		pattern = strings.ReplaceAll(pattern, "YEAR_PLACEHOLDER", `\d{4}`)
		matched, err := regexp.MatchString("^"+pattern+"$", lines[i])
		if err != nil || !matched {
			return false, nil
		}
	}

	// Line after header should be blank
	nextIdx := len(expectedLines)
	if nextIdx < len(lines) && strings.TrimSpace(lines[nextIdx]) != "" {
		return false, nil
	}

	return true, nil
}

func (f *CustomFormat) StripExisting(src []byte, style *language.Style) []byte {
	if f.tmpl == nil {
		return src
	}

	lines := strings.Split(string(src), "\n")
	i := skipPreamble(lines)

	// Try to match the template against the start of the file
	expected := f.Generate(style, "YEAR_PLACEHOLDER", "HOLDER_PLACEHOLDER", "LICENSE_PLACEHOLDER")
	expectedLines := strings.Split(expected, "\n")

	if len(lines)-i < len(expectedLines) {
		return src
	}

	// Check if the beginning matches a header pattern
	headerEnd := i
	matched := true
	for j, eLine := range expectedLines {
		if i+j >= len(lines) {
			matched = false
			break
		}
		pattern := regexp.QuoteMeta(eLine)
		pattern = strings.ReplaceAll(pattern, "YEAR_PLACEHOLDER", `\d{4}`)
		pattern = strings.ReplaceAll(pattern, "HOLDER_PLACEHOLDER", `.+`)
		pattern = strings.ReplaceAll(pattern, "LICENSE_PLACEHOLDER", `.+`)
		ok, _ := regexp.MatchString("^"+pattern+"$", lines[i+j])
		if !ok {
			matched = false
			break
		}
		headerEnd = i + j + 1
	}

	if !matched || headerEnd == i {
		return src
	}

	return reconstructAfterStrip(lines, src, headerEnd)
}
