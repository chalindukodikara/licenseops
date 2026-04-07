# LicenseOps (`lops`)

A fast CLI tool to check, fix, and migrate license headers across 50+ languages. Supports SPDX, REUSE, Apache/GPL boilerplates, and custom templates. Built for CI pipelines, pre-commit hooks, and local development.

> **Why "LicenseOps"?** — "License" says what it manages, "Ops" signals it's an operational tool built for automation — CI pipelines, pre-commit hooks, workflows. The CLI binary is `lops` (**L**icense**Ops**) — 4 characters, fast to type, same pattern as Kubernetes → `kubectl`, Terraform → `tf`.

## Features

- **Check, Fix, Remove** — validate headers, auto-add/replace them, or strip them entirely
- **Multiple formats** — SPDX short (1-line and 2-line), REUSE, Apache 2.0 boilerplate, GPL/LGPL/AGPL boilerplate, custom templates
- **50+ languages** — correct comment syntax for Go, Rust, Python, JavaScript/TypeScript, Java, C/C++, Shell, YAML, CSS, HTML, SQL, and [many more](docs/supported-languages.md)
- **SPDX expressions** — `Apache-2.0 OR MIT`, `GPL-3.0-only WITH Classpath-exception-2.0`
- **Smart handling** — preserves shebangs and Python encoding declarations, skips generated files and binaries
- **Gitignore-aware** — respects `.gitignore` patterns automatically
- **Cross-format migration** — switch from one header format to another without manual cleanup; old headers are fully detected, stripped, and replaced
- **Parallel processing** — files are processed concurrently for fast execution on large codebases
- **Structured output** — `--output json` and `--output sarif` for CI tooling and GitHub Code Scanning
- **CI-ready** — exit codes 0/1/2/3, `--dry-run`, `--diff`, Docker image, GitHub Actions compatible
- **Zero config viable** — works with just `lops check -l MIT -o "Your Name" .`

## Installation

### Binary

