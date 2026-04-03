# Configuration Guide

LicenseOps can be configured via a YAML config file, CLI flags, or a combination of both. CLI flags always take precedence over the config file.

## Config File

By default, lops looks for `.licenseops.yaml` in the current directory. Use `-c` to specify a different path.

### Minimal Config

```yaml
license: MIT
copyright-holder: "Jane Doe"
```

### Full Config

```yaml
# Required
license: Apache-2.0
copyright-holder: "Acme Corp"

# Optional — defaults shown
year: 2026                     # defaults to current year
format: spdx                   # spdx | reuse | apache-long | gpl-long | custom
header-template: ""            # only used with format: custom

# File selection
paths:                         # directories/files to scan (default: ["."])
  - "."
  - "libs/"

# Exclusion patterns (doublestar glob syntax)
exclude:
  - "vendor/**"
  - "node_modules/**"
  - "**/*.pb.go"
  - "**/*_generated.*"
  - "**/testdata/**"

# Behavior
skip-generated: true           # skip files with "DO NOT EDIT" / "@generated" markers
gitignore: true                # respect .gitignore patterns
```

## CLI Flags

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--license` | `-l` | SPDX license ID or expression | `-l MIT`, `-l "Apache-2.0 OR MIT"` |
| `--owner` | `-o` | Copyright holder | `-o "Acme Corp"` |
| `--format` | `-f` | Header format | `-f reuse` |
| `--year` | `-y` | Copyright year | `-y 2025` |
| `--config` | `-c` | Config file path | `-c .licenseops.yaml` |
| `--verbose` | `-v` | Show every file | `-v` |
| `--dry-run` | | Preview changes (fix only) | `--dry-run` |

### Precedence

```
CLI flags  >  config file  >  defaults
```

If both a config file and CLI flags are provided, flags override config values.

## Exclude Patterns

LicenseOps uses [doublestar](https://github.com/bmatcuk/doublestar) glob syntax for exclusion patterns.

### Pattern Syntax

| Pattern | Matches |
|---------|---------|
| `*.py` | Python files in the root directory only |
| `**/*.py` | Python files in any directory |
| `vendor/**` | Everything under the vendor directory |
| `**/*_test.go` | All Go test files |
| `src/**/*.{js,ts}` | JS and TS files under src/ |
| `**/generated_*.go` | Files starting with `generated_` anywhere |
| `docs/**/*.md` | Markdown files under docs/ |

### Excluding by File Type

```yaml
exclude:
  - "**/*.py"           # all Python files
  - "**/*.java"         # all Java files
  - "**/*.{js,jsx}"     # all JS and JSX files
```

### Excluding Directories

```yaml
exclude:
  - "vendor/**"
  - "node_modules/**"
  - "third_party/**"
  - "build/**"
  - "dist/**"
  - ".git/**"
```

### Excluding Generated Code

```yaml
exclude:
  - "**/*.pb.go"           # protobuf generated Go
  - "**/*_generated.go"    # codegen output
  - "**/*.gen.ts"          # generated TypeScript
  - "**/generated/**"      # entire generated directory
```

In addition to exclude patterns, setting `skip-generated: true` (the default) will automatically skip files containing `Code generated ... DO NOT EDIT` or `@generated` markers.

### Default Excludes

These are always excluded even without config:

- `vendor/**`
- `node_modules/**`
- `.git/**`
- `third_party/**`
- `.licenseops.yaml`

User-defined exclude patterns are added on top of these defaults.

## SPDX License Expressions

The `license` field accepts full SPDX expressions, not just single IDs.

### Single License

```yaml
license: MIT
```

### Dual License (choice)

```yaml
license: "Apache-2.0 OR MIT"
```

Produces: `// SPDX-License-Identifier: Apache-2.0 OR MIT`

### Multiple Licenses (all apply)

```yaml
license: "Apache-2.0 AND MIT"
```

### Complex Expressions

```yaml
license: "Apache-2.0 AND (MIT OR GPL-2.0-only)"
```

### License with Exception

```yaml
license: "GPL-3.0-only WITH Classpath-exception-2.0"
```

### Or-Later (non-GNU)

```yaml
license: "EPL-1.0+"
```

### GNU Convention

GNU licenses use `-only` / `-or-later` suffixes:

```yaml
license: GPL-3.0-only       # exactly v3.0
license: GPL-3.0-or-later   # v3.0 or any later version
```

Using bare `GPL-2.0` or `GPL-3.0` will trigger a deprecation warning.

## Format-Specific Config

### SPDX 1-Line Mode

Omit `copyright-holder` to use SPDX-only headers:

```yaml
license: MIT
# no copyright-holder → 1-line mode
```

Produces: `// SPDX-License-Identifier: MIT`

### Apache Long Format

```yaml
license: Apache-2.0
copyright-holder: "Acme Corp"
format: apache-long
```

Requires `license: Apache-2.0` — any other license will error.

### GPL Long Format

```yaml
license: GPL-3.0-only
copyright-holder: "Free Soft Corp"
format: gpl-long
```

Requires a GPL/LGPL/AGPL license ID.

### Custom Template

```yaml
license: MIT
copyright-holder: "Acme Corp"
format: custom
header-template: "headers/my-header.tmpl"
```

See [custom-templates.md](custom-templates.md) for template syntax.
