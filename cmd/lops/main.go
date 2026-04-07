// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/licenseops/licenseops/internal/config"
	"github.com/licenseops/licenseops/internal/engine"
	"github.com/licenseops/licenseops/internal/initcmd"
	"github.com/licenseops/licenseops/internal/output"
)

var version = "dev"

var (
	flagConfig       string
	flagLicense      string
	flagOwner        string
	flagYear         string
	flagFormat       string
	flagVerbose      bool
	flagDryRun       bool
	flagOutputFormat string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "lops",
		Short: "Enforce SPDX license headers in source files",
		Long: `LicenseOps is a CLI tool to check and fix license headers in source files.

Supported formats:
  spdx         2-line (Copyright + SPDX) or 1-line (SPDX only)
  reuse        FSFE REUSE spec (SPDX-FileCopyrightText + SPDX)
  apache-long  Apache License 2.0 boilerplate
  gpl-long     GNU GPL/LGPL/AGPL boilerplate
  custom       User-defined template

Use 'lops check' to validate compliance and 'lops fix' to add or
replace headers automatically. Configuration is read from .licenseops.yaml
or provided via CLI flags.`,
		Version: version,
	}

	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().StringVarP(&flagLicense, "license", "l", "", "SPDX license identifier or expression")
	rootCmd.PersistentFlags().StringVarP(&flagOwner, "owner", "o", "", "copyright holder")
	rootCmd.PersistentFlags().StringVarP(&flagFormat, "format", "f", "", "header format: spdx, reuse, apache-long, gpl-long, custom")
	rootCmd.PersistentFlags().StringVarP(&flagYear, "year", "y", "", "copyright year")
	rootCmd.PersistentFlags().StringVarP(&flagConfig, "config", "c", ".licenseops.yaml", "path to config file")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "print every file checked")
	rootCmd.PersistentFlags().StringVar(&flagOutputFormat, "output", "text", "output format: text, json, sarif")

	rootCmd.AddCommand(checkCmd())
	rootCmd.AddCommand(fixCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(removeCmd())
	rootCmd.AddCommand(validateCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

func loadConfig(cmd *cobra.Command, args []string) (config.Config, []string, error) {
	cfg, err := config.Load(flagConfig)
	if err != nil {
		return cfg, nil, err
	}

	if flagLicense != "" {
		cfg.License = flagLicense
	}
	if cmd.Flags().Changed("owner") {
		cfg.CopyrightHolder = flagOwner
	}
	if flagYear != "" {
		cfg.Year = flagYear
	}
	if flagFormat != "" {
		cfg.Format = flagFormat
	}
	if flagConfig != "" {
		cfg.Exclude = append(cfg.Exclude, flagConfig)
	}

	if len(args) > 0 {
		cfg.Paths = args
	}

	warnings, err := cfg.Validate()
	if err != nil {
		return cfg, nil, err
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	return cfg, warnings, nil
}

// printStructured handles --output json/sarif. Returns true if it printed.
func printStructured(result *engine.Result, totalScanned int) bool {
	if flagOutputFormat == "text" || flagOutputFormat == "" {
		return false
	}
	report := output.FromResult(result, totalScanned)
	var data []byte
	var err error
	switch flagOutputFormat {
	case "json":
		data, err = output.FormatJSON(report)
	case "sarif":
		data, err = output.FormatSARIF(report, version)
	default:
		fmt.Fprintf(os.Stderr, "unknown output format: %s\n", flagOutputFormat)
		return false
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
		return false
	}
	fmt.Println(string(data))
	return true
}

// exitWithResult sets the exit code based on result state (exit code 3 for partial failure).
func exitWithResult(result *engine.Result) {
	hasErrors := len(result.Errors) > 0
	hasNonCompliant := len(result.NonCompliant) > 0
	switch {
	case hasErrors && hasNonCompliant:
		os.Exit(3) // partial failure
	case hasNonCompliant:
		os.Exit(1) // non-compliant
	case hasErrors:
		os.Exit(2) // runtime error
	}
}

func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [paths...]",
		Short: "Check files for compliant license headers",
		Long: `Check that all source files have valid license headers.

Exits with code 0 if all files are compliant, code 1 if any files are
missing or have incorrect headers, code 2 on runtime errors, and
code 3 on partial failure (some errors and some non-compliant).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, warnings, err := loadConfig(cmd, args)
			if err != nil {
				return err
			}

			eng, err := engine.New(cfg)
			if err != nil {
				return err
			}
			eng.SetWarnings(warnings)

			result, err := eng.Check(flagVerbose)
			if err != nil {
				return err
			}

			totalScanned := len(result.NonCompliant) + len(result.Skipped) + len(result.Errors)
			// Count compliant files (not in any other bucket)
			// Total = compliant + non-compliant + skipped + errors
			// We approximate total from what engine tells us; compliant aren't tracked by name.
			// A rough total for structured output:
			totalScanned += estimateCompliant(result)

			if printStructured(result, totalScanned) {
				exitWithResult(result)
				return nil
			}

			for path, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "  error: %s: %v\n", path, e)
			}

			if len(result.NonCompliant) > 0 {
				fmt.Printf("Missing or invalid license headers (%d files):\n", len(result.NonCompliant))
				for _, f := range result.NonCompliant {
					fmt.Printf("  %s\n", f)
				}
				for _, w := range result.Warnings {
					if strings.Contains(w, "deprecated") {
						fmt.Fprintf(os.Stderr, "\nhint: %s\n", w)
						fmt.Fprintf(os.Stderr, "  run 'lops fix' to update headers to the current license ID\n")
					}
				}
			}

			if len(result.NonCompliant) == 0 && len(result.Errors) == 0 {
				fmt.Println("All files have valid license headers.")
				return nil
			}

			exitWithResult(result)
			return nil
		},
	}
}

func fixCmd() *cobra.Command {
	var showDiff bool

	cmd := &cobra.Command{
		Use:   "fix [paths...]",
		Short: "Add or replace license headers in source files",
		Long: `Fix source files by adding or replacing license headers.

Files that already have valid headers are left unchanged. Files with
incorrect or missing headers will have the correct header added or replaced.

Use --dry-run to see what would change without modifying files.
Use --diff to show a unified diff of changes (implies --dry-run).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, warnings, err := loadConfig(cmd, args)
			if err != nil {
				return err
			}

			eng, err := engine.New(cfg)
			if err != nil {
				return err
			}
			eng.SetWarnings(warnings)

			var result *engine.Result

			if showDiff {
				result, err = eng.FixWithDiff(flagVerbose)
			} else {
				result, err = eng.Fix(flagDryRun, flagVerbose)
			}
			if err != nil {
				return err
			}

			if printStructured(result, estimateTotal(result)) {
				return nil
			}

			for path, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "  error: %s: %v\n", path, e)
			}

			if showDiff {
				for _, path := range result.NonCompliant {
					if d, ok := result.Diffs[path]; ok {
						fmt.Print(d)
					}
				}
				if len(result.NonCompliant) > 0 {
					fmt.Printf("\nWould fix %d files.\n", len(result.NonCompliant))
				} else {
					fmt.Println("All files already have valid license headers.")
				}
				return nil
			}

			if flagDryRun {
				if len(result.NonCompliant) > 0 {
					fmt.Printf("Would fix %d files:\n", len(result.NonCompliant))
					for _, f := range result.NonCompliant {
						fmt.Printf("  %s\n", f)
					}
				} else {
					fmt.Println("All files already have valid license headers.")
				}
				return nil
			}

			if len(result.Fixed) > 0 {
				fmt.Printf("Fixed %d files:\n", len(result.Fixed))
				for _, f := range result.Fixed {
					fmt.Printf("  %s\n", f)
				}
			} else {
				fmt.Println("All files already have valid license headers.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "show what would change without modifying files")
	cmd.Flags().BoolVar(&showDiff, "diff", false, "show unified diff of changes (implies --dry-run)")
	return cmd
}

func initCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a starter .licenseops.yaml config file",
		Long: `Interactively generate a .licenseops.yaml configuration file.

Detects your project type (Go, Node, Python, Rust) and suggests
sensible exclude patterns. You can also provide values via flags
for non-interactive use.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			const configFile = ".licenseops.yaml"

			if !force {
				if _, err := os.Stat(configFile); err == nil {
					return fmt.Errorf("%s already exists (use --force to overwrite)", configFile)
				}
			}

			license := flagLicense
			holder := flagOwner
			format := flagFormat

			reader := bufio.NewReader(os.Stdin)

			if license == "" {
				fmt.Print("SPDX license identifier (e.g. MIT, Apache-2.0): ")
				input, _ := reader.ReadString('\n')
				license = strings.TrimSpace(input)
				if license == "" {
					return fmt.Errorf("license is required")
				}
			}

			if holder == "" && !cmd.Flags().Changed("owner") {
				fmt.Print("Copyright holder (leave empty for SPDX 1-line mode): ")
				input, _ := reader.ReadString('\n')
				holder = strings.TrimSpace(input)
			}

			if format == "" {
				format = "spdx"
			}

			pt := initcmd.DetectProjectType(".")
			suggested := initcmd.SuggestExcludes(pt)
			if pt != initcmd.ProjectGeneric {
				fmt.Printf("Detected %s project.\n", pt)
			}

			content := initcmd.GenerateConfig(license, holder, format, suggested)

			if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			fmt.Printf("Created %s\n", configFile)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing config file")
	return cmd
}

func removeCmd() *cobra.Command {
	var dryRun bool
	var excludedOnly bool

	cmd := &cobra.Command{
		Use:   "remove [paths...]",
		Short: "Strip license headers from source files",
		Long: `Remove recognized license headers from source files.

Tries all known header formats (SPDX, REUSE, Apache, GPL) to detect
and strip existing headers. Files without recognized headers are left
unchanged.

Use --dry-run to see what would change without modifying files.

Use --excluded-only to invert the scan and process only files that
match an 'exclude' pattern in your config. This is the cleanup helper
for cases where you've added new excludes and want to strip leftover
headers from those files in one shot.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, warnings, err := loadConfig(cmd, args)
			if err != nil {
				return err
			}

			// In --excluded-only mode we want the inverse scan to target
			// only the patterns the user explicitly declared in their
			// config file — not built-in defaults like .git/** or the
			// auto-added config-file path. Swap to UserExcludes for that.
			if excludedOnly {
				cfg.Exclude = cfg.UserExcludes
			}

			eng, err := engine.New(cfg)
			if err != nil {
				return err
			}
			eng.SetWarnings(warnings)
			if excludedOnly {
				eng.SetInverseExcludes(true)
			}

			result, err := eng.Remove(dryRun, flagVerbose)
			if err != nil {
				return err
			}

			for path, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "  error: %s: %v\n", path, e)
			}

			if dryRun {
				if len(result.NonCompliant) > 0 {
					fmt.Printf("Would remove headers from %d files:\n", len(result.NonCompliant))
					for _, f := range result.NonCompliant {
						fmt.Printf("  %s\n", f)
					}
				} else {
					fmt.Println("No files with recognized license headers found.")
				}
				return nil
			}

			if len(result.Fixed) > 0 {
				fmt.Printf("Removed headers from %d files:\n", len(result.Fixed))
				for _, f := range result.Fixed {
					fmt.Printf("  %s\n", f)
				}
			} else {
				fmt.Println("No files with recognized license headers found.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would change without modifying files")
	cmd.Flags().BoolVar(&excludedOnly, "excluded-only", false, "only process files matching user-defined exclude patterns (cleanup helper)")
	return cmd
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration file without scanning files",
		Long:  `Check that the .licenseops.yaml configuration is valid without scanning any files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagConfig)
			if err != nil {
				return err
			}

			// Apply CLI overrides
			if flagLicense != "" {
				cfg.License = flagLicense
			}
			if cmd.Flags().Changed("owner") {
				cfg.CopyrightHolder = flagOwner
			}
			if flagFormat != "" {
				cfg.Format = flagFormat
			}

			warnings, err := cfg.Validate()
			for _, w := range warnings {
				fmt.Fprintf(os.Stderr, "warning: %s\n", w)
			}
			if err != nil {
				return err
			}

			fmt.Println("Configuration is valid.")
			return nil
		},
	}
}

// estimateCompliant returns a rough count of compliant files (not tracked by name in Result).
// This is used only for structured output totals.
func estimateCompliant(_ *engine.Result) int {
	// The engine doesn't track compliant files by name. For structured output,
	// we can only report what we know. Return 0 for now; the total will be
	// non-compliant + skipped + errors.
	return 0
}

func estimateTotal(result *engine.Result) int {
	return len(result.NonCompliant) + len(result.Skipped) + len(result.Errors)
}
