# Consumer guide: adopting and upgrading golden-path-ci

## Pick a reference style

| Style | Example | You get | Trade-off |
|---|---|---|---|
| Floating major tag | `@v1` | Every `v1.x.y` fix and feature automatically | Behaviour can change under you within v1 (only backwards-compatible changes, see the contract below) |
| Exact release tag | `@v1.2.0` | Nothing changes until you edit the line | Tags can be moved by the maintainer, so this pins a name and not content |
| Full commit SHA | `@<40-hex> # v1.2.0` | Byte-for-byte reproducible CI, and Dependabot opens a PR per release | One small PR per release to review and merge |

Regulated or high-risk repositories should use the SHA form. Dependabot's
`github-actions` ecosystem understands reusable-workflow references and keeps the
`# v1.2.0` comment in sync:

```yaml
jobs:
  go:
    uses: ahmedelgmal310/golden-path-ci/.github/workflows/go-service.yml@<sha> # v1.0.0
```

The reusable workflows check out their own composite actions at
`job.workflow_sha`, the exact commit you referenced. A SHA-pinned consumer
therefore never runs unpinned library code.

## Versioning contract

`golden-path-ci` follows [Semantic Versioning](https://semver.org/) on its public
interface: the reusable workflows' `inputs`, `outputs` and required caller
`permissions`, and the composite actions' `inputs` and `outputs`.

| Change | Release |
|---|---|
| Bug fix, tool or action bump that keeps behaviour, new optional input with a default that keeps behaviour | patch `v1.x.Y` or minor `v1.X.0` |
| New gate that is **off by default**, new output | minor `v1.X.0` |
| Removed or renamed input, new required input, new required caller permission, a gate that is **on by default** and fails builds that used to pass | major `v2.0.0` |

A gate that gets stricter (for example a lower Trivy severity default) is a major
change, even if the YAML interface is unchanged. Failing builds that used to
pass is a breaking change for the consumer.

## Upgrading

### Within a major version (`v1.x` → `v1.y`)

1. Read the [CHANGELOG](../CHANGELOG.md) entries between your version and the new one.
2. `@v1` consumers: nothing to do. SHA consumers: merge the Dependabot PR once
   your CI is green on it.
3. To try a release before it reaches you, point one repository at the release
   SHA on a branch and compare the job summaries.

### Across a major version (`v1` → `v2`)

1. The CHANGELOG `v2.0.0` section has a **Migration** list. Each item names the
   input or permission that changed and the replacement.
2. `v1` keeps receiving security fixes for 90 days after `v2.0.0` ships.
3. Change `@v1` to `@v2` (or update the SHA) and apply the migration items in one PR.

### Rolling back

Point the `uses:` line back at the previous SHA or `v1.x.y` tag. Nothing in the
library stores state, so a rollback takes effect on the next run.

## Required caller permissions

A reusable workflow cannot get more permissions than its caller grants.

| Workflow | Caller must grant |
|---|---|
| `go-service.yml`, `node-service.yml`, `helm-chart.yml`, `terraform.yml` | `contents: read` (the default) |
| `container-image.yml` | `contents: read`, `packages: write`, `id-token: write`, `attestations: write` |
| `security-baseline.yml` | `contents: read`, `security-events: write`, `actions: read` |

## Cache invalidation

Every workflow takes `cache-version` (default `v1`). Bump it, for example to `v2`, to
start from empty caches after a poisoned or corrupted cache. Old entries age out
under GitHub's normal 7-day eviction.
