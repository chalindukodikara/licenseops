// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	ignore "github.com/sabhiram/go-gitignore"

	"github.com/licenseops/licenseops/internal/language"
)

// Scanner walks directories and yields files that should be processed.
type Scanner struct {
	excludes   []string
	gitignore  *ignore.GitIgnore
	skipUnsupported bool
}

// New creates a Scanner with the given exclude patterns.
// If useGitignore is true, it loads .gitignore from root.
func New(root string, excludes []string, useGitignore bool) *Scanner {
	s := &Scanner{
		excludes:        excludes,
		skipUnsupported: true,
	}

	if useGitignore {
		giPath := filepath.Join(root, ".gitignore")
		if gi, err := ignore.CompileIgnoreFile(giPath); err == nil {
			s.gitignore = gi
		}
	}

	return s
}

// Scan walks the given paths and returns all files that should be processed.
// Paths can be files or directories.
func (s *Scanner) Scan(paths []string) ([]string, error) {
	var files []string

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			if s.shouldProcess(p) {
				files = append(files, p)
			}
			continue
		}

		err = filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Get relative path for pattern matching
			rel, relErr := filepath.Rel(p, path)
			if relErr != nil {
				rel = path
			}

			if d.IsDir() {
				if s.isExcluded(rel) || s.isExcluded(path) {
					return filepath.SkipDir
				}
				return nil
			}

			if s.isExcluded(rel) || s.isExcluded(path) {
				return nil
			}

			if s.gitignore != nil && s.gitignore.MatchesPath(rel) {
				return nil
			}

			if s.skipUnsupported && !language.Supported(path) {
				return nil
			}

			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// isExcluded checks if a path matches any exclude pattern.
func (s *Scanner) isExcluded(path string) bool {
	// Normalize separators for matching
	normalized := filepath.ToSlash(path)

	for _, pattern := range s.excludes {
		// Try matching the pattern as-is
		if matched, _ := doublestar.Match(pattern, normalized); matched {
			return true
		}
		// Also try matching against just the filename for simple patterns
		base := filepath.Base(path)
		if !strings.Contains(pattern, "/") {
			if matched, _ := doublestar.Match(pattern, base); matched {
				return true
			}
		}
	}
	return false
}

// shouldProcess checks if a single file should be processed.
func (s *Scanner) shouldProcess(path string) bool {
	if s.isExcluded(path) {
		return false
	}
	if s.skipUnsupported && !language.Supported(path) {
		return false
	}
	return true
}
