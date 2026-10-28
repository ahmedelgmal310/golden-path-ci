# Negative fixture: leaked secret

`deploy-key.pem` has the shape of an OpenSSH private key and matches the
Gitleaks `private-key` rule. Its body is plain text, so it is not a usable key.

The `negative-fixtures` workflow scans this directory with the
`security-baseline` reusable workflow (`gitleaks-mode: dir`). The meta-test passes only
if the `Gitleaks` step fails and its log reports the `private-key` rule for
this file.

The repository-wide scan allowlists this one directory in `.gitleaks.toml`.
Nothing outside it is exempt.
