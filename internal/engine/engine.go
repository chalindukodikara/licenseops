// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"

	"github.com/licenseops/licenseops/internal/config"
	"github.com/licenseops/licenseops/internal/header"
	"github.com/licenseops/licenseops/internal/language"
	"github.com/licenseops/licenseops/internal/scanner"
)

// Result holds the outcome of a check or fix operation.
type Result struct {
	NonCompliant []string
	Fixed        []string
	Skipped      []string
	Errors       map[string]error
	Warnings     []string
	// Diffs maps file paths to unified diff strings (populated when diff mode is on).
	Diffs map[string]string
}

// fileResult is the per-file outcome sent from workers.
type fileResult struct {
	path         string
	nonCompliant bool
	fixed        bool
	skipped      bool
	err          error
	diff         string // populated in diff mode
}

// Engine orchestrates license header checking and fixing.
type Engine struct {
	cfg      config.Config
	scanner  *scanner.Scanner
	format   header.Format
	warnings []string
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
		cfg:     cfg,
		scanner: scanner.New(root, cfg.Exclude, cfg.ShouldUseGitignore()),
		format:  f,
	}, nil
}

// SetWarnings stores config-level warnings (e.g. deprecated SPDX IDs)
// to be included in every result.
func (e *Engine) SetWarnings(w []string) {
	e.warnings = w
}

// SetInverseExcludes flips the scanner's exclude semantics. When true, only
// files MATCHING exclude patterns are returned by the scanner. Useful for
// `lops remove --excluded-only` to clean up headers in excluded files.
func (e *Engine) SetInverseExcludes(b bool) {
	e.scanner.SetInverseExcludes(b)
}

// Check scans files and reports which ones are non-compliant.
func (e *Engine) Check(verbose bool) (*Result, error) {
	return e.run(false, false, verbose)
}

// Fix scans files and adds/replaces headers on non-compliant ones.
func (e *Engine) Fix(dryRun, verbose bool) (*Result, error) {
	return e.run(!dryRun, false, verbose)
}

