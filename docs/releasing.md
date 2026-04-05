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
| `docker-latest-dev` | Builds and pushes `ghcr.io/chalindukodikara/licenseops:latest-dev` |

This means the `:latest-dev` Docker tag **always tracks the `main` branch**. Every merge to `main` updates it.

```bash
# Always gets the latest main build (may be unreleased)
docker pull ghcr.io/chalindukodikara/licenseops:latest-dev
```

## Creating a Release

### Prerequisites

- Push access to the repository
- Check [existing releases](https://github.com/chalindukodikara/licenseops/releases) for the latest version number

Export the environment variables for use in subsequent steps:

```bash
export MAJOR_VERSION=<MAJOR>
export MINOR_VERSION=<MINOR>
export PATCH_VERSION=<PATCH>
export GIT_REMOTE=origin  # remote name for github.com/chalindukodikara/licenseops
```

---

### Major/Minor Release (e.g. v0.2.0, v1.0.0)

Skip these steps if you are creating a patch release.

- [ ] Create the release branch from the latest `main`:
    ```bash
    git fetch ${GIT_REMOTE}
    git checkout -b release-v${MAJOR_VERSION}.${MINOR_VERSION} ${GIT_REMOTE}/main
    ```
- [ ] Run the [pre-release checklist](#pre-release-checklist)
- [ ] Make any release prep changes (doc updates, changelog):
    - For **major releases**: update version references across docs, README, and examples.
      Document breaking changes and add a migration guide if needed.
    ```bash
    git add -A
    git commit -m "chore: prepare release v${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}"
    ```
- [ ] Push the release branch:
    ```bash
    git push ${GIT_REMOTE} release-v${MAJOR_VERSION}.${MINOR_VERSION}
    ```
- [ ] Wait for [CI](https://github.com/chalindukodikara/licenseops/actions/workflows/ci.yml) to pass on the release branch.
- [ ] Proceed to [Tag the Release](#tag-the-release).

---

### Patch Release (e.g. v0.2.1)

Skip these steps if you are creating a major or minor release.

- [ ] Checkout the existing release branch and ensure it is up to date:
    ```bash
    git fetch ${GIT_REMOTE}
    git checkout release-v${MAJOR_VERSION}.${MINOR_VERSION}
    git pull ${GIT_REMOTE} release-v${MAJOR_VERSION}.${MINOR_VERSION}
    ```
- [ ] Apply the fix (cherry-pick from `main` or commit directly):
    ```bash
    git commit -m "fix: <description>"
    ```
- [ ] Run the [pre-release checklist](#pre-release-checklist)
- [ ] Push the changes to the release branch:
    ```bash
    git push ${GIT_REMOTE} release-v${MAJOR_VERSION}.${MINOR_VERSION}
    ```
- [ ] Wait for [CI](https://github.com/chalindukodikara/licenseops/actions/workflows/ci.yml) to pass on the release branch.
- [ ] Proceed to [Tag the Release](#tag-the-release).

---

### Tag the Release

- [ ] Create and push the tag:
    ```bash
    git tag v${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}
    git push ${GIT_REMOTE} v${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}
    ```
- [ ] Wait for [Release](https://github.com/chalindukodikara/licenseops/actions/workflows/release.yml) to pass — see [What Gets Created](#what-gets-created).
- [ ] Verify the [draft release](https://github.com/chalindukodikara/licenseops/releases) created by the workflow.
- [ ] Edit the release notes if needed (add summary, highlights, breaking changes, upgrade instructions).
    For **major releases**, add a **Breaking Changes** section and include migration instructions.
- [ ] Mark as **Latest** if this is the latest release. (If the current release is v0.1.1 while v0.2.0 exists, skip marking as latest.)
- [ ] Publish the release.

---

### What Gets Created

When the tag is pushed, the release workflow (`release.yml`) runs:

| Job | What it does |
|-----|--------------|
| `goreleaser` | Builds binaries, creates a **draft** GitHub Release with auto-generated changelog |
| `docker` | Builds and pushes versioned Docker images to GHCR |

**GitHub Release (draft)** — via GoReleaser:
- `lops_Linux_x86_64.tar.gz`
- `lops_Linux_arm64.tar.gz`
- `lops_Darwin_x86_64.tar.gz`
- `lops_Darwin_arm64.tar.gz`
- `checksums.txt`
- Auto-generated changelog grouped by type:
  - **Features** — commits starting with `feat:`
  - **Bug Fixes** — commits starting with `fix:`
  - **Other** — everything else (excludes `docs:`, `test:`, `ci:`, `chore:`)

**Docker images** (pushed to GHCR):
- `ghcr.io/chalindukodikara/licenseops:0.2.0` — exact version
- `ghcr.io/chalindukodikara/licenseops:0.2` — major.minor (moves with patch releases)
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
| `:latest` | Release tag pushed | Production, latest stable release |
| `:latest-dev` | Every merge to `main` | Development, always up-to-date (may be unreleased) |
| `:0.1.0` | Release `v0.1.0` tag pushed | Production, pinned to exact version |
| `:0.1` | Release `v0.1.x` tag pushed | Production, gets patch updates |

## Pre-Release Checklist

Before tagging a release:

- [ ] All CI checks pass on the release branch
- [ ] `make lint` passes locally
- [ ] `make test` passes locally
- [ ] `lops check` passes on the repo itself
- [ ] Update version references in docs if needed

## Branch Summary

| Branch | Purpose | Lifetime |
|--------|---------|----------|
| `main` | Stable development branch | Permanent |
| `release-v{major}.{minor}` | Release branch, tags are cut from here | Permanent per minor version |
| `feature/*` | Feature development | Merged to `main` via PR |
