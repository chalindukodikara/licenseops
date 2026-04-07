// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/licenseops/licenseops/internal/spdx"
	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the licenser tool.
type Config struct {
	License         string   `yaml:"license"`
	CopyrightHolder string   `yaml:"copyright-holder"`
	Year            string   `yaml:"year"`
	Format          string   `yaml:"format"`
	HeaderTemplate  string   `yaml:"header-template"`
	Paths           []string `yaml:"paths"`
	Exclude         []string `yaml:"exclude"`
	SkipGenerated   *bool    `yaml:"skip-generated"`
	Gitignore       *bool    `yaml:"gitignore"`

	// UserExcludes holds ONLY the exclude patterns the user explicitly
	// declared in their config file (no built-in defaults, no auto-added
	// config-file path). It is populated by Load() and used by features
	// like `lops remove --excluded-only` that need to operate on the
	// user's explicit excludes rather than the merged set.
	UserExcludes []string `yaml:"-"`
}

// Defaults returns a Config with default values applied.
func Defaults() Config {
	t := true
	return Config{
		Year:          fmt.Sprint(time.Now().Year()),
		Format:        "spdx",
		Paths:         []string{"."},
		SkipGenerated: &t,
		Gitignore:     &t,
		Exclude: []string{
			"vendor/**",
			"node_modules/**",
			".git/**",
			"third_party/**",
			".licenseops.yaml",
		},
	}
}

// Load reads a YAML config file and returns the parsed Config.
// If the file does not exist, it returns defaults with no error.
func Load(path string) (Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return cfg, fmt.Errorf("parsing config %s: %w", path, err)
	}

	// Merge file config into defaults
	if fileCfg.License != "" {
		cfg.License = fileCfg.License
	}
	if fileCfg.CopyrightHolder != "" {
		cfg.CopyrightHolder = fileCfg.CopyrightHolder
	}
	if fileCfg.Year != "" {
		cfg.Year = fileCfg.Year
	}
	if fileCfg.Format != "" {
		cfg.Format = fileCfg.Format
	}
	if fileCfg.HeaderTemplate != "" {
		cfg.HeaderTemplate = fileCfg.HeaderTemplate
	}
	if len(fileCfg.Paths) > 0 {
		cfg.Paths = fileCfg.Paths
	}
	if len(fileCfg.Exclude) > 0 {
		cfg.Exclude = append(cfg.Exclude, fileCfg.Exclude...)
		cfg.UserExcludes = append(cfg.UserExcludes, fileCfg.Exclude...)
	}
	if fileCfg.SkipGenerated != nil {
		cfg.SkipGenerated = fileCfg.SkipGenerated
	}
	if fileCfg.Gitignore != nil {
		cfg.Gitignore = fileCfg.Gitignore
	}
	return cfg, nil
}

// Validate checks that required fields are set and consistent.
func (c *Config) Validate() (warnings []string, err error) {
	if c.License == "" {
		return nil, fmt.Errorf("license is required (set via config or --license flag)")
	}

	// Validate license (single ID or expression)
	w, err := spdx.ValidateLicense(c.License)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, w...)

	// Format-specific validation
	switch c.Format {
	case "spdx", "":
		// copyright-holder is optional for spdx (1-line mode without it)
	case "reuse":
		if c.CopyrightHolder == "" {
			return warnings, fmt.Errorf("copyright-holder is required for 'reuse' format")
		}
	case "apache-long":
		if c.CopyrightHolder == "" {
			return warnings, fmt.Errorf("copyright-holder is required for 'apache-long' format")
		}
		if c.License != "Apache-2.0" {
			return warnings, fmt.Errorf("'apache-long' format requires license: Apache-2.0 (got %q)", c.License)
		}
	case "gpl-long":
		if c.CopyrightHolder == "" {
			return warnings, fmt.Errorf("copyright-holder is required for 'gpl-long' format")
		}
		if !spdx.IsGPLFamily(c.License) {
			return warnings, fmt.Errorf("'gpl-long' format requires a GPL/LGPL/AGPL license (got %q)\n  valid: GPL-2.0-only, GPL-2.0-or-later, GPL-3.0-only, GPL-3.0-or-later,\n         LGPL-2.1-only, LGPL-2.1-or-later, LGPL-3.0-only, LGPL-3.0-or-later,\n         AGPL-3.0-only, AGPL-3.0-or-later", c.License)
		}
	case "custom":
		if c.HeaderTemplate == "" {
			return warnings, fmt.Errorf("header-template is required when format is 'custom'")
		}
		if _, err := os.Stat(c.HeaderTemplate); err != nil {
			return warnings, fmt.Errorf("header template file not found: %s", c.HeaderTemplate)
		}
	default:
		return warnings, fmt.Errorf("unknown format: %q (valid: spdx, reuse, apache-long, gpl-long, custom)", c.Format)
	}

	return warnings, nil
}

// ShouldSkipGenerated returns whether generated files should be skipped.
func (c *Config) ShouldSkipGenerated() bool {
	return c.SkipGenerated == nil || *c.SkipGenerated
}

// ShouldUseGitignore returns whether .gitignore patterns should be respected.
func (c *Config) ShouldUseGitignore() bool {
	return c.Gitignore == nil || *c.Gitignore
}
