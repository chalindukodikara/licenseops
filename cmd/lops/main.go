// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/chalindukodikara/licenseops/internal/config"
	"github.com/chalindukodikara/licenseops/internal/engine"
)

var version = "dev"

var (
	flagConfig  string
	flagLicense string
	flagOwner   string
	flagYear    string
	flagFormat  string
	flagVerbose bool
	flagDryRun  bool
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

	rootCmd.AddCommand(checkCmd())
	rootCmd.AddCommand(fixCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

func loadConfig(args []string) (config.Config, error) {
	cfg, err := config.Load(flagConfig)
	if err != nil {
		return cfg, err
	}

	if flagLicense != "" {
		cfg.License = flagLicense
	}
	if flagOwner != "" {
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
		return cfg, err
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	return cfg, nil
}

func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [paths...]",
		Short: "Check files for compliant license headers",
		Long: `Check that all source files have valid license headers.

Exits with code 0 if all files are compliant, code 1 if any files are
missing or have incorrect headers, and code 2 on runtime errors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(args)
			if err != nil {
				return err
			}

			eng, err := engine.New(cfg)
			if err != nil {
				return err
			}

			result, err := eng.Check(flagVerbose)
			if err != nil {
				return err
			}

			for path, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "  error: %s: %v\n", path, e)
			}

			if len(result.NonCompliant) > 0 {
				fmt.Printf("Missing or invalid license headers (%d files):\n", len(result.NonCompliant))
				for _, f := range result.NonCompliant {
					fmt.Printf("  %s\n", f)
				}
				os.Exit(1)
			}

			fmt.Println("All files have valid license headers.")
			return nil
		},
	}
}

func fixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix [paths...]",
		Short: "Add or replace license headers in source files",
		Long: `Fix source files by adding or replacing license headers.

Files that already have valid headers are left unchanged. Files with
incorrect or missing headers will have the correct header added or replaced.

Use --dry-run to see what would change without modifying files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(args)
			if err != nil {
				return err
			}

			eng, err := engine.New(cfg)
			if err != nil {
				return err
			}

			result, err := eng.Fix(flagDryRun, flagVerbose)
			if err != nil {
				return err
			}

			for path, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "  error: %s: %v\n", path, e)
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
	return cmd
}
