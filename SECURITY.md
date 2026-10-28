# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| `v1` (latest `v1.x.y`) | Yes |
| anything older | No |

Fixes ship as a new `v1.x.y` patch release, and the floating `v1` tag moves to it.
Consumers that pin a full commit SHA get the fix through their Dependabot PR.

## Reporting a vulnerability

Please report vulnerabilities privately through
[GitHub Security Advisories](https://github.com/ahmedelgmal310/golden-path-ci/security/advisories/new).
Do not open a public issue. You can expect an acknowledgement within 72 hours.

## Supply-chain controls in this repository

- Every action in every workflow and composite action is pinned to a full commit SHA
  with a version comment. Every tool image is pinned by digest, and every
  downloaded binary is SHA-256 verified. `tools/workflow-policy` enforces this in
  CI, together with actionlint and zizmor.
- Workflows declare least-privilege `permissions`. The image build runs with a
  read-only token, and only the publish job can write packages or request an OIDC token.
- `actions/checkout` never persists credentials. There is no `pull_request_target`
  and no `workflow_run` trigger.
- Images built by `container-image.yml` are signed keyless with cosign and carry a
  CycloneDX SBOM attestation and SLSA build provenance. The signing identity is
  the reusable workflow itself. Verify an image with:

  ```sh
  cosign verify ghcr.io/ahmedelgmal310/golden-path-ci/example-go-service:<tag> \
    --certificate-identity 'https://github.com/ahmedelgmal310/golden-path-ci/.github/workflows/container-image.yml@refs/heads/main' \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com
  gh attestation verify oci://ghcr.io/ahmedelgmal310/golden-path-ci/example-go-service:<tag> \
    --repo ahmedelgmal310/golden-path-ci
  ```

## Test fixtures

`examples/negative/leaked-secret/deploy-key.pem` looks like a private key so that
Gitleaks flags it. It is plain text and is not key material. It is allowlisted for
the repository-wide scan only, in `.gitleaks.toml`.
