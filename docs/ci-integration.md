# CI Integration Guide

How to integrate licenseops into your CI/CD pipelines.

## Exit Codes

| Code | Meaning | When |
|------|---------|------|
| 0 | Success | All files compliant (check) or all fixed (fix) |
| 1 | Non-compliant | Files found with missing/incorrect headers |
| 2 | Error | Bad config, IO error, invalid license |

---

## GitHub Actions

### Basic Check

```yaml
name: License Headers
on: [pull_request]

jobs:
  license-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install licenseops
        run: |
          curl -sSL https://github.com/chalindukodikara/licenseops/releases/latest/download/licenser_Linux_x86_64.tar.gz | tar xz
          sudo mv licenseops /usr/local/bin/

      - name: Check license headers
        run: licenseops check
```

### With Go Install

```yaml
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Install licenseops
        run: go install github.com/chalindukodikara/licenseops/cmd/licenseops@latest

      - name: Check license headers
        run: licenseops check
```

### With Docker

```yaml
      - name: Check license headers
        run: |
          docker run --rm \
            -v "${{ github.workspace }}":/src \
            -w /src \
            ghcr.io/chalindukodikara/licenseops check
```

### Auto-Fix on PR (Commit Changes)

```yaml
name: License Headers Fix
on: [pull_request]

permissions:
  contents: write

jobs:
  license-fix:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.head_ref }}

      - name: Install licenseops
        run: |
          curl -sSL https://github.com/chalindukodikara/licenseops/releases/latest/download/licenser_Linux_x86_64.tar.gz | tar xz
          sudo mv licenseops /usr/local/bin/

      - name: Fix license headers
        run: licenseops fix

      - name: Commit changes
        uses: stefanzweifel/git-auto-commit-action@v5
        with:
          commit_message: "chore: fix license headers"
```

### Combined with Other Checks

```yaml
name: CI
on: [pull_request]

jobs:
  checks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Install licenseops
        run: go install github.com/chalindukodikara/licenseops/cmd/licenseops@latest

      - name: License headers
        run: licenseops check

      - name: Lint
        uses: golangci/golangci-lint-action@v6

      - name: Test
        run: go test ./... -race
```

---

## GitLab CI

```yaml
license-check:
  image: golang:1.26-alpine
  stage: lint
  script:
    - go install github.com/chalindukodikara/licenseops/cmd/licenseops@latest
    - licenseops check
  rules:
    - if: $CI_MERGE_REQUEST_ID
```

---

## Pre-commit

### Setup

Add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/chalindukodikara/licenseops
    rev: v0.1.0
    hooks:
      - id: licenseops-check
        name: Check license headers
        entry: licenseops check
        language: golang
        additional_dependencies: ["github.com/chalindukodikara/licenseops/cmd/licenseops@v0.1.0"]
        always_run: true
        pass_filenames: false
```

### Auto-fix Mode

```yaml
repos:
  - repo: https://github.com/chalindukodikara/licenseops
    rev: v0.1.0
    hooks:
      - id: licenseops-fix
        name: Fix license headers
        entry: licenseops fix
        language: golang
        additional_dependencies: ["github.com/chalindukodikara/licenseops/cmd/licenseops@v0.1.0"]
        always_run: true
        pass_filenames: false
```

---

## Makefile Integration

Add these targets to your project's Makefile:

```makefile
.PHONY: license-check license-fix

license-check: ## Check license headers
	licenseops check

license-fix: ## Fix license headers
	licenseops fix
```

Then use in CI:

```yaml
      - name: License check
        run: make license-check
```

---

## Docker Usage

### Check a Local Project

```bash
docker run --rm -v "$(pwd)":/src -w /src ghcr.io/chalindukodikara/licenseops check
```

### Fix a Local Project

```bash
docker run --rm -v "$(pwd)":/src -w /src ghcr.io/chalindukodikara/licenseops fix
```

### With Custom Config

```bash
docker run --rm -v "$(pwd)":/src -w /src ghcr.io/chalindukodikara/licenseops check -c my-config.yaml
```

### Override All Settings via Flags

```bash
docker run --rm -v "$(pwd)":/src -w /src ghcr.io/chalindukodikara/licenseops check \
  -l "Apache-2.0" \
  -o "My Org" \
  -f spdx
```

---

## Tips

- **Fail fast**: Put the license check early in your pipeline — it's fast and catches common issues.
- **Config in repo**: Commit `.licenseops.yaml` so CI and local dev use the same settings.
- **Dry run first**: Use `licenseops fix --dry-run` in CI to report what would change without modifying files.
- **Verbose in CI**: Use `-v` to see every file in CI logs for debugging.
