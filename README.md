# LicenseOps

A fast CLI tool to check, fix, and migrate license headers across 50+ languages. Supports SPDX, REUSE, Apache/GPL boilerplates, and custom templates. Built for CI pipelines, pre-commit hooks, and local development.

## Features

- **Check & Fix** — validate headers or auto-add/replace them in-place
- **Multiple formats** — SPDX short (1-line and 2-line), REUSE, Apache 2.0 boilerplate, GPL/LGPL/AGPL boilerplate, or custom templates
- **50+ languages** — correct comment syntax for Go, Rust, Python, JavaScript, TypeScript, Java, C/C++, Shell, YAML, CSS, HTML, SQL, and more
- **SPDX expressions** — supports `AND`, `OR`, `WITH` operators (`Apache-2.0 OR MIT`)
- **Smart handling** — preserves shebangs, Python encoding declarations, skips generated files and binaries
- **Gitignore-aware** — respects `.gitignore` patterns automatically
- **Config file + CLI flags** — commit `.licenseops.yaml` to your repo, override with flags
- **Cross-format migration** — switch from one header format to another without manual cleanup
- **Zero config viable** — works with just `licenseops check -l MIT -o "Your Name" .`

## Installation

### Binary (recommended)

Download from [Releases](https://github.com/chalindu/licenseops/releases):

```bash
curl -sSL https://github.com/chalindu/licenseops/releases/latest/download/licenser_Linux_x86_64.tar.gz | tar xz
sudo mv licenseops /usr/local/bin/
```

### Go install

```bash
go install github.com/chalindu/licenseops/cmd/licenseops@latest
```

### Docker

```bash
docker run --rm -v "$PWD":/src -w /src ghcr.io/chalindu/licenseops check -l MIT -o "Your Name"
```

## Quick Start

### Check compliance

```bash
licenseops check -l Apache-2.0 -o "Acme Corp" .
```

### Fix headers

```bash
licenseops fix -l Apache-2.0 -o "Acme Corp" .
```

### Use a config file

Create `.licenseops.yaml` in your project root:

```yaml
license: Apache-2.0
copyright-holder: "Acme Corp"
exclude:
  - "vendor/**"
  - "**/*.pb.go"
```

Then simply run:

```bash
licenseops check
licenseops fix
```

## Header Formats

### `spdx` (default)

**2-line mode** (when `copyright-holder` is set):
```
// Copyright 2026 Acme Corp
// SPDX-License-Identifier: Apache-2.0
```

**1-line mode** (when `copyright-holder` is omitted):
```
// SPDX-License-Identifier: MIT
```

### `reuse`

[FSFE REUSE](https://reuse.software/) specification:
```
// SPDX-FileCopyrightText: 2026 Acme Corp
// SPDX-License-Identifier: Apache-2.0
```

### `apache-long`

Full Apache License 2.0 boilerplate (requires `license: Apache-2.0`):
```
// Copyright 2026 Acme Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// ...
```

### `gpl-long`

GNU GPL/LGPL/AGPL boilerplate (requires a GPL-family license):
```
// Copyright 2026 Acme Corp
//
// This program is free software: ...
```

### `custom`

User-defined Go template:
```yaml
format: custom
header-template: "headers/my-header.tmpl"
```

Template variables: `{{.Year}}`, `{{.Holder}}`, `{{.License}}`, `{{.Comment}}`, `{{.BlockStart}}`, `{{.BlockEnd}}`

## Configuration

### `.licenseops.yaml`

```yaml
# Required
license: Apache-2.0
copyright-holder: "Acme Corp"

# Optional
year: 2026                     # defaults to current year
format: spdx                   # spdx | reuse | apache-long | gpl-long | custom
header-template: ""            # path to template file (for format: custom)

paths:                         # directories to scan (default: ["."])
  - "."

exclude:                       # glob patterns to exclude
  - "vendor/**"
  - "**/*_generated.go"
  - "**/testdata/**"

skip-generated: true           # skip files with "DO NOT EDIT" markers
gitignore: true                # respect .gitignore patterns
```

### CLI Flags

```
-c, --config <path>       config file path (default: .licenseops.yaml)
-l, --license <spdx-id>   SPDX license identifier or expression
-o, --owner <holder>      copyright holder
-y, --year <year>         copyright year
    --format <format>     header format (spdx, reuse, apache-long, gpl-long, custom)
-v, --verbose             show status of every file
    --dry-run             show what would change (fix command only)
```

**Precedence:** CLI flags > config file > defaults

### SPDX Expressions

The `license` field supports full SPDX expressions:

```bash
licenseops check -l "Apache-2.0 OR MIT" -o "Acme Corp" .
licenseops check -l "GPL-3.0-only WITH Classpath-exception-2.0" -o "Acme Corp" .
```

## CI Integration

### GitHub Actions

```yaml
- name: Check license headers
  run: |
    curl -sSL https://github.com/chalindu/licenseops/releases/latest/download/licenser_Linux_x86_64.tar.gz | tar xz
    ./licenseops check
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0    | All files compliant / all files fixed |
| 1    | Non-compliant files found (check mode) |
| 2    | Runtime error (bad config, IO error) |

### Pre-commit

```yaml
repos:
  - repo: https://github.com/chalindu/licenseops
    rev: v0.2.0
    hooks:
      - id: licenseops
        args: ["check"]
```

## Supported Languages

| Comment Style | Extensions |
|---|---|
| `//` | `.go` `.rs` `.java` `.js` `.jsx` `.ts` `.tsx` `.c` `.h` `.cpp` `.cc` `.cs` `.swift` `.kt` `.scala` `.dart` `.proto` `.zig` |
| `#` | `.py` `.rb` `.sh` `.bash` `.pl` `.yaml` `.yml` `.toml` `.tf` `.hcl` `.r` `.ex` `.nix` `Dockerfile` `Makefile` |
| `--` | `.hs` `.lua` `.sql` `.ada` `.elm` |
| `/* */` | `.css` `.scss` `.less` |
| `<!-- -->` | `.html` `.xml` `.svg` `.vue` |

## Development

```bash
make build       # Build binary
make test        # Run tests
make lint        # Run linter
make lint-fix    # Run linter with auto-fix
make run         # Build and run with default args
make docker      # Build Docker image
make clean       # Remove build artifacts
```

## License

Apache-2.0
