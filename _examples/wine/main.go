// Package main demonstrates transition prediction with the UCI Wine dataset.
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
	"strconv"
	"strings"

	"github.com/KEINOS/go-bayes/bayes"
)

var (
	errClassNotString = errors.New("predicted class is not a string")
	errEmptyClass     = errors.New("wine class is empty")
	errEmptyDataset   = errors.New("wine dataset is empty")
	errUnknownClass   = errors.New("wine class must be 1, 2, or 3")
)

const (
	classColumnIndex    int    = 0
	datasetScopeID      uint64 = 109
	expectedSampleCount int    = 178
	featureColumnOffset int    = 1
	featureCount        int    = 13
	firstColumnNumber   int    = 1
)

//go:embed wine.data
var rawWineData []byte

type wineSample struct {
	features [featureCount]string
	class    string
}

func main() {
	err := run(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func readWine(input io.Reader) ([]wineSample, error) {
	csvReader := csv.NewReader(input)
	csvReader.FieldsPerRecord = featureCount + 1

	samples := make([]wineSample, 0, expectedSampleCount)

	for rowIndex := 1; ; rowIndex++ {
		fields, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("read Wine row %d: %w", rowIndex, err)
		}

		sample, err := parseWineRow(fields, rowIndex)
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

func parseWineRow(fields []string, rowIndex int) (wineSample, error) {
	sample := new(wineSample)
	classID := strings.TrimSpace(fields[classColumnIndex])

	if classID == "" {
		return *sample, fmt.Errorf("wine row %d: %w", rowIndex, errEmptyClass)
	}

	if classID != "1" && classID != "2" && classID != "3" {
		return *sample, fmt.Errorf("wine row %d: %w", rowIndex, errUnknownClass)
	}

	sample.class = "cultivar " + classID

	for columnIndex := range sample.features {
		value := strings.TrimSpace(fields[columnIndex+featureColumnOffset])

		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			csvColumnNumber := columnIndex + featureColumnOffset + firstColumnNumber

			return *sample, fmt.Errorf(
				"parse Wine row %d column %d: %w",
				rowIndex,
				csvColumnNumber,
				err,
			)
		}

		sample.features[columnIndex] = value
	}

	return *sample, nil
}

func run(output io.Writer) error {
	samples, err := readWine(bytes.NewReader(rawWineData))
	if err != nil {
		return fmt.Errorf("load embedded Wine data: %w", err)
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
		{"14.23", "1.71", "2.43", "15.6", "127", "2.8", "3.06", ".28", "2.29", "5.64", "1.04", "3.92", "1065"},
		{"12.37", ".94", "1.36", "10.6", "88", "1.98", ".57", ".28", ".42", "1.95", "1.05", "1.82", "520"},
		{"12.86", "1.35", "2.32", "18", "122", "1.51", "1.25", ".21", ".94", "4.1", ".76", "1.29", "630"},
	}

	for _, features := range queries {
		classID, err := predictor.Predict(features[:])
		if err != nil {
			return fmt.Errorf("predict Wine class: %w", err)
		}

		className, ok := predictor.GetClass(classID).(string)
		if !ok {
			return fmt.Errorf("resolve class %d: %w", classID, errClassNotString)
		}

		_, err = fmt.Fprintf(
			output,
			"alcohol=%s, color=%s, proline=%s -> %s\n",
			features[0],
			features[9],
			features[12],
			className,
		)
		if err != nil {
			return fmt.Errorf("write prediction: %w", err)
		}
	}

	return nil
}

func trainPredictor(samples []wineSample) (*bayes.Predictor, error) {
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
