# Examples

This directory contains runnable examples for `go-bayes`.

## Purpose

Each example is isolated so contributors can understand behavior without reading all package internals.

## Available Examples

- [`iris/`](iris/): Trains and queries a predictor with the original comma-separated Iris data distributed by the UCI Machine Learning Repository.
- [`wine/`](wine/): Uses all 13 chemical measurements from the UCI Wine dataset to recover one of three cultivar classes.

## How To Run

From repository root:

```sh
go run ./_examples/iris
go run ./_examples/wine
```

Run the executable example's tests:

```sh
make test_example
```
