// Package main demonstrates transition prediction with the UCI Iris dataset.
package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/KEINOS/go-bayes/bayes"
)

var (
	errClassNotSpecies = errors.New("predicted class is not a species string")
	errEmptyDataset    = errors.New("iris dataset is empty")
	errEmptySpecies    = errors.New("iris species is empty")
)

const (
	datasetScopeID      uint64 = 53
	expectedSampleCount int    = 150
)

//go:embed iris.data
var rawIrisData []byte

type irisSample struct {
	measurements [4]string
	species      string
}

func main() {
	err := run(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func readIris(input io.Reader) ([]irisSample, error) {
	csvReader := csv.NewReader(input)
	csvReader.FieldsPerRecord = 5

	samples := make([]irisSample, 0, expectedSampleCount)

	for rowIndex := 1; ; rowIndex++ {
		fields, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("read Iris row %d: %w", rowIndex, err)
		}

		sample := new(irisSample)

		for columnIndex := range sample.measurements {
			_, err := strconv.ParseFloat(fields[columnIndex], 64)
			if err != nil {
				return nil, fmt.Errorf(
					"parse Iris row %d column %d: %w",
					rowIndex,
					columnIndex+1,
					err,
				)
			}

			sample.measurements[columnIndex] = fields[columnIndex]
		}

		sample.species = strings.TrimSpace(fields[4])
		if sample.species == "" {
			return nil, fmt.Errorf("iris row %d: %w", rowIndex, errEmptySpecies)
		}

		samples = append(samples, *sample)
	}

	if len(samples) == 0 {
		return nil, errEmptyDataset
	}

	return samples, nil
}

func run(output io.Writer) error {
	return runWithDependencies(output, rawIrisData, bayes.MemoryStorage)
}

func runWithDependencies(output io.Writer, data []byte, storage bayes.Storage) error {
	samples, err := readIris(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("load embedded Iris data: %w", err)
	}

	predictor, err := trainPredictor(samples, storage)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "trained: %d samples\n", len(samples))
	if err != nil {
		return fmt.Errorf("write training summary: %w", err)
	}

	queries := [][4]string{
		{"5.1", "3.5", "1.4", "0.2"},
		{"7.0", "3.2", "4.7", "1.4"},
		{"6.3", "3.3", "6.0", "2.5"},
	}

	for _, measurements := range queries {
		// Predict returns a fixed-width class ID. GetClass resolves that ID to
		// the original species value supplied during training.
		classID, err := predictor.Predict(context.Background(), measurements[:])
		if err != nil {
			return fmt.Errorf("predict Iris species: %w", err)
		}

		species, ok := predictor.GetClass(classID).(string)
		if !ok {
			return fmt.Errorf("resolve class %d: %w", classID, errClassNotSpecies)
		}

		_, err = fmt.Fprintf(
			output,
			"%s, %s, %s, %s -> %s\n",
			measurements[0],
			measurements[1],
			measurements[2],
			measurements[3],
			species,
		)
		if err != nil {
			return fmt.Errorf("write prediction: %w", err)
		}
	}

	return nil
}

func trainPredictor(samples []irisSample, storage bayes.Storage) (*bayes.Predictor, error) {
	predictor, err := bayes.New(context.Background(), storage, datasetScopeID)
	if err != nil {
		return nil, fmt.Errorf("create predictor: %w", err)
	}

	for _, sample := range samples {
		// Train treats the final value as the expected next class for the four
		// ordered measurements that precede it.
		sequence := []any{
			sample.measurements[0],
			sample.measurements[1],
			sample.measurements[2],
			sample.measurements[3],
			sample.species,
		}

		err := predictor.Train(context.Background(), sequence)
		if err != nil {
			return nil, fmt.Errorf("train predictor: %w", err)
		}
	}

	return predictor, nil
}
