package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the licenser tool.
type Config struct {
	License         string   `yaml:"license"`
	CopyrightHolder string   `yaml:"copyright-holder"`
	Year            string   `yaml:"year"`
	HeaderTemplate  string   `yaml:"header-template"`
	Paths           []string `yaml:"paths"`
	Exclude         []string `yaml:"exclude"`
	SkipGenerated   *bool    `yaml:"skip-generated"`
	Gitignore       *bool    `yaml:"gitignore"`
}

// Defaults returns a Config with default values applied.
func Defaults() Config {
	t := true
	return Config{
		Year:          fmt.Sprint(time.Now().Year()),
		Paths:         []string{"."},
		SkipGenerated: &t,
		Gitignore:     &t,
		Exclude: []string{
			"vendor/**",
			"node_modules/**",
			".git/**",
			"third_party/**",
			".licenser.yaml",
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
	if fileCfg.HeaderTemplate != "" {
		cfg.HeaderTemplate = fileCfg.HeaderTemplate
	}
	if len(fileCfg.Paths) > 0 {
		cfg.Paths = fileCfg.Paths
	}
	if len(fileCfg.Exclude) > 0 {
		cfg.Exclude = append(cfg.Exclude, fileCfg.Exclude...)
	}
	if fileCfg.SkipGenerated != nil {
		cfg.SkipGenerated = fileCfg.SkipGenerated
	}
	if fileCfg.Gitignore != nil {
		cfg.Gitignore = fileCfg.Gitignore
	}

	return cfg, nil
}

// Validate checks that required fields are set.
func (c *Config) Validate() error {
	if c.License == "" {
		return fmt.Errorf("license is required (set via config or --license flag)")
	}
	if c.CopyrightHolder == "" {
		return fmt.Errorf("copyright-holder is required (set via config or --owner flag)")
	}
	if !IsValidLicense(c.License) {
		return fmt.Errorf("unknown SPDX license identifier: %q (use a valid SPDX ID)", c.License)
	}
	return nil
}

// ShouldSkipGenerated returns whether generated files should be skipped.
func (c *Config) ShouldSkipGenerated() bool {
	return c.SkipGenerated == nil || *c.SkipGenerated
}

// ShouldUseGitignore returns whether .gitignore patterns should be respected.
func (c *Config) ShouldUseGitignore() bool {
	return c.Gitignore == nil || *c.Gitignore
}

// validLicenses is the set of recognized SPDX identifiers.
var validLicenses = map[string]bool{
	"Apache-2.0":      true,
	"MIT":             true,
	"GPL-2.0-only":    true,
	"GPL-2.0-or-later": true,
	"GPL-3.0-only":    true,
	"GPL-3.0-or-later": true,
	"LGPL-2.1-only":   true,
	"LGPL-2.1-or-later": true,
	"LGPL-3.0-only":   true,
	"LGPL-3.0-or-later": true,
	"BSD-2-Clause":    true,
	"BSD-3-Clause":    true,
	"MPL-2.0":         true,
	"AGPL-3.0-only":   true,
	"AGPL-3.0-or-later": true,
	"ISC":             true,
	"Unlicense":       true,
	"BSL-1.0":         true,
	"0BSD":            true,
	"CC0-1.0":         true,
}

// IsValidLicense checks if the given string is a recognized SPDX identifier.
func IsValidLicense(id string) bool {
	return validLicenses[id]
}
