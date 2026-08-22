# Examples

This directory contains runnable examples for `go-bayes`.

## Purpose

Each example is isolated so contributors can understand behavior without
reading all package internals.

## Available Examples

- `iris/`: Uses the Iris dataset in `testdata/iris.json`.

## How To Run

From repository root:

```sh
go test ./bayes -run Example_iris -v
```

Or run all package examples:

```sh
go test ./bayes -run Example -v
```
