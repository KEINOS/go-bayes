// Package main demonstrates integer sequence prediction with HTTP status codes.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/KEINOS/go-bayes/bayes"
)

const datasetScopeID uint64 = 101

func main() {
	err := run(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func run(output io.Writer) error {
	return runWithStorage(output, bayes.MemoryStorage)
}

func runWithStorage(output io.Writer, storage bayes.Storage) error {
	predictor, err := bayes.New(context.Background(), storage, datasetScopeID)
	if err != nil {
		return fmt.Errorf("create predictor: %w", err)
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

	err = predictor.Train(context.Background(), statusHistory)
	if err != nil {
		return fmt.Errorf("train HTTP status history: %w", err)
	}

	classID, err := predictor.Predict(context.Background(), []int{
		http.StatusOK,
		http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
	})

	if err != nil {
		return fmt.Errorf("predict next HTTP status: %w", err)
	}

	_, err = fmt.Fprintln(output, predictor.GetClass(classID))
	if err != nil {
		return fmt.Errorf("write prediction: %w", err)
	}

	return nil
}
