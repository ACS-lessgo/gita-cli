# CI Pipeline — Design Spec

## Context

gita-cli has no CI/CD today — this was flagged as a gap in an earlier project
hygiene audit and explicitly deferred as out of core-hygiene scope. The
install scripts (`install.sh`, `install.ps1`) download prebuilt release
binaries from GitHub Releases, but nothing in the repo builds or verifies
those binaries automatically; the whole release process is manual.

This spec covers **CI only**: automated build/vet/test verification on every
push and pull request. Release automation (cross-compiling and publishing
the 5 platform binaries `install.sh`/`install.ps1` expect) was explicitly
scoped out — a separate future spec if wanted.

## Design

A single GitHub Actions workflow, `.github/workflows/ci.yml`, triggered on
`push` and `pull_request` (any branch). One job, matrixed across
`ubuntu-latest`, `macos-latest`, and `windows-latest` — matching the three
platforms the install scripts target, so platform-specific breakage (path
handling, build tags, etc.) is caught before merge rather than discovered by
a user. Each matrix leg:

1. Checks out the repo (`actions/checkout@v4`).
2. Sets up Go 1.22 (`actions/setup-go@v5`), matching the floor in `go.mod`.
3. Runs `go build ./...`.
4. Runs `go vet ./...`.
5. Runs `go test ./...`.

No Makefile, no GoReleaser, no lint config — none exist in the repo today
and adding one solely to wrap three commands would be scope creep beyond
what was asked. `actions/setup-go` caches Go modules by default, so no
explicit caching config is needed.

## File

`.github/workflows/ci.yml`:

```yaml
name: CI
on:
  push:
  pull_request:

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go build ./...
      - run: go vet ./...
      - run: go test ./...
```

## Error handling

A failing step (build, vet, or test) fails that matrix leg's job; GitHub
marks the check red on the PR/commit. No custom error handling needed — this
is standard GitHub Actions behavior.

## Testing

Verification is the workflow running successfully against the current
`main` branch (which already has a clean `go build`/`vet`/`test`) once
pushed. No local test harness exists for GitHub Actions workflows
themselves — correctness is confirmed by observing the Actions run.
