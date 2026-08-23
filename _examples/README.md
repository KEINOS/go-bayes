# Examples

This directory contains runnable examples for `go-bayes`.

## Purpose

Each example is isolated so contributors can understand behavior without reading all package internals.

## Available Examples

- [`http_status/`](http_status/): Learns integer HTTP status history and predicts the status after an observed recovery context.
- [`iris/`](iris/): Trains and queries a predictor with the original comma-separated Iris data distributed by the UCI Machine Learning Repository.
- [`melody/`](melody/): Learns an ordered string melody and predicts its next note.
- [`mushroom/`](mushroom/): Uses 22 categorical features, including a missing-value token, to recover an edible or poisonous class.
- [`wine/`](wine/): Uses all 13 chemical measurements from the UCI Wine dataset to recover one of three cultivar classes.

## How To Run

From repository root:

```sh
go run ./_examples/http_status
go run ./_examples/iris
go run ./_examples/melody
go run ./_examples/mushroom
go run ./_examples/wine
```

Run the executable example's tests:

```sh
make test_example
```
