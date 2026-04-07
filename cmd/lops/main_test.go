// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// lopsBin is the path to the compiled CLI binary, set by TestMain.
var lopsBin string

func TestMain(m *testing.M) {
	os.Exit(buildAndRun(m))
}

func buildAndRun(m *testing.M) int {
	tmpDir, err := os.MkdirTemp("", "lops-test-bin-")
	if err != nil {
		panic("creating temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	bin := filepath.Join(tmpDir, "lops")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	build := exec.Command("go", "build", "-o", bin, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic("building lops binary: " + err.Error())
	}
	lopsBin = bin

	return m.Run()
}

// runLops runs the CLI binary with the given args inside dir and returns
// stdout, stderr, and exit code. Exit code is 0 on success.
func runLops(t *testing.T, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(lopsBin, args...)
	cmd.Dir = dir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("running lops: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return full
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// --- check ---

func TestCLI_Check_AllCompliant(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n")
	stdout, _, code := runLops(t, dir, "check", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d, want 0; stdout=%q", code, stdout)
	}
	if !strings.Contains(stdout, "All files have valid license headers") {
		t.Errorf("stdout missing success message: %q", stdout)
	}
}

func TestCLI_Check_NonCompliant(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "good.go", "// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n")
	writeFile(t, dir, "bad.go", "package main\n")
	stdout, _, code := runLops(t, dir, "check", "-l", "MIT", "-o", "Acme", ".")
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(stdout, "bad.go") {
		t.Errorf("non-compliant file not listed: %q", stdout)
	}
	if !strings.Contains(stdout, "Missing or invalid license headers") {
		t.Errorf("missing 'Missing or invalid' message: %q", stdout)
	}
}

func TestCLI_Check_Verbose_OK(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n")
	stdout, _, code := runLops(t, dir, "check", "-l", "MIT", "-o", "Acme", "-v", ".")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(stdout, "ok:") {
		t.Errorf("verbose mode should print 'ok:' lines: %q", stdout)
	}
}

// --- fix ---

func TestCLI_Fix_AddsHeaders(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "main.go", "package main\n")
	stdout, _, code := runLops(t, dir, "fix", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d, want 0; stdout=%q", code, stdout)
	}
	content := readFile(t, p)
	if !strings.Contains(content, "// SPDX-License-Identifier: MIT") {
		t.Errorf("file should contain SPDX header after fix:\n%s", content)
	}
	if !strings.Contains(content, "// Copyright") || !strings.Contains(content, "Acme") {
		t.Errorf("file should contain copyright with Acme:\n%s", content)
	}
}

func TestCLI_Fix_DryRun(t *testing.T) {
	dir := t.TempDir()
	original := "package main\n"
	p := writeFile(t, dir, "main.go", original)
	stdout, _, code := runLops(t, dir, "fix", "--dry-run", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(stdout, "Would fix") {
		t.Errorf("dry-run should say 'Would fix': %q", stdout)
	}
	if got := readFile(t, p); got != original {
		t.Errorf("file should be unchanged in dry-run, got:\n%s", got)
	}
}

func TestCLI_Fix_Diff(t *testing.T) {
	dir := t.TempDir()
	original := "package main\n"
	p := writeFile(t, dir, "main.go", original)
	stdout, _, code := runLops(t, dir, "fix", "--diff", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(stdout, "---") || !strings.Contains(stdout, "+++") {
		t.Errorf("diff should look like unified diff: %q", stdout)
	}
	if !strings.Contains(stdout, "SPDX-License-Identifier") {
		t.Errorf("diff should contain new header: %q", stdout)
	}
	if got := readFile(t, p); got != original {
		t.Errorf("file should not be modified by --diff, got:\n%s", got)
	}
}

func TestCLI_Fix_Idempotent(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "main.go", "package main\n")
	runLops(t, dir, "fix", "-l", "MIT", "-o", "Acme", ".")
	first := readFile(t, p)
	stdout, _, code := runLops(t, dir, "fix", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(stdout, "All files already have valid license headers") {
		t.Errorf("second fix should be a no-op: %q", stdout)
	}
	if got := readFile(t, p); got != first {
		t.Errorf("idempotency broken; second fix modified file:\nfirst:\n%s\nsecond:\n%s", first, got)
	}
}

func TestCLI_Fix_MultiLanguage(t *testing.T) {
	dir := t.TempDir()
	goPath := writeFile(t, dir, "main.go", "package main\n")
	pyPath := writeFile(t, dir, "script.py", "print('hi')\n")
	sqlPath := writeFile(t, dir, "query.sql", "SELECT 1;\n")
	cssPath := writeFile(t, dir, "style.css", ".body { color: red; }\n")

	_, _, code := runLops(t, dir, "fix", "-l", "MIT", ".")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.HasPrefix(readFile(t, goPath), "// SPDX") {
		t.Error("Go file should use //")
	}
	if !strings.HasPrefix(readFile(t, pyPath), "# SPDX") {
		t.Error("Python file should use #")
	}
	if !strings.HasPrefix(readFile(t, sqlPath), "-- SPDX") {
		t.Error("SQL file should use --")
	}
	if !strings.HasPrefix(readFile(t, cssPath), "/*") {
		t.Error("CSS file should use /* block */")
	}
}

func TestCLI_Fix_PreservesShebang(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "deploy.sh", "#!/usr/bin/env bash\necho hi\n")
	_, _, code := runLops(t, dir, "fix", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	content := readFile(t, p)
	if !strings.HasPrefix(content, "#!/usr/bin/env bash\n") {
		t.Errorf("shebang should be preserved at the top:\n%s", content)
	}
	if !strings.Contains(content, "# Copyright") {
		t.Errorf("header should be present:\n%s", content)
	}
}

func TestCLI_Fix_GeneratedFilesSkipped(t *testing.T) {
	dir := t.TempDir()
	gen := writeFile(t, dir, "gen.go", "// Code generated by tool. DO NOT EDIT.\npackage gen\n")
	_, _, code := runLops(t, dir, "fix", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	if got := readFile(t, gen); !strings.HasPrefix(got, "// Code generated") {
		t.Errorf("generated file should not be modified, got:\n%s", got)
	}
}

func TestCLI_Fix_CrossFormatMigration(t *testing.T) {
	dir := t.TempDir()
	apacheBoilerplate := strings.Join([]string{
		"// Copyright 2025 OldCorp",
		"//",
		"// Licensed under the Apache License, Version 2.0 (the \"License\");",
		"// you may not use this file except in compliance with the License.",
		"// You may obtain a copy of the License at",
		"//",
		"//     http://www.apache.org/licenses/LICENSE-2.0",
		"//",
		"// Unless required by applicable law or agreed to in writing, software",
		"// distributed under the License is distributed on an \"AS IS\" BASIS,",
		"// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.",
		"// See the License for the specific language governing permissions and",
		"// limitations under the License.",
		"",
		"package main",
		"",
	}, "\n")
	p := writeFile(t, dir, "main.go", apacheBoilerplate)

	_, _, code := runLops(t, dir, "fix", "-l", "MIT", "-o", "NewCorp", "-f", "spdx", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	content := readFile(t, p)
	if strings.Contains(content, "Apache License") {
		t.Errorf("Apache header should be stripped:\n%s", content)
	}
	if !strings.Contains(content, "SPDX-License-Identifier: MIT") {
		t.Errorf("SPDX header should be added:\n%s", content)
	}
	if !strings.Contains(content, "NewCorp") {
		t.Errorf("new copyright holder should appear:\n%s", content)
	}
}

func TestCLI_Fix_SPDXOneLine(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "main.go", "package main\n")
	// no -o flag → 1-line mode
	_, _, code := runLops(t, dir, "fix", "-l", "MIT", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	content := readFile(t, p)
	if !strings.HasPrefix(content, "// SPDX-License-Identifier: MIT\n") {
		t.Errorf("expected 1-line SPDX header, got:\n%s", content)
	}
	if strings.Contains(content, "Copyright") {
		t.Errorf("1-line mode should not have Copyright line:\n%s", content)
	}
}

func TestCLI_Fix_SPDXExpression(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "main.go", "package main\n")
	_, _, code := runLops(t, dir, "fix", "-l", "Apache-2.0 OR MIT", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	content := readFile(t, p)
	if !strings.Contains(content, "SPDX-License-Identifier: Apache-2.0 OR MIT") {
		t.Errorf("expression should be in header:\n%s", content)
	}
}

// --- format flag ---

func TestCLI_Format_Reuse(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "main.go", "package main\n")
	_, _, code := runLops(t, dir, "fix", "-l", "MIT", "-o", "Acme", "-f", "reuse", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	content := readFile(t, p)
	if !strings.Contains(content, "SPDX-FileCopyrightText: 2026 Acme") &&
		!strings.Contains(content, "SPDX-FileCopyrightText:") {
		t.Errorf("reuse header should be present:\n%s", content)
	}
}

func TestCLI_Format_ApacheLong(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "main.go", "package main\n")
	_, _, code := runLops(t, dir, "fix", "-l", "Apache-2.0", "-o", "Acme", "-f", "apache-long", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	content := readFile(t, p)
	if !strings.Contains(content, "Licensed under the Apache License") {
		t.Errorf("apache-long boilerplate should be present:\n%s", content)
	}
}

func TestCLI_Format_GPLLong(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "main.go", "package main\n")
	_, _, code := runLops(t, dir, "fix", "-l", "GPL-3.0-only", "-o", "Acme", "-f", "gpl-long", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	content := readFile(t, p)
	if !strings.Contains(content, "GNU General Public License") {
		t.Errorf("gpl-long boilerplate should be present:\n%s", content)
	}
}

// --- validation errors ---

func TestCLI_Error_MissingLicense(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	_, stderr, code := runLops(t, dir, "check", ".")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "license is required") {
		t.Errorf("stderr should mention missing license: %q", stderr)
	}
}

func TestCLI_Error_ApacheLongWithMIT(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	_, stderr, code := runLops(t, dir, "check", "-l", "MIT", "-o", "Acme", "-f", "apache-long", ".")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "Apache-2.0") {
		t.Errorf("error should mention Apache-2.0: %q", stderr)
	}
}

func TestCLI_Error_GPLLongWithMIT(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	_, stderr, code := runLops(t, dir, "check", "-l", "MIT", "-o", "Acme", "-f", "gpl-long", ".")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "GPL") {
		t.Errorf("error should mention GPL: %q", stderr)
	}
}

func TestCLI_Error_InvalidLicense(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	_, stderr, code := runLops(t, dir, "check", "-l", "INVALID-LICENSE", ".")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "unknown") {
		t.Errorf("error should mention unknown: %q", stderr)
	}
}

func TestCLI_Error_UnknownFormat(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	_, stderr, code := runLops(t, dir, "check", "-l", "MIT", "-f", "nonexistent", ".")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "unknown format") {
		t.Errorf("error should mention unknown format: %q", stderr)
	}
}

