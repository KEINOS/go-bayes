<!-- markdownlint-disable MD041 -->
[![Go Version](https://img.shields.io/badge/Go-1.18+-blue?logo=go)](https://github.com/KEINOS/go-bayes/blob/main/go.mod)
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-bayes.svg)](https://pkg.go.dev/github.com/KEINOS/go-bayes)

# go-bayes

`github.com/KEINOS/go-bayes` is a Go package for Bayesian inference.

```go
go get github.com/KEINOS/go-bayes
```

```go
import "github.com/KEINOS/go-bayes"

func Example() {
    // "Happy Birthday", the train data. The types of slices available for the
    // training are as follows:
    //   bool, int, int16-int64, uint, uint16-uint64, float32, float64, string.
    score := []string{
        "So", "So", "La", "So", "Do", "Si",
        "So", "So", "La", "So", "Re", "Do",
        "So", "So", "So", "Mi", "Do", "Si", "La",
        "Fa", "Fa", "Mi", "Do", "Re", "Do",
    }

    // Reset the trained model
    bayes.Reset()

    // Train
    if err := bayes.Train(score); err != nil {
        log.Fatal(err)
    }

    // Predict the next note from the introduction notes
    for _, intro := range [][]string{
        {"So", "So", "La", "So", "Do", "Si"},             // --> So
        {"So", "So", "La", "So", "Do", "Si", "So", "So"}, // --> La
        {"So", "So", "La"},                               // --> So
        {"So", "So", "So"},                               // --> Mi
    } {
        nextNoteID, err := bayes.Predict(intro)
        if err != nil {
            log.Fatal(err)
        }

        // Print the predicted next note
        nextNoteString := bayes.GetClass(nextNoteID)

        fmt.Printf("Next is: %v (Class ID: %v)\n", nextNoteString, nextNoteID)
    }

    // Output:
    // Next is: So (Class ID: 10062876669317908741)
    // Next is: La (Class ID: 17627200281938459623)
    // Next is: So (Class ID: 10062876669317908741)
    // Next is: Mi (Class ID: 6586414841969023711)
}
```

## Examples

- [Training with a slice of boolean values](https://pkg.go.dev/github.com/KEINOS/go-bayes#example-Train-Bool)
- [Training with a slice of int values](https://pkg.go.dev/github.com/KEINOS/go-bayes#example-Train-Int)

## Contribute

[![unit-test](https://github.com/KEINOS/go-bayes/actions/workflows/unit-test.yml/badge.svg)](https://github.com/KEINOS/go-bayes/actions/workflows/unit-test.yml)
[![golangci-lint](https://github.com/KEINOS/go-bayes/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/KEINOS/go-bayes/actions/workflows/golangci-lint.yml "Static Analysis")
[![codecov](https://codecov.io/gh/KEINOS/go-bayes/branch/main/graph/badge.svg?token=k0VCclM4G7)](https://codecov.io/gh/KEINOS/go-bayes "Code Coverage")
[![CodeQL](https://github.com/KEINOS/go-bayes/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/KEINOS/go-bayes/actions/workflows/codeql-analysis.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/KEINOS/go-bayes)](https://goreportcard.com/report/github.com/KEINOS/go-bayes "View Report Card")

- Any PullRequest for improvement are welcome!
- Branch to PR: `main`
  - [Draft PR](https://github.blog/2019-02-14-introducing-draft-pull-requests/) before full implementation is recommended.
- We will merge any PR for the better, as long as it passes the [CI](https://github.com/KEINOS/go-bayes/actions)s and not a prank-kind commit. ;-)

## License

- MIT, Copyright (c) 2020 [KEINOS](https://github.com/KEINOS/) and the [go-bayes contributors](https://github.com/KEINOS/go-bayes/graphs/contributors).
