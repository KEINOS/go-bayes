[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/KEINOS/go-bayes)](https://github.com/KEINOS/go-bayes/blob/main/go.mod)
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-bayes/bayes.svg)](https://pkg.go.dev/github.com/KEINOS/go-bayes/bayes)

# go-bayes

`github.com/KEINOS/go-bayes/bayes` is a Bayesian inference package for Go.

It learns ordered sequences and predicts the value that is likely to come next. The current model is a Folded Context Transition Predictor (FCTP).

> [!IMPORTANT]
> **The API can change before the first stable release.**
> The latest published release is `v0.0.3`. Until `v1.0.0`, a clear API, ease of use, and measured performance are more important than backward compatibility.

## Quick Start

Install the module:

```sh
go get github.com/KEINOS/go-bayes@latest
```

Use the package:

```go
import "github.com/KEINOS/go-bayes/bayes"
```

### String Sequence

Train a sequence and predict its next value. Imports and the outer `main` function are omitted here:

```go
const datasetID uint64 = 100
predictor, err := bayes.New(bayes.MemoryStorage, datasetID)
if err != nil {
 log.Fatal(err)
}

melody := []string{
 "So", "So", "La", "So", "Do", "Si",
 "So", "So", "La", "So", "Re", "Do",
}

err = predictor.Train(melody)
if err != nil {
 log.Fatal(err)
}

classID, err := predictor.Predict([]string{"So", "So", "La", "So", "Do", "Si"})
if err != nil {
 log.Fatal(err)
}

fmt.Println(predictor.GetClass(classID))
// Output: So
```

See the [complete melody program](_examples/melody/README.md).

### Integer Sequence

The same API accepts discrete integer values. This example learns HTTP status history and predicts recovery after rate limiting and a temporary outage. Imports and the outer `main` function are omitted here:

```go
const datasetID uint64 = 101
predictor, err := bayes.New(bayes.MemoryStorage, datasetID)
if err != nil {
 log.Fatal(err)
}

statusHistory := []int{
 http.StatusOK,
 http.StatusCreated,
 http.StatusNoContent,
 http.StatusOK,
 http.StatusTooManyRequests,
 http.StatusServiceUnavailable,
 http.StatusOK,
 http.StatusCreated,
 http.StatusNoContent,
 http.StatusOK,
 http.StatusTooManyRequests,
 http.StatusServiceUnavailable,
 http.StatusOK,
}

err = predictor.Train(statusHistory)
if err != nil {
 log.Fatal(err)
}

classID, err := predictor.Predict([]int{
 http.StatusOK,
 http.StatusTooManyRequests,
 http.StatusServiceUnavailable,
})
if err != nil {
 log.Fatal(err)
}

fmt.Println(predictor.GetClass(classID))
// Output: 200
```

See the [complete HTTP status program](_examples/http_status/README.md).

> [!NOTE]
>
> * `New` creates an isolated predictor backed by in-memory storage and uses xxHash3 for context folding by default.
> * A `Predictor` is not safe for concurrent use; callers must synchronize shared access.

## Features and Behavior

The predictor learns transitions from an ordered context to a possible next value. It uses learned probabilities to select the most likely class for the supplied context. This is Bayesian inference, but it is not a Naive Bayes classifier.

Its current model is a Folded Context Transition Predictor (FCTP). It converts each value to a fixed-width ID and folds an ordered context of any supported length into one context ID. `GetClass` resolves the predicted class ID to the original value recorded during training.

Training `A -> B -> C -> D` also records the suffix contexts that lead to `D`:

```text
C             -> D
FOLD(C)       -> D
FOLD(B, C)    -> D
FOLD(A, B, C) -> D
```

Context order matters. The predictor matches exact folded IDs. It does not measure similarity between values or retry shorter contexts when the complete context is unknown.

### Supported Values

`Train`, `Predict`, and `HashTrans` accept slices or values composed of these built-in types:

* `bool` and `string`;
* `int`, `int16`, `int32`, and `int64`;
* `uint`, `uint16`, `uint32`, and `uint64`;
* `float32` and `float64`.

Integer values preserve their bit pattern. Strings receive deterministic fixed-width IDs. Floating-point values are currently converted with `uint64(value)`, so fractional parts are discarded; use scaled integers or canonical strings when fractions are significant.

IDs are deterministic identifiers, not collision-free or reversible encodings. `GetClass` depends on the predictor's class map and returns `nil` for an unknown class ID.

### Hashers

xxHash3 is the default context-folding algorithm. Select BLAKE3 when compatibility with BLAKE3-derived context IDs is required:

```go
predictor, err := bayes.New(
 bayes.MemoryStorage,
 42,
 bayes.WithHasher("blake3"),
)
```

Use `NewPredictor` to inject a custom implementation of `bayes.Hasher`:

```go
predictor, err := bayes.NewPredictor(bayes.PredictorConfig{
 Storage: bayes.MemoryStorage,
 ScopeID: 42,
 Hasher:  customHasher,
})
```

Changing the hasher changes context IDs. Use the same algorithm for training and prediction. String values use fixed BLAKE3-based item IDs before the selected hasher folds the context.

### Persistence

`Predictor` implements `encoding/json` marshaling and unmarshaling for `MemoryStorage`:

```go
payload, err := json.Marshal(predictor)
if err != nil {
 log.Fatal(err)
}

var restored bayes.Predictor

err = json.Unmarshal(payload, &restored)
if err != nil {
 log.Fatal(err)
}
```

The JSON format stores transition state, scope, storage type, and the class map. JSON does not preserve the exact Go type of numeric class values. For example, an `int` class value is restored as `float64`.

The format does not store the selected hasher. Unmarshaling selects the default xxHash3 hasher. After restoring data trained with BLAKE3 or a custom hasher, call `SetHasher` with the matching implementation before prediction.

`Reset` clears learned transitions and classes. `SetStorage` selects the backend used by the next `Reset`. Only `MemoryStorage` is currently implemented.

## Technical Details

Read the [technical specification](SPEC.md) for the two-ID learning model, token IDs, context folding, suffix expansion, Bayesian scoring, class recovery, and current design limits.

## Examples

The [examples overview](_examples/README.md) links to all runnable programs:

* [Melody](_examples/melody/README.md): String sequence prediction.
* [HTTP status](_examples/http_status/README.md): Integer sequence prediction.
* [Iris](_examples/iris/README.md): 150 records from the original UCI Iris data.
* [Wine](_examples/wine/README.md): 13 ordered numeric measurements.
* [Mushroom](_examples/mushroom/README.md): 22 categorical values.

## Contributing

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/KEINOS/go-bayes)](https://github.com/KEINOS/go-bayes/blob/main/go.mod)
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-bayes/bayes.svg)](https://pkg.go.dev/github.com/KEINOS/go-bayes/bayes)
[![unit-test](https://github.com/KEINOS/go-bayes/actions/workflows/unit-test.yml/badge.svg)](https://github.com/KEINOS/go-bayes/actions/workflows/unit-test.yml)
[![golangci-lint](https://github.com/KEINOS/go-bayes/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/KEINOS/go-bayes/actions/workflows/golangci-lint.yml "Static analysis")
[![codecov](https://codecov.io/gh/KEINOS/go-bayes/branch/main/graph/badge.svg?token=k0VCclM4G7)](https://codecov.io/gh/KEINOS/go-bayes "Code coverage")
[![CodeQL](https://github.com/KEINOS/go-bayes/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/KEINOS/go-bayes/actions/workflows/codeql-analysis.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/KEINOS/go-bayes)](https://goreportcard.com/report/github.com/KEINOS/go-bayes "View report card")

Contributor resources:

* [Contributing guide](.github/CONTRIBUTING.md): Implementation details and validation commands.
* [Security policy](.github/SECURITY.md): How to report a vulnerability.
* [Issues and roadmap](https://github.com/KEINOS/go-bayes/issues): Bugs, feature requests, and planned work.

## License

`go-bayes` is available under the [MIT License](LICENSE).
