// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"encoding/json"

	"github.com/licenseops/licenseops/internal/engine"
)

// FileEntry represents a single file in structured output.
type FileEntry struct {
	Path   string `json:"path"`
	Status string `json:"status"` // "non-compliant", "fixed", "skipped", "error"
	Detail string `json:"detail,omitempty"`
}

// Summary holds aggregate counts.
type Summary struct {
	Total        int `json:"total"`
	Compliant    int `json:"compliant"`
	NonCompliant int `json:"non_compliant"`
	Fixed        int `json:"fixed"`
	Skipped      int `json:"skipped"`
	Errors       int `json:"errors"`
}

// Report is the top-level structured output.
type Report struct {
	Summary Summary     `json:"summary"`
	Files   []FileEntry `json:"files"`
}

// FromResult converts an engine.Result into a structured Report.
func FromResult(result *engine.Result, totalScanned int) Report {
	r := Report{
		Summary: Summary{
			Total:        totalScanned,
			NonCompliant: len(result.NonCompliant),
			Fixed:        len(result.Fixed),
			Skipped:      len(result.Skipped),
			Errors:       len(result.Errors),
		},
	}
	r.Summary.Compliant = totalScanned - r.Summary.NonCompliant - r.Summary.Skipped - r.Summary.Errors

	for _, p := range result.NonCompliant {
		status := "non-compliant"
		if contains(result.Fixed, p) {
			status = "fixed"
		}
		r.Files = append(r.Files, FileEntry{Path: p, Status: status})
	}
	for _, p := range result.Skipped {
		r.Files = append(r.Files, FileEntry{Path: p, Status: "skipped"})
	}
	for p, err := range result.Errors {
		r.Files = append(r.Files, FileEntry{Path: p, Status: "error", Detail: err.Error()})
	}

	return r
}

// FormatJSON renders the report as indented JSON.
func FormatJSON(r Report) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// SARIFReport is a minimal SARIF v2.1.0 report.
type SARIFReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

// SARIFRun represents a single analysis run.
type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

// SARIFTool describes the tool that produced results.
type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

// SARIFDriver identifies the tool.
type SARIFDriver struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// SARIFResult is a single finding.
type SARIFResult struct {
	RuleID    string          `json:"ruleId"`
	Message   SARIFMessage    `json:"message"`
	Locations []SARIFLocation `json:"locations"`
	Level     string          `json:"level"`
}

// SARIFMessage holds the result text.
type SARIFMessage struct {
	Text string `json:"text"`
}

// SARIFLocation points to a file and region.
type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

// SARIFPhysicalLocation contains the artifact and region.
type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

// SARIFArtifactLocation holds the file URI.
type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

// SARIFRegion specifies the line.
type SARIFRegion struct {
	StartLine int `json:"startLine"`
}

// FormatSARIF renders the report as SARIF JSON.
func FormatSARIF(r Report, toolVersion string) ([]byte, error) {
	sarif := SARIFReport{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []SARIFRun{{
			Tool: SARIFTool{
				Driver: SARIFDriver{
					Name:    "lops",
					Version: toolVersion,
				},
			},
		}},
	}

	for _, f := range r.Files {
		if f.Status != "non-compliant" {
			continue
		}
		sarif.Runs[0].Results = append(sarif.Runs[0].Results, SARIFResult{
			RuleID:  "missing-license-header",
			Message: SARIFMessage{Text: "Missing or invalid license header"},
			Level:   "error",
			Locations: []SARIFLocation{{
				PhysicalLocation: SARIFPhysicalLocation{
					ArtifactLocation: SARIFArtifactLocation{URI: f.Path},
					Region:           SARIFRegion{StartLine: 1},
				},
			}},
		})
	}

	return json.MarshalIndent(sarif, "", "  ")
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
