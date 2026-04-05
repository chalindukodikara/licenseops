// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"fmt"
	"os"

	"github.com/chalindukodikara/licenseops/internal/config"
	"github.com/chalindukodikara/licenseops/internal/header"
	"github.com/chalindukodikara/licenseops/internal/language"
	"github.com/chalindukodikara/licenseops/internal/scanner"
)

// Result holds the outcome of a check or fix operation.
type Result struct {
	NonCompliant []string
	Fixed        []string
	Skipped      []string
	Errors       map[string]error
	Warnings     []string
}

// Engine orchestrates license header checking and fixing.
type Engine struct {
	cfg      config.Config
	scanner  *scanner.Scanner
	format   header.Format
	warnings []string // SPDX deprecation warnings from config validation
}

// New creates an Engine from the given config.
func New(cfg config.Config) (*Engine, error) {
	root := "."
	if len(cfg.Paths) > 0 {
		root = cfg.Paths[0]
	}

	f, err := header.FormatByName(cfg.Format)
	if err != nil {
		return nil, err
	}

	// Load custom template if needed
	if cfg.Format == "custom" {
		cf := f.(*header.CustomFormat)
		if err := cf.LoadTemplate(cfg.HeaderTemplate); err != nil {
			return nil, err
		}
	}

	return &Engine{
		cfg:      cfg,
		scanner:  scanner.New(root, cfg.Exclude, cfg.ShouldUseGitignore()),
		format:   f,
	}, nil
}

// SetWarnings stores config-level warnings (e.g. deprecated SPDX IDs)
// to be included in every result.
func (e *Engine) SetWarnings(w []string) {
	e.warnings = w
}

// Check scans files and reports which ones are non-compliant.
func (e *Engine) Check(verbose bool) (*Result, error) {
	return e.run(false, verbose)
}

// Fix scans files and adds/replaces headers on non-compliant ones.
func (e *Engine) Fix(dryRun, verbose bool) (*Result, error) {
	return e.run(!dryRun, verbose)
}

// Remove strips license headers from all scanned files.
func (e *Engine) Remove(dryRun, verbose bool) (*Result, error) {
	files, err := e.scanner.Scan(e.cfg.Paths)
	if err != nil {
		return nil, fmt.Errorf("scanning files: %w", err)
	}

	result := &Result{
		Errors:   make(map[string]error),
		Warnings: e.warnings,
	}

	allFormats := header.AllFormats()
	// Also include the configured format (may be custom)
	allFormats = append(allFormats, e.format)

	for _, path := range files {
		style := language.ForPath(path)
		if style == nil {
			result.Skipped = append(result.Skipped, path)
			continue
		}

		src, err := os.ReadFile(path)
		if err != nil {
			result.Errors[path] = err
			continue
		}

		// Try stripping with all formats and pick the result that removed the most
		best := src
		for _, f := range allFormats {
			candidate := f.StripExisting(src, style)
			if len(candidate) < len(best) {
				best = candidate
			}
		}

		if len(best) == len(src) {
			// No header found
			if verbose {
				fmt.Printf("  skip (no header): %s\n", path)
			}
			continue
		}

		result.NonCompliant = append(result.NonCompliant, path)

		if !dryRun {
			info, err := os.Stat(path)
			if err != nil {
				result.Errors[path] = err
				continue
			}
			if err := os.WriteFile(path, best, info.Mode()); err != nil {
				result.Errors[path] = err
				continue
			}
			result.Fixed = append(result.Fixed, path)
			if verbose {
				fmt.Printf("  removed: %s\n", path)
			}
		}
	}

	return result, nil
}

func (e *Engine) run(fix, verbose bool) (*Result, error) {
	files, err := e.scanner.Scan(e.cfg.Paths)
	if err != nil {
		return nil, fmt.Errorf("scanning files: %w", err)
	}

	result := &Result{
		Errors:   make(map[string]error),
		Warnings: e.warnings,
	}

	for _, path := range files {
		style := language.ForPath(path)
		if style == nil {
			result.Skipped = append(result.Skipped, path)
			continue
		}

		if e.cfg.ShouldSkipGenerated() {
			gen, err := header.IsGenerated(path)
			if err != nil {
				result.Errors[path] = err
				continue
			}
			if gen {
				result.Skipped = append(result.Skipped, path)
				if verbose {
					fmt.Printf("  skip (generated): %s\n", path)
				}
				continue
			}
		}

		valid, err := e.format.HasValid(path, style, e.cfg.CopyrightHolder, e.cfg.License)
		if err != nil {
			result.Errors[path] = err
			continue
		}

		if valid {
			if verbose {
				fmt.Printf("  ok: %s\n", path)
			}
			continue
		}

		result.NonCompliant = append(result.NonCompliant, path)

		if fix {
			headerText := e.format.Generate(style, e.cfg.Year, e.cfg.CopyrightHolder, e.cfg.License)
			if err := header.Prepend(path, style, headerText, e.format); err != nil {
				result.Errors[path] = err
				continue
			}
			result.Fixed = append(result.Fixed, path)
		}
	}

	return result, nil
}
