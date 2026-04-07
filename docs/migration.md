# Cross-Format Migration Guide

This guide explains how to migrate license headers from one format to another using `lops`.

## Overview

When you switch header formats (e.g., from Apache boilerplate to SPDX short), `lops fix` handles the entire migration automatically:

1. **Detect** — Every file is checked against all known formats (SPDX, REUSE, Apache-long, GPL-long, custom). The format whose header occupies the most space wins, ensuring the old header is fully identified regardless of which format was originally used.
2. **Strip** — The detected old header is completely removed. Shebangs (`#!/...`) and Python encoding declarations (`# -*- coding: utf-8 -*-`) are preserved.
3. **Add** — A new header is generated in the target format and inserted at the top of the file (after any shebang/encoding lines).

This works for any combination of source and target formats — you don't need to tell `lops` what the old format was.

## Step-by-Step Migration

### 1. Preview what will change

Always preview before migrating. Use `--dry-run` to see which files will be affected, or `--diff` to see the exact changes:

```bash
# See which files will be modified
lops fix -f spdx -l Apache-2.0 -o "Acme Corp" --dry-run .

# See unified diffs of every change
lops fix -f spdx -l Apache-2.0 -o "Acme Corp" --diff .
```

### 2. Run the migration

```bash
lops fix -f spdx -l Apache-2.0 -o "Acme Corp" .
```

Or update your `.licenseops.yaml` and run:

```bash
lops fix
```

### 3. Verify

```bash
lops check
```

All files should now pass. If you want to inspect every file:

```bash
lops check -v
```

## Common Migration Scenarios

### Apache boilerplate to SPDX short

**Before** (14-line Apache boilerplate in every file):
```go
// Copyright 2024 Old Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main
```

**Config change:**
```yaml
format: spdx           # was: apache-long (or omit — spdx is the default)
license: Apache-2.0
copyright-holder: "New Corp"
```

**Command:**
```bash
lops fix
```

**After** (2-line SPDX header):
```go
// Copyright 2026 New Corp
// SPDX-License-Identifier: Apache-2.0

package main
```

### GPL boilerplate to REUSE

**Config change:**
```yaml
format: reuse           # was: gpl-long
license: GPL-3.0-only
copyright-holder: "My Project"
```

**Command:**
```bash
lops fix
```

**After:**
```go
// SPDX-FileCopyrightText: 2026 My Project
// SPDX-License-Identifier: GPL-3.0-only

package main
```

### SPDX to Apache boilerplate

**Config change:**
```yaml
format: apache-long     # was: spdx
license: Apache-2.0
copyright-holder: "Acme Corp"
```

**Command:**
```bash
lops fix
```

The 2-line SPDX header is removed and replaced with the full 14-line Apache boilerplate.

### Changing license and format at the same time

You can change both the license and the format in one step:

```yaml
format: reuse           # was: apache-long
license: MIT            # was: Apache-2.0
copyright-holder: "New Owner"
```

```bash
lops fix
```

The old Apache boilerplate is stripped and replaced with a REUSE header for MIT.

## Using CLI Flags vs Config File

Both approaches work. CLI flags override the config file:

```bash
# Pure CLI — no config file needed
lops fix -f reuse -l MIT -o "Acme Corp" .

# Config file with a one-time format override
lops fix -f reuse

# Config file only
lops fix
```

For a permanent migration, update `.licenseops.yaml` so future `lops check` runs validate against the new format.

## Mixed-Format Codebases

If different files have different old formats (e.g., some have Apache, some have REUSE, some have no header), `lops fix` handles all of them in one run. Each file is independently detected and migrated:

```bash
# All three types of files are handled:
#   - Files with Apache headers → stripped and replaced
#   - Files with REUSE headers → stripped and replaced
#   - Files with no header → new header added
lops fix -f spdx -l MIT -o "Acme Corp" .
```

## Removing All Headers

The `remove` command strips headers from all files using the same multi-format detection. It works regardless of which format is in your config:

```bash
# Preview
lops remove --dry-run

# Execute
lops remove
```

Even if your config says `format: spdx`, `remove` will detect and strip Apache, GPL, REUSE, custom, and SPDX headers.

### Cleaning Up Excluded Files

`lops remove` normally respects the `exclude:` list — excluded files are
skipped just like with `check` and `fix`. After adding new entries to your
exclude list, those files keep any headers they already had. Use
`--excluded-only` to invert the scan and process only those files:

```bash
# Preview which excluded files still carry a header
lops remove --excluded-only --dry-run

# Strip the headers
lops remove --excluded-only
```

In this mode the scanner walks the entire tree but returns only files that
match a pattern from your config's `exclude:` block. Built-in defaults
(`.git/**`, `vendor/**`, `node_modules/**`, `third_party/**`,
`.licenseops.yaml`) and gitignore filtering are bypassed so excluded files
inside ignored directories remain reachable.

Typical workflow when adding excludes for files that shouldn't carry headers
(e.g. workflow YAML, `Dockerfile`, `.golangci.yml`):

```bash
# 1. Add entries to .licenseops.yaml exclude: list
# 2. Strip leftover headers from those files
lops remove --excluded-only

# 3. Verify the rest of the repo still passes
lops check
```

## What Gets Preserved

During migration, `lops` preserves:

- **Shebangs** (`#!/usr/bin/env python3`) — stays at line 1
- **Python encoding declarations** (`# -*- coding: utf-8 -*-`) — stays after shebang
- **All code below the header** — untouched

Example with shebang:

**Before:**
```python
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# Copyright 2024 Old Corp
# SPDX-License-Identifier: MIT

import os
```

**After** (migrated to REUSE):
```python
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# SPDX-FileCopyrightText: 2026 New Corp
# SPDX-License-Identifier: Apache-2.0

import os
```

## Comment Style Adaptation

Each format automatically adapts to the file's comment syntax. During migration, the new header uses the correct comment style for each file type:

| File type | Comment style |
|-----------|--------------|
| `.go`, `.java`, `.rs`, `.ts` | `// ...` |
| `.py`, `.rb`, `.sh`, `.yaml` | `# ...` |
| `.sql`, `.hs`, `.lua` | `-- ...` |
| `.css`, `.scss` | `/* ... */` |
| `.html`, `.xml`, `.vue` | `<!-- ... -->` |

## Idempotency

Running `lops fix` multiple times is safe. After the first run migrates the headers, subsequent runs detect the new header as valid and make no changes:

```bash
lops fix    # migrates headers
lops fix    # no changes — already compliant
lops fix    # still no changes
```
