# Negative fixtures

Each directory here must **fail** a golden-path gate for one specific reason.
`.github/workflows/negative-fixtures.yml` runs the real reusable workflows
against them, and `scripts/meta-test.sh` (the `Meta-test` job in self-test)
asserts each result:

| Fixture | Reusable workflow | Must fail at step | Log must contain |
|---|---|---|---|
| [`leaked-secret/`](leaked-secret) | `security-baseline.yml` (`gitleaks-mode: dir`) | `Gitleaks` | `private-key` and `leaks found: 1` |
| [`critical-cve-image/`](critical-cve-image) | `container-image.yml` (`push: false`) | `Trivy vulnerability gate` | `CVE-2021-36159` |
| [`unformatted-terraform/`](unformatted-terraform) | `terraform.yml` | `terraform fmt` | `main.tf` |

The meta-test also checks that every step before the expected one succeeded,
so a broken checkout or a typo cannot pass as the right failure.
