EXAMPLE_DIRS := $(sort $(dir $(wildcard _examples/*/*.go)))

.PHONY: all
all: check

.PHONY: check
check: lint test test_example

.PHONY: clean
clean:
	@rm -f coverage.out go-carpet-coverage-out*

.PHONY: test
test:
	@go test -race ./...

.PHONY: test_example
test_example:
	@set -e; for dir in $(EXAMPLE_DIRS); do \
		(cd "$$dir" && go test -race ./...); \
	done

.PHONY: lint
lint:
	@golangci-lint run --fix ./... $(EXAMPLE_DIRS)
	@markdownlint-cli2 "**/*.md" --fix

.PHONY: fuzz
fuzz:
	@go test -fuzz=FuzzBayes -fuzztime=5s ./bayes/internal/theorem

.PHONY: coverage
coverage:
	@go-carpet -mincov 99.9

.PHONY: bench
bench:
	@go test -bench=. ./...
