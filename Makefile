.PHONY: all
all: check

.PHONY: check
check: lint test

.PHONY: clean
clean:
	@rm -f coverage.out go-carpet-coverage-out*

.PHONY: test
test:
	@go test -race ./...

.PHONY: lint
lint:
	@golangci-lint run --fix
	@markdownlint-cli2 "**/*.md" --fix

.PHONY: fuzz
fuzz:
	go test -fuzz=FuzzBayes -fuzztime=5s ./bayes/internal/theorem

.PHONY: coverage
coverage:
	@go-carpet -mincov 99.9

.PHONY: bench
bench:
	@go test -bench=. ./...
