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
	NonCompliant []string
	Fixed        []string
	Skipped      []string
	Errors       map[string]error
}

// Engine orchestrates license header checking and fixing.
type Engine struct {
	cfg     config.Config
	scanner *scanner.Scanner
	format  header.Format
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
