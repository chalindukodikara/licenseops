package engine

import (
	"fmt"

	"github.com/chalindu/licenser/internal/config"
	"github.com/chalindu/licenser/internal/header"
	"github.com/chalindu/licenser/internal/language"
	"github.com/chalindu/licenser/internal/scanner"
)

// Result holds the outcome of a check or fix operation.
type Result struct {
	// NonCompliant is the list of files that are missing or have incorrect headers.
	NonCompliant []string
	// Fixed is the list of files that were modified (fix mode only).
	Fixed []string
	// Skipped is the list of files that were skipped (generated, unsupported, etc.).
	Skipped []string
	// Errors maps file paths to errors encountered during processing.
	Errors map[string]error
}

// Engine orchestrates license header checking and fixing.
type Engine struct {
	cfg     config.Config
	scanner *scanner.Scanner
}

// New creates an Engine from the given config.
func New(cfg config.Config) *Engine {
	// Determine root for gitignore — use first path or "."
	root := "."
	if len(cfg.Paths) > 0 {
		root = cfg.Paths[0]
	}

	return &Engine{
		cfg:     cfg,
		scanner: scanner.New(root, cfg.Exclude, cfg.ShouldUseGitignore()),
	}
}

// Check scans files and reports which ones are non-compliant.
func (e *Engine) Check(verbose bool) (*Result, error) {
	return e.run(false, verbose)
}

// Fix scans files and adds/replaces headers on non-compliant ones.
func (e *Engine) Fix(dryRun, verbose bool) (*Result, error) {
	return e.run(!dryRun, verbose)
}

func (e *Engine) run(fix, verbose bool) (*Result, error) {
	files, err := e.scanner.Scan(e.cfg.Paths)
	if err != nil {
		return nil, fmt.Errorf("scanning files: %w", err)
	}

	result := &Result{
		Errors: make(map[string]error),
	}

	for _, path := range files {
		style := language.ForPath(path)
		if style == nil {
			result.Skipped = append(result.Skipped, path)
			continue
		}

		// Check for generated files
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

		valid, err := header.HasValid(path, style, e.cfg.CopyrightHolder, e.cfg.License)
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
			headerText := header.Generate(style, e.cfg.Year, e.cfg.CopyrightHolder, e.cfg.License)
			if err := header.Prepend(path, style, headerText); err != nil {
				result.Errors[path] = err
				continue
			}
			result.Fixed = append(result.Fixed, path)
		}
	}

	return result, nil
}
