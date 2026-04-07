// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/licenseops/licenseops/internal/engine"
)

func testResult() *engine.Result {
	return &engine.Result{
		NonCompliant: []string{"a.go", "b.go"},
		Fixed:        []string{"a.go"},
		Skipped:      []string{"c.go"},
		Errors:       map[string]error{"d.go": &testErr{}},
	}
}

type testErr struct{}

func (e *testErr) Error() string { return "read error" }

func TestFromResult(t *testing.T) {
	r := FromResult(testResult(), 10)
	if r.Summary.Total != 10 {
		t.Errorf("total = %d", r.Summary.Total)
	}
	if r.Summary.NonCompliant != 2 {
		t.Errorf("non-compliant = %d", r.Summary.NonCompliant)
	}
	if r.Summary.Fixed != 1 {
		t.Errorf("fixed = %d", r.Summary.Fixed)
	}
	if r.Summary.Skipped != 1 {
		t.Errorf("skipped = %d", r.Summary.Skipped)
	}
	if r.Summary.Errors != 1 {
		t.Errorf("errors = %d", r.Summary.Errors)
	}
	if r.Summary.Compliant != 6 {
		t.Errorf("compliant = %d, want 6", r.Summary.Compliant)
	}
}

func TestFormatJSON_Roundtrip(t *testing.T) {
	r := FromResult(testResult(), 10)
	data, err := FormatJSON(r)
	if err != nil {
		t.Fatal(err)
	}
	var parsed Report
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed.Summary.Total != 10 {
		t.Error("roundtrip failed")
	}
}

func TestFormatSARIF_ValidJSON(t *testing.T) {
	r := FromResult(testResult(), 10)
	data, err := FormatSARIF(r, "dev")
	if err != nil {
		t.Fatal(err)
	}
	var sarif SARIFReport
	if err := json.Unmarshal(data, &sarif); err != nil {
		t.Fatalf("invalid SARIF JSON: %v", err)
	}
	if sarif.Version != "2.1.0" {
		t.Errorf("version = %q", sarif.Version)
	}
	if len(sarif.Runs) != 1 {
		t.Fatal("expected 1 run")
	}
	if sarif.Runs[0].Tool.Driver.Name != "lops" {
		t.Error("tool name should be lops")
	}
	// 2 non-compliant files = 2 results (a.go fixed is still non-compliant status in the report)
	// Actually a.go has status "fixed", not "non-compliant", so only b.go is non-compliant
	found := 0
	for _, r := range sarif.Runs[0].Results {
		if r.RuleID == "missing-license-header" {
			found++
		}
	}
	if found != 1 {
		t.Errorf("expected 1 SARIF result (b.go), got %d", found)
	}
}

func TestFormatSARIF_ContainsSchema(t *testing.T) {
	r := FromResult(testResult(), 5)
	data, _ := FormatSARIF(r, "1.0.0")
	if !strings.Contains(string(data), "sarif-schema") {
		t.Error("should contain SARIF schema URL")
	}
}

func TestFormatJSON_EmptyResult(t *testing.T) {
	r := FromResult(&engine.Result{Errors: make(map[string]error)}, 0)
	data, err := FormatJSON(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"total": 0`) {
		t.Error("empty result should have total 0")
	}
}
