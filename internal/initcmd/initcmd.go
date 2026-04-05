// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package initcmd

import (
	"fmt"
	"os"
	"strings"
)

// ProjectType represents a detected project ecosystem.
type ProjectType string

const (
	ProjectGo      ProjectType = "go"
	ProjectNode    ProjectType = "node"
	ProjectPython  ProjectType = "python"
	ProjectRust    ProjectType = "rust"
	ProjectGeneric ProjectType = "generic"
)

// DetectProjectType inspects the directory for common project files.
func DetectProjectType(dir string) ProjectType {
	checks := []struct {
		file string
		typ  ProjectType
	}{
		{"go.mod", ProjectGo},
		{"package.json", ProjectNode},
		{"pyproject.toml", ProjectPython},
		{"setup.py", ProjectPython},
		{"requirements.txt", ProjectPython},
		{"Cargo.toml", ProjectRust},
	}
	for _, c := range checks {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", dir, c.file)); err == nil {
			return c.typ
		}
	}
	return ProjectGeneric
}

// SuggestExcludes returns sensible exclude patterns for the project type.
func SuggestExcludes(pt ProjectType) []string {
	switch pt {
	case ProjectGo:
		return []string{"**/*.pb.go", "**/testdata/**"}
	case ProjectNode:
		return []string{"dist/**", "build/**", "coverage/**"}
	case ProjectPython:
		return []string{"venv/**", ".venv/**", "**/__pycache__/**", "*.egg-info/**", "dist/**"}
	case ProjectRust:
		return []string{"target/**"}
	default:
		return nil
	}
}

// GenerateConfig produces the YAML config content.
func GenerateConfig(license, holder, format string, extraExcludes []string) string {
	var sb strings.Builder

	sb.WriteString("# LicenseOps configuration\n")
	sb.WriteString("# See: https://github.com/chalindukodikara/licenseops\n\n")

	fmt.Fprintf(&sb, "license: %s\n", license)

	if holder != "" {
		fmt.Fprintf(&sb, "copyright-holder: %q\n", holder)
	}

	if format != "" && format != "spdx" {
		fmt.Fprintf(&sb, "format: %s\n", format)
	}

	if len(extraExcludes) > 0 {
		sb.WriteString("\nexclude:\n")
		for _, e := range extraExcludes {
			fmt.Fprintf(&sb, "  - %q\n", e)
		}
	}

	return sb.String()
}
