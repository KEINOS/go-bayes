# Contributing

Thanks for contributing to `go-bayes`.

## Prerequisites

- Go 1.22 or newer
- `golangci-lint`
- `markdownlint-cli2`

## Local Commands

- Run unit tests:

```sh
go test ./...
```

- Run tests with race and coverage:

```sh
go test -cover -race ./...
```

- Run lints:

```sh
golangci-lint run --fix
markdownlint-cli2 "**/*.md" --fix
```

- Run fuzz test quickly:

```sh
go test -fuzz=FuzzBayes -fuzztime=5s ./bayes/internal/theorem
```

- Run benchmarks:

```sh
go test -bench=. ./...
```

## Branch And PR Conventions

- Use short branch names prefixed by intent:
  - `fix/...`
  - `feat/...`
  - `docs/...`
  - `refactor/...`
- Keep each PR focused on one topic.
- Include test evidence in the PR description.

## Style Expectations

- Keep comments simple and clear in English.
- Prefer small functions and explicit error handling.
- Follow existing naming and package boundaries.
