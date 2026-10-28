# Pinned versions

Every action, tool and image is pinned. Actions use a full commit SHA with the
version in a comment. Images use `tag@sha256:digest`. Downloaded binaries are
SHA-256 verified. Go modules are locked by `go.sum` (none needed: the examples
use the standard library only, and the linter uses `gopkg.in/yaml.v3`), npm by
`package-lock.json` and the Terraform provider by `.terraform.lock.hcl`.

A version is adopted only after it has been public for **at least 14 days**.
Dependabot enforces the same cooldown (`.github/dependabot.yml`). The versions were
checked on **2026-10-28**, so the cut-off date is 2026-10-14.

Where each pin lives: workflow defaults in `.github/workflows/*.yml` (inputs),
composite actions in `actions/*/action.yml`, policy tooling images in
`.github/workflows/self-test.yml`, example images in each `Dockerfile`.

## Exception: Go 1.27.3

Go 1.27.3 was released on 2026-10-19, 9 days before the check, with security
fixes to `crypto/x509`. The previous release, 1.27.2 (2026-10-08), passes the
cooldown. However, the container-image Trivy gate scans the Go standard library
compiled into the image binary and would fail on the 1.27.2 `crypto/x509`
advisory. Shipping a toolchain with a known, fixed vulnerability is worse than
skipping the cooldown for a security patch from the Go team. The `go-version`
default and the example builder image therefore use 1.27.3. This is the only exception.

## GitHub Actions

