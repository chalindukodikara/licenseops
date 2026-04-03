# Custom Header Templates

When the built-in formats don't fit your needs, use `format: custom` with a Go template file.

## Setup

**.licenseops.yaml:**
```yaml
license: MIT
copyright-holder: "Acme Corp"
format: custom
header-template: "headers/header.tmpl"
```

## Template Variables

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `{{.Year}}` | Copyright year | `2026` |
| `{{.Holder}}` | Copyright holder | `Acme Corp` |
| `{{.License}}` | SPDX license ID | `MIT` |
| `{{.Comment}}` | Line comment prefix | `//`, `#`, `--` |
| `{{.BlockStart}}` | Block comment open | `/*`, `<!--` |
| `{{.BlockEnd}}` | Block comment close | `*/`, `-->` |

- For line-comment languages (Go, Python, etc.): `{{.Comment}}` is set, `{{.BlockStart}}`/`{{.BlockEnd}}` are empty.
- For block-comment-only languages (CSS, HTML): `{{.BlockStart}}`/`{{.BlockEnd}}` are set, `{{.Comment}}` is empty.

## Examples

### Standard Copyright + License Reference

**headers/standard.tmpl:**
```
{{.Comment}} Copyright (c) {{.Year}} {{.Holder}}. All rights reserved.
{{.Comment}} Use of this source code is governed by a {{.License}} license
{{.Comment}} that can be found in the LICENSE file.
```

**Output (Go):**
```go
// Copyright (c) 2026 Acme Corp. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main
```

**Output (Python):**
```python
# Copyright (c) 2026 Acme Corp. All rights reserved.
# Use of this source code is governed by a MIT license
# that can be found in the LICENSE file.

def main():
    pass
```

### SPDX + Proprietary Notice

**headers/proprietary.tmpl:**
```
{{.Comment}} Copyright {{.Year}} {{.Holder}}
{{.Comment}} SPDX-License-Identifier: {{.License}}
{{.Comment}}
{{.Comment}} CONFIDENTIAL - Do not distribute without written permission.
```

**Output:**
```go
// Copyright 2026 Acme Corp
// SPDX-License-Identifier: MIT
//
// CONFIDENTIAL - Do not distribute without written permission.

package main
```

### Block Comment Template (for CSS/HTML)

For languages that only support block comments, use `{{.BlockStart}}` and `{{.BlockEnd}}`:

**headers/block.tmpl:**
```
{{if .Comment}}{{.Comment}} Copyright {{.Year}} {{.Holder}}
{{.Comment}} SPDX-License-Identifier: {{.License}}{{else}}{{.BlockStart}}
 Copyright {{.Year}} {{.Holder}}
 SPDX-License-Identifier: {{.License}}
{{.BlockEnd}}{{end}}
```

This template uses Go's `{{if}}` to switch between line and block styles automatically.

### Minimal — Year and License Only

**headers/minimal.tmpl:**
```
{{.Comment}} {{.Year}} | {{.License}}
```

**Output:**
```go
// 2026 | MIT

package main
```

## How Check Works with Custom Templates

When running `licenseops check` with a custom template:

1. The template is rendered with `{{.Year}}` replaced by a wildcard pattern (`\d{4}`)
2. The rendered text is converted to a regex
3. The first N lines of each file are matched against this regex
4. If the match succeeds and is followed by a blank line, the file is compliant

This means:
- Year changes don't cause false failures (any 4-digit year is accepted)
- The rest of the header must match exactly (holder, license, text)

## How Fix Works with Custom Templates

When running `licenseops fix`:

1. Any existing header (in any known format) is stripped first
2. The template is rendered with the configured values
3. The rendered header is prepended to the file
4. Shebangs and encoding declarations are preserved

## Limitations

- Templates using complex Go template logic (loops, functions) may not detect correctly during `check`
- The `{{.Comment}}` variable is empty for block-comment-only languages — use `{{if .Comment}}` to handle both styles
- Template files must be valid Go `text/template` syntax