func TestCLI_Error_DeprecatedGNUWarning(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	_, stderr, code := runLops(t, dir, "check", "-l", "GPL-2.0", ".")
	// deprecated is a warning, not an error — exit 1 because the file is non-compliant
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "deprecat") {
		t.Errorf("stderr should contain deprecation warning: %q", stderr)
	}
}

// --- config file ---

func TestCLI_ConfigFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".licenseops.yaml", "license: MIT\ncopyright-holder: \"Acme Corp\"\n")
	writeFile(t, dir, "main.go", "package main\n")

	stdout, _, code := runLops(t, dir, "check", ".")
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(stdout, "main.go") {
		t.Errorf("config should be read from .licenseops.yaml: %q", stdout)
	}
}

func TestCLI_ConfigFile_FlagOverrides(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".licenseops.yaml", "license: MIT\ncopyright-holder: \"FromConfig\"\n")
	p := writeFile(t, dir, "main.go", "package main\n")

	_, _, code := runLops(t, dir, "fix", "-o", "FromFlag", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	content := readFile(t, p)
	if !strings.Contains(content, "FromFlag") {
		t.Errorf("flag should override config holder:\n%s", content)
	}
	if strings.Contains(content, "FromConfig") {
		t.Errorf("config holder should be overridden:\n%s", content)
	}
}

