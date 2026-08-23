# Melody Example

This program trains an ordered melody and predicts its next note. It shows the
basic `go-bayes` flow with string values.

## Run

From the repository root:

```sh
go run ./_examples/melody
```

Expected output:

```text
So
```

The queried notes are present in the training sequence. The predictor matches
their exact ordered context and returns the next observed note. It does not
compare melodies by similarity.