// FixWithDiff is like Fix but also computes unified diffs for changed files.
// When diff is true, files are not modified (implies dry-run).
func (e *Engine) FixWithDiff(verbose bool) (*Result, error) {
	return e.run(false, true, verbose)
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

		best := src
		for _, f := range allFormats {
			candidate := f.StripExisting(src, style)
			if len(candidate) < len(best) {
				best = candidate
			}
		}

		if len(best) == len(src) {
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

func (e *Engine) run(fix, diff, verbose bool) (*Result, error) {
	files, err := e.scanner.Scan(e.cfg.Paths)
	if err != nil {
		return nil, fmt.Errorf("scanning files: %w", err)
	}

	workers := runtime.NumCPU()
	// Don't spin up more workers than files
	if workers > len(files) && len(files) > 0 {
		workers = len(files)
	}

	result := &Result{
		Errors:   make(map[string]error),
		Warnings: e.warnings,
	}

	if len(files) == 0 {
		return result, nil
	}

	// If diff mode, prepare the diffs map
	if diff {
		result.Diffs = make(map[string]string)
	}

	fileCh := make(chan string, len(files))
	for _, f := range files {
		fileCh <- f
	}
	close(fileCh)

	resultCh := make(chan fileResult, len(files))

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileCh {
				fr := e.processFile(path, fix, diff)
				resultCh <- fr
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Serialize verbose output
	for fr := range resultCh {
		if fr.err != nil {
			result.Errors[fr.path] = fr.err
			continue
		}
		if fr.skipped {
			result.Skipped = append(result.Skipped, fr.path)
			if verbose {
				fmt.Printf("  skip: %s\n", fr.path)
			}
			continue
		}
		if fr.nonCompliant {
			result.NonCompliant = append(result.NonCompliant, fr.path)
			if fr.fixed {
				result.Fixed = append(result.Fixed, fr.path)
			}
			if fr.diff != "" {
				result.Diffs[fr.path] = fr.diff
			}
		} else if verbose {
			fmt.Printf("  ok: %s\n", fr.path)
		}
	}

	// Sort for deterministic output
	sort.Strings(result.NonCompliant)
	sort.Strings(result.Fixed)
	sort.Strings(result.Skipped)

	return result, nil
}

// processFile handles a single file and returns the result.
// Safe for concurrent use — reads/writes only the given file.
func (e *Engine) processFile(path string, fix, diff bool) fileResult {
	style := language.ForPath(path)
	if style == nil {
		return fileResult{path: path, skipped: true}
	}

	if e.cfg.ShouldSkipGenerated() {
		gen, err := header.IsGenerated(path)
		if err != nil {
			return fileResult{path: path, err: err}
		}
		if gen {
			return fileResult{path: path, skipped: true}
		}
	}

	valid, err := e.format.HasValid(path, style, e.cfg.CopyrightHolder, e.cfg.License)
	if err != nil {
		return fileResult{path: path, err: err}
	}

	if valid {
		return fileResult{path: path}
	}

	fr := fileResult{path: path, nonCompliant: true}

	if diff {
		// Read original content
		original, err := os.ReadFile(path)
		if err != nil {
			return fileResult{path: path, err: err}
		}

		// Generate the fixed content by prepending to a temp copy
		headerText := e.format.Generate(style, e.cfg.Year, e.cfg.CopyrightHolder, e.cfg.License)
		modified := simulatePrepend(original, style, headerText, e.format)
		fr.diff = unifiedDiff(path, original, modified)
		return fr
	}

	if fix {
		headerText := e.format.Generate(style, e.cfg.Year, e.cfg.CopyrightHolder, e.cfg.License)
		if err := header.Prepend(path, style, headerText, e.format); err != nil {
			return fileResult{path: path, err: err}
		}
		fr.fixed = true
	}

	return fr
}

// simulatePrepend computes what Prepend would produce without writing to disk.
func simulatePrepend(src []byte, style *language.Style, headerText string, targetFmt header.Format) []byte {
	// Strip existing headers
	best := src
	for _, f := range header.AllFormats() {
		candidate := f.StripExisting(src, style)
		if len(candidate) < len(best) {
			best = candidate
		}
	}
	candidate := targetFmt.StripExisting(src, style)
	if len(candidate) < len(best) {
		best = candidate
	}

	// Build result the same way Prepend does
	stripped := string(best)
	return []byte(headerText + "\n\n" + stripped)
}

// unifiedDiff produces a simple unified diff between two byte slices.
func unifiedDiff(path string, original, modified []byte) string {
	origLines := splitLines(string(original))
	modLines := splitLines(string(modified))

	// Simple diff: show the full header change (good enough for license headers)
	var out []byte
	out = append(out, []byte(fmt.Sprintf("--- a/%s\n+++ b/%s\n", path, path))...)

	// Find first differing line
	minLen := len(origLines)
	if len(modLines) < minLen {
		minLen = len(modLines)
	}

	firstDiff := 0
	for firstDiff < minLen && origLines[firstDiff] == modLines[firstDiff] {
		firstDiff++
	}

	// Show context around the change
	ctxStart := firstDiff - 3
	if ctxStart < 0 {
		ctxStart = 0
	}

	// Find last differing line from end
	oi, mi := len(origLines)-1, len(modLines)-1
	for oi > firstDiff && mi > firstDiff && origLines[oi] == modLines[mi] {
		oi--
		mi--
	}
	origEnd := oi + 1
	modEnd := mi + 1

	ctxEnd := origEnd + 3
	if ctxEnd > len(origLines) {
		ctxEnd = len(origLines)
	}
	modCtxEnd := modEnd + 3
	if modCtxEnd > len(modLines) {
		modCtxEnd = len(modLines)
	}

	out = append(out, []byte(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
		ctxStart+1, ctxEnd-ctxStart,
		ctxStart+1, modCtxEnd-ctxStart))...)

	// Context before
	for i := ctxStart; i < firstDiff; i++ {
		out = append(out, []byte(" "+origLines[i]+"\n")...)
	}
	// Removed lines
	for i := firstDiff; i < origEnd; i++ {
		out = append(out, []byte("-"+origLines[i]+"\n")...)
	}
	// Added lines
	for i := firstDiff; i < modEnd; i++ {
		out = append(out, []byte("+"+modLines[i]+"\n")...)
	}
	// Context after
	for i := origEnd; i < ctxEnd; i++ {
		out = append(out, []byte(" "+origLines[i]+"\n")...)
	}

	return string(out)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := range len(s) {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