// --- excludes ---

func TestCLI_ExcludePattern(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n")
	writeFile(t, dir, "vendor/lib.go", "package vendor\n") // not compliant, but excluded by default

	_, _, code := runLops(t, dir, "check", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d, want 0 (vendor should be excluded)", code)
	}
}

// --- structured output ---

func TestCLI_OutputJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "good.go", "// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n")
	writeFile(t, dir, "bad.go", "package main\n")

	stdout, _, code := runLops(t, dir, "check", "-l", "MIT", "-o", "Acme", "--output", "json", ".")
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	var report map[string]any
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, stdout)
	}
	if _, ok := report["summary"]; !ok {
		t.Error("JSON should have summary key")
	}
	if _, ok := report["files"]; !ok {
		t.Error("JSON should have files key")
	}
}

func TestCLI_OutputSARIF(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.go", "package main\n")

	stdout, _, code := runLops(t, dir, "check", "-l", "MIT", "--output", "sarif", ".")
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	var sarif map[string]any
	if err := json.Unmarshal([]byte(stdout), &sarif); err != nil {
		t.Fatalf("output is not valid SARIF JSON: %v\n%s", err, stdout)
	}
	if v, _ := sarif["version"].(string); v != "2.1.0" {
		t.Errorf("SARIF version = %v, want 2.1.0", sarif["version"])
	}
	if _, ok := sarif["runs"]; !ok {
		t.Error("SARIF should have runs key")
	}
}

// --- init ---