| Action | Version | Released | SHA |
|---|---|---|---|
| actions/checkout | v7.0.1 | 2026-07-20 | `3d3c42e5aac5ba805825da76410c181273ba90b1` |
| actions/setup-go | v7.0.0 | 2026-07-16 | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` |
| actions/setup-node | v7.1.0 | 2026-10-08 | `949feb2413d6458794dcd2491c4babbbce0c15c1` |
| actions/cache | v6.1.0 | 2026-06-26 | `55cc8345863c7cc4c66a329aec7e433d2d1c52a9` |
| actions/upload-artifact | v7.0.2 | 2026-10-07 | `cf430e030ddbb5b0abf93d22962f4752f3646cd9` |
| actions/download-artifact | v8.0.2 | 2026-10-07 | `9000827ccba6bdab643e8b6fd33ac0654aef8333` |
| actions/attest-build-provenance | v4.2.2 | 2026-08-06 | `4d101475d8b20a2381f78447822ac1eab6504dd8` |
| actions/dependency-review-action | v5.0.0 | 2026-05-08 | `a1d282b36b6f3519aa1f3fc636f609c47dddb294` |
| github/codeql-action (init, analyze) | v4.38.3 | 2026-10-08 | `24c54180a607b1449ed407dd24f251e4e9147c8d` |
| docker/setup-qemu-action | v4.4.0 | 2026-09-15 | `99012661954931238ded8c8b007157a8430204e1` |
| docker/setup-buildx-action | v4.4.1 | 2026-09-16 | `f87e5991a6d7451dcb8d9637bfbc97413f497069` |
| docker/build-push-action | v7.4.0 | 2026-09-15 | `c3c9e263c25d99ce0380d002d59b67737d91b0dc` |
| docker/login-action | v4.6.0 | 2026-07-29 | `dbcb813823bdd20940b903addbd779551569679f` |
| sigstore/cosign-installer | v4.1.2 | 2026-05-07 | `6f9f17788090df1f26f669e9d70d6ae9567deba6` |
| azure/setup-helm | v5.0.1 | 2026-06-23 | `9bc31f4ebc9c6b171d7bfbaa5d006ae7abdb4310` |
| hashicorp/setup-terraform | v4.0.1 | 2026-05-12 | `dfe3c3f87815947d99a8997f908cb6525fc44e9e` |
| golangci/golangci-lint-action | v9.3.0 | 2026-06-29 | `ba0d7d2ec06a0ea1cb5fa41b2e4a3ab91d21278a` |

Newer releases that were skipped because of the cooldown: codeql-action v4.39.x
(2026-10-15 and later) and docker/login-action v4.7.0 (2026-10-23).

## Toolchains and CLIs (workflow input defaults)

| Tool | Version | Released | How it is pinned |
|---|---|---|---|
| Go | 1.27.3 | 2026-10-19 | `go-version` (security exception above) |
| Node.js | 24.22.0 (LTS "Krypton") | 2026-10-14 | `node-version` |
| golangci-lint | v2.14.0 | 2026-09-24 | `golangci-lint-version` |
| govulncheck (golang.org/x/vuln) | v1.9.0 | 2026-10-11 | `govulncheck-version`, Go module proxy checksum |
| Helm | v4.3.0 | 2026-09-09 | `helm-version` (v4.3.1 is from 2026-10-15) |
| helm-unittest | v1.2.1 | 2026-10-03 | `helm-unittest-version` |
| kubeconform | v0.8.0 | 2026-06-04 | `kubeconform-version` + SHA-256 |
| Kubernetes schemas | 1.34.0 | n/a | `kubernetes-version` (newest line published in yannh/kubernetes-json-schema) |
| Terraform | 1.16.5 | 2026-10-02 | `terraform-version` |
| tflint | v0.64.0 | 2026-07-17 | `tflint-version` + SHA-256 |
| cosign | v3.1.3 | 2026-08-06 | `cosign-installer` `cosign-release` |

## Images

| Image | Version | Released | Pin |
|---|---|---|---|
| Trivy | 0.75.0 | 2026-10-01 | `mirror.gcr.io/aquasec/trivy@sha256:af6acf9a6b85dfe389a1941505c0ce9efef52a4719635e1a962f022a3d855daa` |
| Gitleaks | v8.30.1 | 2026-03-21 | `ghcr.io/gitleaks/gitleaks@sha256:c00b6bd0aeb3071cbcb79009cb16a60dd9e0a7c60e2be9ab65d25e6bc8abbb7f` |
| Checkov | 3.3.29 | 2026-10-13 | `bridgecrew/checkov@sha256:3cc125b5337717b7303dbddd4e79374c120687b5ff6a24f9abdad56b562469f3` |
| hadolint | v2.15.1 | 2026-07-31 | `mirror.gcr.io/hadolint/hadolint@sha256:32dac94127fd60b7b7e3fbfc65e1383b9b5e25c9bfd7b8536de7a539fe68a12d` |
| actionlint | 1.7.12 | 2026-03-30 | `mirror.gcr.io/rhysd/actionlint@sha256:b1934ee5f1c509618f2508e6eb47ee0d3520686341fec936f3b79331f9315667` |
| zizmor | 1.30.1 | 2026-09-09 | `ghcr.io/zizmorcore/zizmor@sha256:a2eb396d886c053073405c7a980f2139ba2248ec172243cfa3841e57196e8101` |
| golang (example builder) | 1.27.3-alpine3.24 | 2026-10-19 | `golang@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb` |
| distroless static (example runtime) | debian13 nonroot | rolling | `gcr.io/distroless/static-debian13@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99` |

Scanners run as official images pinned by digest, not through wrapper actions
such as `aquasecurity/trivy-action` or `gitleaks/gitleaks-action`. A moved tag or a
compromised wrapper therefore cannot change what runs.

## Example dependencies

| Dependency | Version | Released |
|---|---|---|
| typescript | 6.0.3 | 2026-04-16 |
| @types/node | 24.19.2 | 2026-10-09 |
| hashicorp/aws provider | 6.68.0 (`~> 6.68`, locked) | 2026-10-07 |
| gopkg.in/yaml.v3 (workflow-policy) | v3.0.1 | 2022-05-27 |

## Deliberately outdated (negative fixtures)

`examples/negative/critical-cve-image` uses `alpine:3.14.0@sha256:adab3844f497ab9171f070d4cae4114b5aec565ac772e2f2579405b78be67c96`
on purpose. It carries CVE-2021-36159 (CRITICAL) and must fail the Trivy gate.
Dependabot does not watch that directory.
