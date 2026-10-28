# Changelog

All notable changes to the public interface of `golden-path-ci` are listed here:
reusable workflow inputs, outputs and required permissions, composite action
inputs and outputs, and gate behaviour. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow
[Semantic Versioning](https://semver.org/). See
[docs/consumer-guide.md](docs/consumer-guide.md) for the versioning contract and
the upgrade path.

## [Unreleased]

## [1.0.0] - 2026-10-28

### Added

- Reusable workflows (`workflow_call`):
  - `go-service.yml`: `go mod verify`, `go vet`, golangci-lint, race tests with a
    coverage gate (`coverage-threshold`, default 70), govulncheck, `go build`.
    Output: `coverage`.
  - `node-service.yml`: `npm ci --ignore-scripts`, `npm audit` (`audit-level`,
    default `high`), lint, type check, test and build scripts.
  - `container-image.yml`: hadolint, a linux/amd64 build scanned by a Trivy gate
    (`severity` default `CRITICAL,HIGH`, `ignore-unfixed` default `true`) with
    a read-only token. Then, only when `push` is true: a multi-arch build and
    push to GHCR, cosign keyless signature, CycloneDX SBOM attestation, SLSA
    build provenance (`actions/attest-build-provenance`), and verification of
    all three. Outputs: `digest`, `image`.
  - `helm-chart.yml`: `helm lint --strict`, helm-unittest, and `helm template`
    piped through `kubeconform -strict` for the chart defaults and every
    `values-files` entry.
  - `terraform.yml`: `terraform fmt -check`, `init -backend=false
    -lockfile=readonly`, `validate`, tflint, Checkov.
  - `security-baseline.yml`: Gitleaks (`gitleaks-mode` `git` or `dir`),
    dependency review on pull requests, and CodeQL for `codeql-languages`
    (default `["actions"]`).
- Composite actions: `setup-toolchain` (Go or Node.js with dependency, build
  and lint caches), `version-stamp` (version, short SHA, reproducible build
  date, image tags), `job-summary` (uniform results table).
- `cache-version` input on every workflow to invalidate caches.
- Self-test on `examples/` (Go service, TypeScript service, Helm chart,
  Terraform module) and a meta-test that proves three negative fixtures fail at
  the expected gate for the expected reason.
- `tools/workflow-policy`: a linter that fails on unpinned actions, missing
  version comments, missing permissions, `write-all`, missing job timeouts,
  `pull_request_target` / `workflow_run`, and checkouts that persist credentials.
  It runs in CI together with actionlint and zizmor.
- Release workflow: verifies the tag is on `main` and that self-test passed for
  the commit, publishes the GitHub Release from this file, and moves the
  floating major tag.

[Unreleased]: https://github.com/ahmedelgmal310/golden-path-ci/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/ahmedelgmal310/golden-path-ci/releases/tag/v1.0.0