Download from [Releases](https://github.com/licenseops/licenseops/releases):

```bash
curl -sSL https://github.com/licenseops/licenseops/releases/latest/download/lops_Linux_x86_64.tar.gz | tar xz
sudo mv lops /usr/local/bin/
```

### Go install

```bash
go install github.com/licenseops/licenseops/cmd/lops@latest
```

### Docker

```bash
# Latest stable release
docker run --rm -v "$PWD":/src -w /src ghcr.io/licenseops/licenseops:latest check

# Pinned to exact version
docker run --rm -v "$PWD":/src -w /src ghcr.io/licenseops/licenseops:0.1.0 check

# Latest development build (tracks main branch)
docker run --rm -v "$PWD":/src -w /src ghcr.io/licenseops/licenseops:latest-dev check
```

## Quick Start

**Initialize** a config file (interactive):

```bash
lops init
```

Or **check** compliance directly:

```bash
lops check -l Apache-2.0 -o "Acme Corp" .
```

**Fix** headers:

```bash
lops fix -l Apache-2.0 -o "Acme Corp" .
```

**Preview** changes as a unified diff:

```bash
lops fix --diff
```

**Remove** all headers:

```bash
lops remove --dry-run    # preview first
lops remove              # then remove
```

**Config file** — create `.licenseops.yaml` in your project root (or use `lops init`):

```yaml
license: Apache-2.0
copyright-holder: "Acme Corp"
exclude:
  - "vendor/**"
  - "**/*.pb.go"
```

Then simply:

```bash
lops check
lops fix
```

## Header Formats

| Format | Config value | Description |
|--------|-------------|-------------|
| SPDX 2-line | `spdx` (default) | `// Copyright 2026 Acme Corp` + `// SPDX-License-Identifier: Apache-2.0` |
| SPDX 1-line | `spdx` (no owner) | `// SPDX-License-Identifier: MIT` |
| REUSE | `reuse` | `// SPDX-FileCopyrightText: 2026 Acme Corp` + `// SPDX-License-Identifier: MIT` |
| Apache 2.0 | `apache-long` | Full 14-line Apache boilerplate |
| GPL/LGPL/AGPL | `gpl-long` | Full GNU boilerplate (auto-selects GPL, LGPL, or AGPL) |
| Custom | `custom` | User-defined Go template file |

See [docs/formats.md](docs/formats.md) for side-by-side comparison.

## Cross-Format Migration

When you switch from one header format to another (e.g., Apache boilerplate to SPDX short), `lops fix` automatically handles the migration — no manual cleanup needed.

### How it works

1. **Detect** — For each file, `lops` tries stripping headers using every known format (SPDX, REUSE, Apache-long, GPL-long, custom). The format that removes the most content wins, ensuring the old header is fully identified regardless of which format was originally used.
2. **Strip** — The detected old header is completely removed, including blank separator lines. Shebangs (`#!/...`) and Python encoding declarations (`# -*- coding: utf-8 -*-`) are preserved.
3. **Add** — The new header is generated in the target format and inserted at the top of the file (after any shebang/encoding lines).

This means you can migrate between any combination of formats in a single `lops fix` run.

### Example: Apache boilerplate to SPDX

```bash
# Before: files have 14-line Apache boilerplate headers
# Change your config to the new format
# .licenseops.yaml
#   format: spdx
#   license: Apache-2.0
#   copyright-holder: "Acme Corp"

lops fix          # strips Apache boilerplate, adds SPDX 2-line header
lops check        # confirms all files are compliant
```

### Removing all headers

The `remove` command uses the same multi-format detection to strip headers from all files, regardless of which format they were written in:

```bash
lops remove --dry-run    # preview which files have headers
lops remove              # strip all recognized headers
```

Even if your config says `format: spdx`, `remove` will detect and strip Apache, GPL, REUSE, and custom headers too.

## Configuration

`lops` automatically looks for `.licenseops.yaml` in the current directory. If not found, it silently uses built-in defaults — no error. You can always override with CLI flags, or skip the config file entirely.

### `.licenseops.yaml`

```yaml
license: Apache-2.0                # required — SPDX ID or expression
copyright-holder: "Acme Corp"      # omit for SPDX 1-line mode
format: spdx                       # spdx | reuse | apache-long | gpl-long | custom
year: 2026                         # defaults to current year
header-template: ""                # path to template (format: custom only)

paths:                             # directories to scan (default: ["."])
  - "."

exclude:                           # doublestar glob patterns
  - "vendor/**"
  - "**/*_generated.go"

skip-generated: true               # skip "DO NOT EDIT" / "@generated" files
gitignore: true                    # respect .gitignore patterns
```

### CLI Flags

```
-l, --license     SPDX license identifier or expression
-o, --owner       copyright holder
-f, --format      header format (spdx, reuse, apache-long, gpl-long, custom)
-y, --year        copyright year
-c, --config      config file path (default: .licenseops.yaml)
-v, --verbose     show status of every file
    --dry-run     preview changes without modifying files (fix/remove)
    --diff        show unified diff of changes (fix, implies dry-run)
    --output      output format: text (default), json, sarif
```

Precedence: **CLI flags > config file > defaults**

### SPDX Expressions

```bash
lops check -l "Apache-2.0 OR MIT" -o "Acme Corp" .
lops check -l "GPL-3.0-only WITH Classpath-exception-2.0" -o "Acme Corp" .
```

## CI Integration

### GitHub Actions

```yaml
- name: Check license headers
  run: |
    curl -sSL https://github.com/licenseops/licenseops/releases/latest/download/lops_Linux_x86_64.tar.gz | tar xz
    ./lops check
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All files compliant / all files fixed |
| 1 | Non-compliant files found |
| 2 | Runtime error |
| 3 | Partial failure (some errors and some non-compliant) |

### Structured Output

```bash
# JSON output for CI scripting
lops check --output json

# SARIF output for GitHub Code Scanning
lops check --output sarif > results.sarif
```

See [docs/ci-integration.md](docs/ci-integration.md) for GitHub Actions, GitLab CI, pre-commit, and Docker examples.

## Supported Languages

| Style | Languages |
|-------|-----------|
| `//` | Go, Rust, Java, JS/TS, C/C++, C#, Swift, Kotlin, Scala, Dart, Protobuf, Zig |
| `#` | Python, Ruby, Shell, Perl, YAML, TOML, Terraform, Dockerfile, Makefile |
| `--` | Haskell, Lua, SQL, Ada, Elm |
| `/* */` | CSS, SCSS, Less |
| `<!-- -->` | HTML, XML, SVG, Vue |

Full list: [docs/supported-languages.md](docs/supported-languages.md)

## Documentation

| Doc | Description |
|-----|-------------|
| [Configuration](docs/configuration.md) | Config file reference, exclude patterns, expressions |
| [Use Cases](docs/use-cases.md) | 12 real-world scenarios with example configs |
| [Formats](docs/formats.md) | Side-by-side format comparison |
| [Migration](docs/migration.md) | Cross-format migration guide |
| [CI Integration](docs/ci-integration.md) | GitHub Actions, GitLab CI, pre-commit, Docker |
| [Custom Templates](docs/custom-templates.md) | Template syntax and examples |
| [Supported Languages](docs/supported-languages.md) | All 50+ file types |
| [Releasing](docs/releasing.md) | Release process, branching, CI pipeline |
| [Release Plan](RELEASE.md) | Roadmap and planned features per version |

## Development

```bash
make build       # Build binary
make test        # Run tests with race detector
make lint        # Run golangci-lint
make lint-fix    # Auto-fix lint issues
make vet         # Run go vet
make fmt         # Format code
make check       # Self-check license headers
make docker      # Build Docker image
make clean       # Remove build artifacts
make help        # Show all targets
```

## License

Apache-2.0 — see [LICENSE](LICENSE)
