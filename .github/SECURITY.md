# Security Policy

## Fail-Fast Policy

This project follows a fail-fast policy. The supported Go version is the version declared in [`go.mod`](../go.mod). Dependencies listed in `go.mod` are kept at their latest stable versions.

If a Go or dependency update exposes a problem, CI should fail as soon as possible. We fix the cause instead of hiding or delaying the failure.

## Reporting a Vulnerability

Please [open an issue](https://github.com/KEINOS/go-bayes/issues) to report a vulnerability.

## Minimum Measures

The repository uses these checks:

- CodeQL analysis runs every week.
- Dependabot checks Go modules for updates every week.
- CI tests each proposed change before merge.
