// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"fmt"
	"strings"

	"github.com/licenseops/licenseops/internal/language"
)

// GPLLongFormat implements the GNU GPL/LGPL/AGPL long-form boilerplate header.
type GPLLongFormat struct{}

func (f *GPLLongFormat) Name() string { return "gpl-long" }

const gplBoilerplate = `Copyright %s %s

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version %s of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.`

const lgplBoilerplate = `Copyright %s %s

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version %s of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.`

const agplBoilerplate = `Copyright %s %s

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version %s of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.`

func gplVersion(license string) string {
	if strings.Contains(license, "2.0") || strings.Contains(license, "2.1") {
		return "2"
	}
	return "3"
}

func gplTemplate(license string) string {
	switch {
	case strings.HasPrefix(license, "LGPL-"):
		return lgplBoilerplate
	case strings.HasPrefix(license, "AGPL-"):
		return agplBoilerplate
	default:
		return gplBoilerplate
	}
}

func (f *GPLLongFormat) Generate(style *language.Style, year, holder, license string) string {
	tmpl := gplTemplate(license)
	ver := gplVersion(license)
	body := fmt.Sprintf(tmpl, year, holder, ver)

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

func (f *GPLLongFormat) HasValid(path string, style *language.Style, holder, license string) (bool, error) {
	lines, err := ReadHeaderLines(path, 25)
	if err != nil {
		return false, err
	}

	if len(lines) < 10 {
		return false, nil
	}

	foundCopyright := false
	foundAnchor := false

	for _, line := range lines {
		if !foundCopyright && strings.Contains(line, "Copyright") && strings.Contains(line, holder) {
			foundCopyright = true
		}
		if reGPLAnchor.MatchString(line) {
			foundAnchor = true
		}
	}

	return foundCopyright && foundAnchor, nil
}

func (f *GPLLongFormat) StripExisting(src []byte, style *language.Style) []byte {
	lines := strings.Split(string(src), "\n")
	i := skipPreamble(lines)
	headerStart := i

	if style.IsBlock() {
		i = stripBlockHeaderGeneric(lines, i, style)
	} else {
		i = stripLongCommentHeader(lines, i, style, reGPLAnchor)
	}

	if i == headerStart {
		return src
	}
	return reconstructAfterStrip(lines, src, i)
}
