// Package main demonstrates string sequence prediction with a melody.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/KEINOS/go-bayes/bayes"
)

const datasetScopeID uint64 = 100

func main() {
	err := run(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func run(output io.Writer) error {
	predictor, err := bayes.New(context.Background(), bayes.MemoryStorage, datasetScopeID)
	if err != nil {
		return fmt.Errorf("create predictor: %w", err)
	}

	melody := []string{
		"So", "So", "La", "So", "Do", "Si",
		"So", "So", "La", "So", "Re", "Do",
	}

	err = predictor.Train(context.Background(), melody)
	if err != nil {
		return fmt.Errorf("train melody: %w", err)
	}

	classID, err := predictor.Predict(context.Background(), []string{"So", "So", "La", "So", "Do", "Si"})
	if err != nil {
		return fmt.Errorf("predict next note: %w", err)
	}

	_, err = fmt.Fprintln(output, predictor.GetClass(classID))
	if err != nil {
		return fmt.Errorf("write prediction: %w", err)
	}

	return nil
}
