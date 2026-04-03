# Use Cases

Real-world scenarios and how to set up licenseops for each.

---

## 1. Open-Source Go Project

A typical Go project using Apache-2.0 with standard layout.

**.licenseops.yaml:**
```yaml
license: Apache-2.0
copyright-holder: "The MyProject Authors"

exclude:
  - "vendor/**"
  - "**/*.pb.go"
  - "**/zz_generated.*"
```

**Usage:**
```bash
# Check in CI
licenseops check

# Fix locally before committing
licenseops fix
```

**Result:**
```go
// Copyright 2026 The MyProject Authors
// SPDX-License-Identifier: Apache-2.0

package main
```

---

## 2. JavaScript/TypeScript Monorepo

A Node.js monorepo with MIT license, multiple packages, and build output to exclude.

**.licenseops.yaml:**
```yaml
license: MIT
copyright-holder: "Acme Corp"

paths:
  - "packages/"
  - "apps/"

exclude:
  - "node_modules/**"
  - "**/dist/**"
  - "**/build/**"
  - "**/*.min.js"
  - "**/*.d.ts"
  - "**/coverage/**"
```

**Result:**
```typescript
// Copyright 2026 Acme Corp
// SPDX-License-Identifier: MIT

export function hello(): string {
  return "world";
}
```

---

## 3. Python Package

A Python project with shebang scripts and encoding declarations.

**.licenseops.yaml:**
```yaml
license: BSD-3-Clause
copyright-holder: "University Research Lab"

exclude:
  - "**/__pycache__/**"
  - "*.egg-info/**"
  - "venv/**"
  - ".venv/**"
```

**Result for a normal file:**
```python
# Copyright 2026 University Research Lab
# SPDX-License-Identifier: BSD-3-Clause

import os
```

**Result for a script with shebang:**
```python
#!/usr/bin/env python3

# Copyright 2026 University Research Lab
# SPDX-License-Identifier: BSD-3-Clause

def main():
    pass
```

**Result with encoding declaration:**
```python
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# Copyright 2026 University Research Lab
# SPDX-License-Identifier: BSD-3-Clause

def main():
    pass
```

LicenseOps preserves shebangs and encoding declarations at the top automatically.

---

## 4. Rust Project with Dual License

Many Rust projects use dual Apache-2.0/MIT licensing.

**.licenseops.yaml:**
```yaml
license: "Apache-2.0 OR MIT"
copyright-holder: "The Rust Project Contributors"

exclude:
  - "target/**"
```

**Result:**
```rust
// Copyright 2026 The Rust Project Contributors
// SPDX-License-Identifier: Apache-2.0 OR MIT

fn main() {
    println!("Hello, world!");
}
```

---

## 5. SPDX-Only Headers (No Copyright Line)

Some projects (like the Linux kernel) use only the SPDX identifier line without a copyright line.

**.licenseops.yaml:**
```yaml
license: GPL-2.0-only
# copyright-holder intentionally omitted → 1-line mode
```

**Result for Go:**
```go
// SPDX-License-Identifier: GPL-2.0-only

package main
```

**Result for CSS:**
```css
/* SPDX-License-Identifier: GPL-2.0-only */

body { margin: 0; }
```

---

## 6. FSFE REUSE Compliant Project

Projects following the [REUSE specification](https://reuse.software/) use `SPDX-FileCopyrightText` tags.

**.licenseops.yaml:**
```yaml
license: GPL-3.0-or-later
copyright-holder: "Jane Doe <jane@example.com>"
format: reuse
```

**Result:**
```python
# SPDX-FileCopyrightText: 2026 Jane Doe <jane@example.com>
# SPDX-License-Identifier: GPL-3.0-or-later

def hello():
    pass
```

---

## 7. Apache Foundation Style Project

Projects that follow the Apache Foundation's guidelines with the full boilerplate.

**.licenseops.yaml:**
```yaml
license: Apache-2.0
copyright-holder: "The Apache Software Foundation"
format: apache-long
```

**Result:**
```java
// Copyright 2026 The Apache Software Foundation
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

public class App {
}
```

---

## 8. GPL Project with Full Boilerplate

**.licenseops.yaml:**
```yaml
license: GPL-3.0-or-later
copyright-holder: "Free Software Project"
format: gpl-long
```

**Result:**
```c
// Copyright 2026 Free Software Project
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

#include <stdio.h>
```

---

## 9. Custom Header Template

For organizations with non-standard header requirements.

**headers/company.tmpl:**
```
{{.Comment}} Copyright (c) {{.Year}} {{.Holder}}. All rights reserved.
{{.Comment}} Use of this source code is governed by the {{.License}} license
{{.Comment}} that can be found in the LICENSE file in the root of this repository.
{{.Comment}}
{{.Comment}} NOTICE: This software is proprietary and confidential.
```

**.licenseops.yaml:**
```yaml
license: MIT
copyright-holder: "MegaCorp Inc."
format: custom
header-template: "headers/company.tmpl"
```

**Result:**
```go
// Copyright (c) 2026 MegaCorp Inc. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file in the root of this repository.
//
// NOTICE: This software is proprietary and confidential.

package main
```

---

## 10. Migrating Between Formats

Switching from Apache long boilerplate to SPDX short headers across a codebase.

**Before** (files have Apache long headers):
```go
// Copyright 2024 Old Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// ...13 more lines...

package main
```

**Command:**
```bash
licenseops fix -l Apache-2.0 -o "New Corp" -f spdx .
```

**After:**
```go
// Copyright 2026 New Corp
// SPDX-License-Identifier: Apache-2.0

package main
```

LicenseOps detects the old format and fully replaces it with the new one — no manual cleanup needed.

---

## 11. Multi-Language Project

A project with Go, Python, Shell, SQL, and CSS files. No special config needed — licenseops auto-detects the correct comment style for each file type.

**.licenseops.yaml:**
```yaml
license: Apache-2.0
copyright-holder: "DevTeam"

exclude:
  - "vendor/**"
  - "node_modules/**"
  - "**/*.min.*"
```

**Results across languages:**

| File | Header Style |
|------|-------------|
| `main.go` | `// Copyright 2026 DevTeam` |
| `app.py` | `# Copyright 2026 DevTeam` |
| `deploy.sh` | Shebang preserved, then `# Copyright 2026 DevTeam` |
| `query.sql` | `-- Copyright 2026 DevTeam` |
| `style.css` | `/* Copyright 2026 DevTeam ... */` (block comment) |
| `index.html` | `<!-- Copyright 2026 DevTeam ... -->` |
| `config.yaml` | `# Copyright 2026 DevTeam` |
| `data.json` | Skipped (no comment syntax) |

---

## 12. CI-Only Usage (No Config File)

For simple projects or one-off checks, use CLI flags only — no config file needed.

```bash
# Check
licenseops check -l MIT -o "My Name" .

# Fix
licenseops fix -l MIT -o "My Name" .

# Check with verbose output
licenseops check -l MIT -o "My Name" -v .

# Preview what fix would change
licenseops fix -l MIT -o "My Name" --dry-run .
```