func TestCLI_Init_NonInteractive(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runLops(t, dir, "init", "-l", "MIT", "-o", "Acme")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	cfg := filepath.Join(dir, ".licenseops.yaml")
	if _, err := os.Stat(cfg); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	content := readFile(t, cfg)
	if !strings.Contains(content, "license: MIT") {
		t.Errorf("config should contain license: %s", content)
	}
	if !strings.Contains(content, "Acme") {
		t.Errorf("config should contain holder: %s", content)
	}
}

func TestCLI_Init_FailsIfExists(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".licenseops.yaml", "license: MIT\n")
	_, stderr, code := runLops(t, dir, "init", "-l", "MIT", "-o", "Acme")
	if code == 0 {
		t.Error("init should fail when config exists without --force")
	}
	if !strings.Contains(stderr, "already exists") {
		t.Errorf("stderr should mention existing file: %q", stderr)
	}
}

func TestCLI_Init_Force(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".licenseops.yaml", "license: OLD\n")
	_, _, code := runLops(t, dir, "init", "-l", "MIT", "-o", "Acme", "--force")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	content := readFile(t, filepath.Join(dir, ".licenseops.yaml"))
	if !strings.Contains(content, "license: MIT") {
		t.Errorf("config should be overwritten: %s", content)
	}
}

func TestCLI_Init_DetectsGoProject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/foo\n")
	stdout, _, code := runLops(t, dir, "init", "-l", "MIT", "-o", "Acme")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	if !strings.Contains(stdout, "Detected go project") {
		t.Errorf("should detect Go project: %q", stdout)
	}
	cfg := readFile(t, filepath.Join(dir, ".licenseops.yaml"))
	if !strings.Contains(cfg, "*.pb.go") && !strings.Contains(cfg, "testdata") {
		t.Errorf("Go project should suggest excludes: %s", cfg)
	}
}

// --- remove ---

func TestCLI_Remove_StripsHeader(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "main.go", "// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n")
	stdout, _, code := runLops(t, dir, "remove", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(stdout, "Removed headers") {
		t.Errorf("stdout missing 'Removed headers': %q", stdout)
	}
	content := readFile(t, p)
	if strings.Contains(content, "SPDX-License-Identifier") {
		t.Errorf("header should be stripped:\n%s", content)
	}
	if !strings.Contains(content, "package main") {
		t.Errorf("code should be preserved:\n%s", content)
	}
}

func TestCLI_Remove_DryRun(t *testing.T) {
	dir := t.TempDir()
	original := "// Copyright 2026 Acme\n// SPDX-License-Identifier: MIT\n\npackage main\n"
	p := writeFile(t, dir, "main.go", original)
	stdout, _, code := runLops(t, dir, "remove", "--dry-run", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	if !strings.Contains(stdout, "Would remove") {
		t.Errorf("dry-run should say 'Would remove': %q", stdout)
	}
	if got := readFile(t, p); got != original {
		t.Errorf("file should be unchanged in dry-run, got:\n%s", got)
	}
}

func TestCLI_Remove_NoHeader(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	stdout, _, code := runLops(t, dir, "remove", "-l", "MIT", "-o", "Acme", ".")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	if !strings.Contains(stdout, "No files with recognized") {
		t.Errorf("stdout should report no files: %q", stdout)
	}
}

// --- validate ---

func TestCLI_Validate_Valid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".licenseops.yaml", "license: MIT\ncopyright-holder: \"Acme\"\n")
	stdout, _, code := runLops(t, dir, "validate")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(stdout, "Configuration is valid") {
		t.Errorf("stdout should confirm valid: %q", stdout)
	}
}

func TestCLI_Validate_Invalid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".licenseops.yaml", "license: MIT\nformat: apache-long\ncopyright-holder: \"Acme\"\n")
	_, stderr, code := runLops(t, dir, "validate")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "Apache-2.0") {
		t.Errorf("stderr should mention Apache-2.0: %q", stderr)
	}
}

func TestCLI_Validate_MissingLicense(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".licenseops.yaml", "format: spdx\n")
	_, stderr, code := runLops(t, dir, "validate")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "license") {
		t.Errorf("stderr should mention license: %q", stderr)
	}
}

// --- version ---

func TestCLI_Version(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := runLops(t, dir, "--version")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(stdout, "lops") && !strings.Contains(stdout, "version") {
		t.Errorf("stdout should mention version: %q", stdout)
	}
}

// --- help ---

func TestCLI_Help(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := runLops(t, dir, "--help")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	for _, sub := range []string{"check", "fix", "init", "remove", "validate"} {
		if !strings.Contains(stdout, sub) {
			t.Errorf("help should mention %q subcommand: %q", sub, stdout)
		}
	}
}
