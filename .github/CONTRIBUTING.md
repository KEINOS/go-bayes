# Contributing

Thanks for contributing to `go-bayes`.

## Prerequisites

- The Go version declared in [`go.mod`](../go.mod)
- `golangci-lint`
- `markdownlint-cli2`

## Local Commands

Run the complete local validation gate:

```sh
make check
```

Run individual checks:

```sh
make test
make test_example
make coverage
make bench
make fuzz
```

The library packages have 100% statement coverage. `make coverage` uses a 99.9% minimum. Directories beginning with `_` are excluded from Go's default `./...` expansion, so `make test_example` finds and tests each runnable example explicitly.

## Package Structure

- `bayes`: Public constructor, predictor, persistence, and hasher selection API.
- `bayes/modelstore`: Public transition and class storage contract.
- `bayes/internal/hashers/xxHash3base`: Default xxHash3 value and context hasher.
- `bayes/internal/hashers/blake3base`: Optional BLAKE3 value and context hasher.
- `bayes/internal/modelstores/mapstore`: Built-in in-memory model store.
- `bayes/internal/modelstores/sqlitestore`: Built-in cgo SQLite model store.
- `bayes/internal/theorem`: Current Bayes-based scoring helper.

```mermaid
flowchart TD
    A[bayes Predictor] --> B[value and context Hasher]
    A --> C[ModelStore]
    C --> D[theorem scoring helper]
    B --> E[xxHash3 default]
    B --> F[BLAKE3 optional]
    C --> G[in-memory map store]
    C --> H[SQLite file store]
```

Read the [technical specification](../SPEC.md) before changing context folding, training expansion, scoring, or class recovery.

## Reporting Bugs

Open an issue when you find a bug. When possible, include a small test or program that reproduces it. Reproduction code helps us confirm the problem and respond faster.

## Branch and Pull Request Conventions

Use short branch names prefixed by intent: `fix/...`, `feat/...`, `docs/...`, or `refactor/...`.

- Follow the "fix one thing" rule. Keep each PR focused on one topic.
- Open an issue before starting a large PR so that we can discuss its scope.
- Include test evidence in the PR description.

## Style Expectations

- Keep comments simple and clear in English.
- Prefer small functions and explicit error handling.
- Follow existing naming and package boundaries.
