# Configuration Guide

LicenseOps can be configured via a YAML config file, CLI flags, or a combination of both. CLI flags always take precedence over the config file.

## Config File

By default, `lops` looks for `.licenseops.yaml` in the current working directory. Use `-c` to specify a different path.

### Config File Resolution

| Scenario | Behavior |
|----------|----------|
| `.licenseops.yaml` exists in cwd | Loaded automatically — no flags needed |
| `.licenseops.yaml` doesn't exist | Silently falls back to built-in defaults, no error |
| `-c path/to/config.yaml` | Uses that file — errors if not found |
| `-c /dev/null` | Empty config — only CLI flags and defaults apply |

This means `lops` works in two modes:
- **With config**: commit `.licenseops.yaml` to your repo, run `lops check` / `lops fix` with no flags
- **Without config**: pass everything via flags — `lops check -l MIT -o "Your Name" .`

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

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--license` | `-l` | *none* | SPDX license ID or expression |
| `--owner` | `-o` | *none* | Copyright holder |
| `--format` | `-f` | `spdx` | Header format |
| `--year` | `-y` | Current year | Copyright year |
| `--config` | `-c` | `.licenseops.yaml` | Config file path |
| `--verbose` | `-v` | `false` | Show every file |
| `--dry-run` | | `false` | Preview changes (fix only) |

No flag is required on the command line — all values can be provided via the config file instead. The tool validates that required **values** are present (from either source) before running.

### Defaults and Fallbacks

Every setting is resolved through a three-layer chain:

```
CLI flag  →  config file  →  built-in default
```

If a CLI flag is set, it wins. Otherwise the config file value is used. If neither provides a value, the built-in default applies. If there is no default and no value is provided, validation errors with a clear message.

**Detailed default values:**

| Setting | Built-in Default | Notes |
|---------|-----------------|-------|
| `license` | *none* — required | Must be set via config or `-l` flag. Errors if missing. |
| `copyright-holder` | *none* — optional | If omitted, SPDX format uses 1-line mode (no copyright line). Required for `reuse`, `apache-long`, `gpl-long` formats. |
| `format` | `spdx` | Can be overridden in config or with `-f` flag. |
| `year` | Current year (e.g. `2026`) | Auto-detected. Override with `-y` for a specific year. |
| `config` | `.licenseops.yaml` | Looked up in cwd. If the file doesn't exist, silently skipped — no error. |
| `paths` | `["."]` | Scans current directory. Config can set multiple paths. CLI args override entirely. |
| `exclude` | `vendor/**`, `node_modules/**`, `.git/**`, `third_party/**`, `.licenseops.yaml` | Config patterns are **appended** to these defaults, not replaced. |
| `skip-generated` | `true` | Skips files with `DO NOT EDIT` / `@generated` markers. |
| `gitignore` | `true` | Respects `.gitignore` patterns from the project root. |
| `header-template` | *none* | Only required when `format: custom`. |

### Required vs Optional by Format

| Setting | `spdx` | `reuse` | `apache-long` | `gpl-long` | `custom` |
|---------|--------|---------|---------------|------------|----------|
| `license` | Required | Required | Must be `Apache-2.0` | Must be GPL/LGPL/AGPL | Required |
| `copyright-holder` | Optional (omit for 1-line) | Required | Required | Required | Optional |
| `header-template` | — | — | — | — | Required |

If a required field is missing, `lops` exits with code 2 and a message like:

```
Error: license is required (set via config or --license flag)
Error: copyright-holder is required for 'reuse' format
Error: 'apache-long' format requires license: Apache-2.0 (got "MIT")
```

### Flag Behavior Details

**`-o ""` (explicit empty owner):** Clears the copyright holder even if the config file sets one. This forces SPDX 1-line mode.

```bash
# Config has copyright-holder, but you want 1-line mode for this run:
lops fix -l MIT -o "" .
```

**`-l` (license):** Accepts single IDs (`MIT`) or expressions (`Apache-2.0 OR MIT`). Validated against the SPDX license list. Deprecated GNU IDs (`GPL-2.0`) trigger a warning but still work.

**Paths (positional args):** If paths are passed as arguments, they **replace** the `paths` from config entirely (not appended).

```bash
# Config says paths: ["."], but only scan src/:
lops check src/

# Scan multiple specific paths:
lops check src/ libs/ cmd/
```

### Merge Rules Summary

| Setting | Merge behavior |
|---------|---------------|
| `license` | Flag replaces config |
| `copyright-holder` | Flag replaces config (including `-o ""` to clear) |
| `format` | Flag replaces config |
| `year` | Flag replaces config |
| `paths` | CLI args **replace** config (not appended) |
| `exclude` | Config **appends** to built-in defaults (not replaced) |
| `skip-generated` | Config replaces default |
| `gitignore` | Config replaces default |

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
