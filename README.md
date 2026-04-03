# LicenseOps (`lops`)

A fast CLI tool to check, fix, and migrate license headers across 50+ languages. Supports SPDX, REUSE, Apache/GPL boilerplates, and custom templates. Built for CI pipelines, pre-commit hooks, and local development.

> **Why "LicenseOps"?** — "License" says what it manages, "Ops" signals it's an operational tool built for automation — CI pipelines, pre-commit hooks, workflows. The CLI binary is `lops` (**L**icense**Ops**) — 4 characters, fast to type, same pattern as Kubernetes → `kubectl`, Terraform → `tf`.

## Features

- **Check & Fix** — validate headers or auto-add/replace them in-place
- **Multiple formats** — SPDX short (1-line and 2-line), REUSE, Apache 2.0 boilerplate, GPL/LGPL/AGPL boilerplate, custom templates
- **50+ languages** — correct comment syntax for Go, Rust, Python, JavaScript/TypeScript, Java, C/C++, Shell, YAML, CSS, HTML, SQL, and [many more](docs/supported-languages.md)
- **SPDX expressions** — `Apache-2.0 OR MIT`, `GPL-3.0-only WITH Classpath-exception-2.0`
- **Smart handling** — preserves shebangs and Python encoding declarations, skips generated files and binaries
- **Gitignore-aware** — respects `.gitignore` patterns automatically
- **Cross-format migration** — switch from one header format to another without manual cleanup
- **CI-ready** — exit codes, `--dry-run`, Docker image, GitHub Actions compatible
- **Zero config viable** — works with just `lops check -l MIT -o "Your Name" .`

## Installation

### Binary

Download from [Releases](https://github.com/chalindukodikara/licenseops/releases):

```bash
curl -sSL https://github.com/chalindukodikara/licenseops/releases/latest/download/lops_Linux_x86_64.tar.gz | tar xz
sudo mv lops /usr/local/bin/
```

### Go install

```bash
go install github.com/chalindukodikara/licenseops/cmd/lops@latest
```

### Docker

```bash
docker run --rm -v "$PWD":/src -w /src ghcr.io/chalindukodikara/licenseops check
```

## Quick Start

**Check** compliance:

```bash
lops check -l Apache-2.0 -o "Acme Corp" .
```

**Fix** headers:

```bash
lops fix -l Apache-2.0 -o "Acme Corp" .
```

**Config file** — create `.licenseops.yaml` in your project root:

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
    --dry-run     preview changes without modifying files (fix only)
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
    curl -sSL https://github.com/chalindukodikara/licenseops/releases/latest/download/lops_Linux_x86_64.tar.gz | tar xz
    ./lops check
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All files compliant / all files fixed |
| 1 | Non-compliant files found |
| 2 | Runtime error |

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
| [CI Integration](docs/ci-integration.md) | GitHub Actions, GitLab CI, pre-commit, Docker |
| [Custom Templates](docs/custom-templates.md) | Template syntax and examples |
| [Supported Languages](docs/supported-languages.md) | All 50+ file types |
| [Releasing](docs/releasing.md) | Release process, branching, CI pipeline |

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
