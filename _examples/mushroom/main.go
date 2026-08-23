// Package main demonstrates transition prediction with the UCI Mushroom dataset.
package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/KEINOS/go-bayes/bayes"
)

var (
	errClassNotString = errors.New("predicted class is not a string")
	errEmptyDataset   = errors.New("mushroom dataset is empty")
	errInvalidFeature = errors.New("mushroom feature must be one character")
	errUnknownClass   = errors.New("mushroom class must be e or p")
)

const (
	classColumnIndex    int    = 0
	datasetScopeID      uint64 = 73
	expectedSampleCount int    = 8124
	featureColumnOffset int    = 1
	featureCount        int    = 22
)

//go:embed agaricus-lepiota.data
var rawMushroomData []byte

type mushroomSample struct {
	features [featureCount]string
	class    string
}

func main() {
	err := run(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func parseMushroomRow(fields []string, rowIndex int) (mushroomSample, error) {
	sample := new(mushroomSample)
	classCode := strings.TrimSpace(fields[classColumnIndex])

	switch classCode {
	case "e":
		sample.class = "edible"
	case "p":
		sample.class = "poisonous"
	default:
		return *sample, fmt.Errorf("mushroom row %d: %w", rowIndex, errUnknownClass)
	}

	for columnIndex := range sample.features {
		feature := strings.TrimSpace(fields[columnIndex+featureColumnOffset])
		if len(feature) != 1 {
			return *sample, fmt.Errorf(
				"mushroom row %d feature %d: %w",
				rowIndex,
				columnIndex+featureColumnOffset,
				errInvalidFeature,
			)
		}

		sample.features[columnIndex] = feature
	}

	return *sample, nil
}

func readMushrooms(input io.Reader) ([]mushroomSample, error) {
	csvReader := csv.NewReader(input)
	csvReader.FieldsPerRecord = featureCount + featureColumnOffset

	samples := make([]mushroomSample, 0, expectedSampleCount)

	for rowIndex := 1; ; rowIndex++ {
		fields, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("read Mushroom row %d: %w", rowIndex, err)
		}

		sample, err := parseMushroomRow(fields, rowIndex)
		if err != nil {
			return nil, err
		}

		samples = append(samples, sample)
	}

	if len(samples) == 0 {
		return nil, errEmptyDataset
	}

	return samples, nil
}

func run(output io.Writer) error {
	samples, err := readMushrooms(bytes.NewReader(rawMushroomData))
	if err != nil {
		return fmt.Errorf("load embedded Mushroom data: %w", err)
	}

	predictor, err := trainPredictor(samples)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "trained: %d samples\n", len(samples))
	if err != nil {
		return fmt.Errorf("write training summary: %w", err)
	}

	queries := [][featureCount]string{
		{"x", "s", "n", "t", "p", "f", "c", "n", "k", "e", "e", "s", "s", "w", "w", "p", "w", "o", "p", "k", "s", "u"},
		{"x", "s", "y", "t", "a", "f", "c", "b", "k", "e", "c", "s", "s", "w", "w", "p", "w", "o", "p", "n", "n", "g"},
		{"x", "y", "b", "t", "n", "f", "c", "b", "e", "e", "?", "s", "s", "e", "w", "p", "w", "t", "e", "w", "c", "w"},
	}

	for _, features := range queries {
		classID, err := predictor.Predict(features[:])
		if err != nil {
			return fmt.Errorf("predict Mushroom class: %w", err)
		}

		className, ok := predictor.GetClass(classID).(string)
		if !ok {
			return fmt.Errorf("resolve class %d: %w", classID, errClassNotString)
		}

		_, err = fmt.Fprintf(
			output,
			"cap=%s, odor=%s, stalk-root=%s, habitat=%s -> %s\n",
			features[0],
			features[4],
			features[10],
			features[21],
			className,
		)
		if err != nil {
			return fmt.Errorf("write prediction: %w", err)
		}
	}

	return nil
}

func trainPredictor(samples []mushroomSample) (*bayes.Predictor, error) {
	predictor, err := bayes.New(bayes.MemoryStorage, datasetScopeID)
	if err != nil {
		return nil, fmt.Errorf("create predictor: %w", err)
	}

	for _, sample := range samples {
		sequence := make([]any, 0, featureCount+1)

		for _, feature := range sample.features {
			sequence = append(sequence, feature)
		}

		sequence = append(sequence, sample.class)

		err := predictor.Train(sequence)
		if err != nil {
			return nil, fmt.Errorf("train predictor: %w", err)
		}
	}

	return predictor, nil
}
