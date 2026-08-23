# HTTP Status Example

This program learns HTTP status history and predicts recovery after rate
limiting and a temporary outage. It shows the `go-bayes` flow with discrete
integer values from Go's `net/http` package.

## Run

From the repository root:

```sh
go run ./_examples/http_status
```

Expected output:

```text
200
```

The queried `200 -> 429 -> 503` context is present in the training history and
is followed by `200`. The predictor matches that exact ordered context. This
example does not claim to forecast unobserved service failures.
