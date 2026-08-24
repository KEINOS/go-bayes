[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/KEINOS/go-bayes)](https://github.com/KEINOS/go-bayes/blob/main/go.mod)
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-bayes/bayes.svg)](https://pkg.go.dev/github.com/KEINOS/go-bayes/bayes)

# go-bayes

`github.com/KEINOS/go-bayes/bayes` is a Bayesian inference package for Go.

It learns ordered sequences and predicts the value that is likely to come next. The current model is a Folded Context Transition Predictor (FCTP).

> [!IMPORTANT]
> **The API can change before the first stable release.**
> The latest published release is `v0.0.4`. Until `v1.0.0`, a clear API, ease of use, and measured performance are more important than backward compatibility.

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

Train a sequence and predict its next value.

```go
ctx := context.Background()
const datasetID uint64 = 100
predictor, err := bayes.New(ctx, bayes.MemoryStorage, datasetID)
if err != nil {
 log.Fatal(err)
}
defer predictor.Close()

melody := []string{
 "So", "So", "La", "So", "Do", "Si",
 "So", "So", "La", "So", "Re", "Do",
}

err = predictor.Train(ctx, melody)
if err != nil {
 log.Fatal(err)
}

classID, err := predictor.Predict(ctx, []string{"So", "So", "La", "So", "Do", "Si"})
if err != nil {
 log.Fatal(err)
}

fmt.Println(predictor.GetClass(classID))
// Output: So
```

View the [complete source](_examples/melody/main.go), or [run it online](https://go.dev/play/p/FYL7mIvcSNa).

### Integer Sequence

The same API accepts discrete integer values. This example learns HTTP status history and predicts recovery after rate limiting and a temporary outage.

```go
ctx := context.Background()
const datasetID uint64 = 101
predictor, err := bayes.New(ctx, bayes.MemoryStorage, datasetID)
if err != nil {
 log.Fatal(err)
}
defer predictor.Close()

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

err = predictor.Train(ctx, statusHistory)
if err != nil {
 log.Fatal(err)
}

classID, err := predictor.Predict(ctx, []int{
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

View the [complete source](_examples/http_status/main.go), or [run it online](https://go.dev/play/p/QPbxmh_g_4a).

> [!NOTE]
>
> * `New` creates an isolated predictor backed by in-memory storage and uses xxHash3 for value and context IDs by default.
> * A `Predictor` is not safe for concurrent use; callers must synchronize shared access.

### More Examples

A context can be a time sequence or an ordered set of features. The [examples overview](_examples/README.md) links to runnable programs for both uses:

* [Melody](_examples/melody/README.md): Predict the next note in a string sequence.
* [HTTP status](_examples/http_status/README.md): Predict the next status in an integer sequence.
* [Iris](_examples/iris/README.md): Predict a species from four ordered measurements.
* [Wine](_examples/wine/README.md): Predict a cultivar class from 13 ordered measurements.
* [Mushroom](_examples/mushroom/README.md): Predict the edible or poisonous class from 22 ordered categorical features.

The [Go Reference examples](https://pkg.go.dev/github.com/KEINOS/go-bayes/bayes#pkg-examples) provide shorter examples for [melody](https://pkg.go.dev/github.com/KEINOS/go-bayes/bayes#example-package), [Iris](https://pkg.go.dev/github.com/KEINOS/go-bayes/bayes#example-package-Iris), [`New`](https://pkg.go.dev/github.com/KEINOS/go-bayes/bayes#example-New), [`WithHasher`](https://pkg.go.dev/github.com/KEINOS/go-bayes/bayes#example-WithHasher), and [`Storage.Type`](https://pkg.go.dev/github.com/KEINOS/go-bayes/bayes#example-Storage.Type).

## Features and Behavior

The predictor learns transitions from an ordered context to a possible next value. It uses learned probabilities to select the most likely class for the supplied context. This is Bayesian inference, but it is not a Naive Bayes classifier.

Its current model is a Folded Context Transition Predictor (FCTP). It converts each value to a fixed-width ID and folds an ordered context of any supported length into one context ID. `GetClass` resolves the predicted class ID to the original value recorded during training.

Training `A -> B -> C -> D` records the suffix contexts that lead to `D`:

```text
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

Each value is encoded with its Go type before it is hashed. For example, `true`, `int(1)`, `uint64(1)`, and `float64(1)` have different IDs. Integer signs and floating-point fractions are preserved.

IDs are deterministic identifiers, not collision-free or reversible encodings. Training fails with `ErrHashCollision` instead of replacing a class when two class values produce the same ID. `GetClass` depends on the predictor's class map and returns `nil` for an unknown class ID.

### Hashers

xxHash3 is the default algorithm for value and context IDs. Select BLAKE3 when you need BLAKE3-based IDs:

```go
predictor, err := bayes.New(
 context.Background(),
 bayes.MemoryStorage,
 42,
 bayes.WithHasher("blake3"),
)
```

Use `NewPredictor` to inject a custom implementation of `bayes.Hasher`:

```go
predictor, err := bayes.NewPredictor(context.Background(), bayes.PredictorConfig{
 Storage: bayes.MemoryStorage,
 ScopeID: 42,
 Hasher:  customHasher,
})
```

The selected hasher creates every value ID and context ID. It is fixed when the predictor is created. A custom hasher must return a stable, non-empty name for model-file compatibility.

### Model Storage

Memory storage is the simplest choice for a short-lived predictor. Save its complete state as a portable SQLite model file:

```go
err := predictor.Save(ctx, "model.db")
```

`Load` copies a saved model into memory. `Open` operates directly on the model file and keeps new training data there:

```go
inMemory, err := bayes.Load(ctx, "model.db")
onDisk, err := bayes.Open(ctx, "model.db")
```

Call `Close` for every predictor. A directly opened model has exclusive lifetime ownership, so another cooperating process cannot open or replace the same path until it is closed.

Create a new file-backed model with `SQLiteStorage`:

```go
predictor, err := bayes.New(
 ctx,
 bayes.SQLiteStorage,
 datasetID,
 bayes.WithSQLitePath("model.db"),
)
```

SQLite support uses `github.com/mattn/go-sqlite3` and requires cgo. Builds with `CGO_ENABLED=0` can use memory storage, but `Save`, `Load`, `Open`, and `SQLiteStorage` return `ErrSQLiteUnavailable`.

Model files preserve exact supported Go value types, transition counts, scope, codec version, and hasher identity. A custom-hasher model can be loaded only when the same compatible `Hasher` is supplied. JSON model persistence is no longer supported.

`Train` and `Reset` are atomic store operations. `Reset` clears learned transitions and classes but keeps the current storage backend, scope, and hasher.

## Technical Details

Read the [technical specification](SPEC.md) for the two-ID learning model, token IDs, context folding, suffix expansion, Bayesian scoring, class recovery, and current design limits.

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
