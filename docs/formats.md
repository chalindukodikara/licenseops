# Header Format Comparison

LicenseOps supports 5 header formats. This guide helps you choose the right one.

## Quick Comparison

| Format | Lines | Use When |
|--------|-------|----------|
| `spdx` (2-line) | 2 | Default choice. Clean, machine-readable, widely recognized. |
| `spdx` (1-line) | 1 | SPDX tag without per-file copyright (e.g. Linux kernel style). |
| `reuse` | 2 | FSFE REUSE compliance required. |
| `apache-long` | 14 | Apache Foundation projects or projects that require the full boilerplate. |
| `gpl-long` | 14 | FSF/GNU projects that require the full GPL/LGPL/AGPL boilerplate. |
| `custom` | Varies | Organization-specific headers that don't fit any built-in format. |

## Side-by-Side Examples

All examples below use Go (`//` comments), `Apache-2.0` license, copyright holder `Acme Corp`.

### `spdx` (2-line, default)

```
// Copyright 2026 Acme Corp
// SPDX-License-Identifier: Apache-2.0
```

- **Pros**: Short, clean, machine-parseable by all SPDX-aware tools
- **Cons**: Doesn't include license terms — relies on a root LICENSE file
- **Config**: `format: spdx` (or omit — it's the default)

### `spdx` (1-line)

```
// SPDX-License-Identifier: Apache-2.0
```

- **Pros**: Minimal footprint, one line per file
- **Cons**: No per-file copyright attribution
- **Config**: `format: spdx` with `copyright-holder` omitted

### `reuse`

```
// SPDX-FileCopyrightText: 2026 Acme Corp
// SPDX-License-Identifier: Apache-2.0
```

- **Pros**: FSFE REUSE standard, explicitly machine-readable copyright line
- **Cons**: Less common outside European open-source projects
- **Config**: `format: reuse`

### `apache-long`

```
// Copyright 2026 Acme Corp
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
```

- **Pros**: Legally self-contained, no dependency on root LICENSE file
- **Cons**: 14 lines of boilerplate in every file
- **Config**: `format: apache-long` (requires `license: Apache-2.0`)

### `gpl-long`

```
// Copyright 2026 Acme Corp
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
```

- **Pros**: FSF-recommended, legally self-contained
- **Cons**: Long, only valid for GPL-family licenses
- **Config**: `format: gpl-long` (requires a GPL/LGPL/AGPL license)
- **Variants**: Automatically selects GPL, LGPL, or AGPL wording based on the license ID

## Supported Licenses per Format

| Format | Valid Licenses |
|--------|---------------|
| `spdx` | Any SPDX ID or expression |
| `reuse` | Any SPDX ID or expression |
| `apache-long` | `Apache-2.0` only |
| `gpl-long` | `GPL-2.0-only`, `GPL-2.0-or-later`, `GPL-3.0-only`, `GPL-3.0-or-later`, `LGPL-*`, `AGPL-*` |
| `custom` | Any SPDX ID or expression |

## Comment Style by Language

Each format adapts to the file's comment syntax automatically.

**Line comment languages** (`//`, `#`, `--`):

Headers are written with the line comment prefix on each line.

**Block comment only languages** (CSS `/* */`, HTML `<!-- -->`):

Headers are wrapped in a block comment:
```css
/*
 Copyright 2026 Acme Corp
 SPDX-License-Identifier: Apache-2.0
*/
```

For SPDX 1-line mode, a single-line block comment is used:
```css
/* SPDX-License-Identifier: Apache-2.0 */
```

## Migration Between Formats

LicenseOps handles format migration automatically. When you run `fix` with a different format than what's currently in the files, it:

1. Detects and strips the old header (any known format)
2. Inserts the new header in the target format

```bash
# Files currently have apache-long headers
# Migrate to SPDX short:
lops fix -f spdx -l Apache-2.0 -o "Acme Corp" .
```

No manual cleanup needed — the old boilerplate is fully removed and replaced.
