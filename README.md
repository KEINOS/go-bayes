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

## License

- MIT, Copyright (c) 2020 [KEINOS](https://github.com/KEINOS/) and the [go-bayes contributors](https://github.com/KEINOS/go-bayes/graphs/contributors).
