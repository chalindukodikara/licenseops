// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"fmt"
	"strings"

	"github.com/licenseops/licenseops/internal/language"
)

// ApacheLongFormat implements the Apache License 2.0 long-form boilerplate header.
type ApacheLongFormat struct{}

func (f *ApacheLongFormat) Name() string { return "apache-long" }

const apacheBoilerplate = `Copyright %s %s

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.`

func (f *ApacheLongFormat) Generate(style *language.Style, year, holder, license string) string {
	body := fmt.Sprintf(apacheBoilerplate, year, holder)

	if style.IsBlock() {
		var sb strings.Builder
		sb.WriteString(style.BlockStart)
		sb.WriteString("\n")
		for _, line := range strings.Split(body, "\n") {
			if line == "" {
				sb.WriteString("\n")
			} else {
				sb.WriteString(" ")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		}
		sb.WriteString(style.BlockEnd)
		return sb.String()
	}

	var sb strings.Builder
	for i, line := range strings.Split(body, "\n") {
		if i > 0 {
			sb.WriteString("\n")
		}
		if line == "" {
			sb.WriteString(style.LinePrefix)
		} else {
			sb.WriteString(style.LinePrefix)
			sb.WriteString(" ")
			sb.WriteString(line)
		}
	}
	return sb.String()
}

func (f *ApacheLongFormat) HasValid(path string, style *language.Style, holder, license string) (bool, error) {
	lines, err := ReadHeaderLines(path, 25)
	if err != nil {
		return false, err
	}

	if len(lines) < 10 {
		return false, nil
	}

	// Look for copyright line with the correct holder
	foundCopyright := false
	foundAnchor := false

	for _, line := range lines {
		if !foundCopyright && strings.Contains(line, "Copyright") && strings.Contains(line, holder) {
			foundCopyright = true
		}
		if reApacheAnchor.MatchString(line) {
			foundAnchor = true
		}
	}

	return foundCopyright && foundAnchor, nil
}

func (f *ApacheLongFormat) StripExisting(src []byte, style *language.Style) []byte {
	lines := strings.Split(string(src), "\n")
	i := skipPreamble(lines)
	headerStart := i

	if style.IsBlock() {
		i = stripBlockHeaderGeneric(lines, i, style)
	} else {
		i = stripLongCommentHeader(lines, i, style, reApacheAnchor)
	}

	if i == headerStart {
		return src
	}
	return reconstructAfterStrip(lines, src, i)
}
