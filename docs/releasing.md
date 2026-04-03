# Releasing

How the CI/CD pipeline handles builds, image publishing, and releases.

## What Happens on Each Commit

### Pull Requests

Every PR triggers the **CI workflow** (`ci.yml`):

| Job | What it does |
|-----|--------------|
| `build` | Builds the binary, runs `go vet`, runs tests with race detector |
| `lint` | Runs `golangci-lint` |
| `docker` | Builds the Docker image and verifies it starts (`--version`) |

No images are pushed. No artifacts are published. This is validation only.

### Merge to `main`

When a PR is merged (push to `main`), the same CI jobs run, plus:

| Job | What it does |
|-----|--------------|
| `docker-latest` | Builds and pushes `ghcr.io/chalindukodikara/licenseops:latest` |

This means the `:latest` Docker tag **always tracks the `main` branch**. Every merge to `main` updates it.

```bash
# Always gets the latest main build
docker pull ghcr.io/chalindukodikara/licenseops:latest
```

## Creating a Release

Releases are cut from a **release branch** and triggered by pushing a Git tag.

### Steps

1. **Make sure `main` is clean and CI passes.**

2. **Create a release branch from `main`:**
   ```bash
   git checkout main
   git pull origin main
   git checkout -b release/v0.1.0
   ```

3. **Prepare the release on the branch:**
   - Run the full pre-release checklist (see below)
   - Make any last-minute fixes (version bumps, changelog updates)
   - Push the branch:
   ```bash
   git push origin release/v0.1.0
   ```

4. **Open a PR from the release branch to `main`** for final review.
   This ensures CI runs against the exact code that will be released.

5. **Merge the PR**, then tag from `main`:**
   ```bash
   git checkout main
   git pull origin main
   git tag v0.1.0
   git push origin v0.1.0
   ```

6. **The release workflow runs automatically and creates a draft release:**

   | Job | What it does |
   |-----|--------------|
   | `goreleaser` | Builds binaries, creates a **draft** GitHub Release with auto-generated changelog |
   | `docker` | Builds and pushes versioned Docker images to GHCR |

7. **Review the draft release on GitHub:**
   - Go to [Releases](https://github.com/chalindukodikara/licenseops/releases)
   - The draft will be at the top with a "Draft" badge
   - Review the auto-generated changelog — it groups commits into **Features**, **Bug Fixes**, and **Other**
   - Edit the release notes if needed (add summary, highlights, breaking changes, upgrade instructions)
   - Verify the binary assets are attached

8. **Publish the release:**
   - Click **"Publish release"** to make it public
   - This makes the binaries downloadable and the release visible to users

### What Gets Created

**GitHub Release (draft)** — via GoReleaser:
- `licenseops_Linux_x86_64.tar.gz`
- `licenseops_Linux_arm64.tar.gz`
- `licenseops_Darwin_x86_64.tar.gz`
- `licenseops_Darwin_arm64.tar.gz`
- `checksums.txt`
- Auto-generated changelog grouped by type:
  - **Features** — commits starting with `feat:`
  - **Bug Fixes** — commits starting with `fix:`
  - **Other** — everything else (excludes `docs:`, `test:`, `ci:`, `chore:`)

**Docker images** (pushed to GHCR):
- `ghcr.io/chalindukodikara/licenseops:0.1.0` — exact version
- `ghcr.io/chalindukodikara/licenseops:0.1` — major.minor (moves with patch releases)
- `ghcr.io/chalindukodikara/licenseops:latest` — latest release

### Commit Message Convention

Use [Conventional Commits](https://www.conventionalcommits.org/) for clean changelogs:

```
feat: add REUSE format support
fix: preserve shebang in Python files
docs: add CI integration guide
test: add header stripping tests
ci: update golangci-lint to v2
chore: update dependencies
```

Commits prefixed with `docs:`, `test:`, `ci:`, and `chore:` are excluded from the release changelog.

### Version Convention

Follow [Semantic Versioning](https://semver.org/):

```
v0.x.y  — pre-1.0 development (breaking changes allowed)
v1.0.0  — first stable release
v1.x.y  — backwards-compatible changes after 1.0
```

Tag format must be `v` followed by semver: `v0.1.0`, `v0.2.0`, `v1.0.0`.

## Docker Tag Summary

| Tag | Updated when | Use for |
|-----|-------------|---------|
| `:latest` | Every merge to `main` | Development, always up-to-date |
| `:0.1.0` | Release `v0.1.0` tag pushed | Production, pinned to exact version |
| `:0.1` | Release `v0.1.x` tag pushed | Production, gets patch updates |

## Pre-Release Checklist

Before tagging a release:

- [ ] All CI checks pass on `main`
- [ ] `make lint` passes locally
- [ ] `make test` passes locally
- [ ] `licenseops check` passes on the repo itself
- [ ] Update version references in docs if needed
- [ ] Verify `RELEASE.md` changelog is up to date

## Hotfix Process

For urgent fixes to a released version:

```bash
# Branch from the release tag
git checkout -b hotfix/v0.1.1 v0.1.0

# Make the fix, commit
git commit -m "fix: critical bug"

# Push the hotfix branch and open a PR to main
git push origin hotfix/v0.1.1

# After PR is reviewed and merged to main, tag from main
git checkout main
git pull origin main
git tag v0.1.1
git push origin v0.1.1
```

The release workflow triggers on any `v*` tag, regardless of branch.

## Branch Summary

| Branch | Purpose | Lifetime |
|--------|---------|----------|
| `main` | Stable development branch, all releases are tagged here | Permanent |
| `release/v0.x.0` | Release preparation (final fixes, changelog) | Merged to `main`, then deleted |
| `hotfix/v0.x.y` | Urgent patches to a released version | Merged to `main`, then deleted |
| `feature/*` | Feature development | Merged to `main` via PR |
