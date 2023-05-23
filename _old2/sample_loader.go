package bayes

import (
	"github.com/KEINOS/go-bayes/pkg/utils"
	"github.com/pkg/errors"
)

// ----------------------------------------------------------------------------
//  Type: DataType
// ----------------------------------------------------------------------------

type DataType int

// ----------------------------------------------------------------------------
//  Type: Sample
// ----------------------------------------------------------------------------

type Sample struct {
	Data  [][]uint64
	Class map[uint64]any
}

func (s Sample) GetID(input any) (uint64, error) {
	id, err := utils.AnyToStateID(input)
	if err != nil {
		return 0, errors.Wrap(err, "failed to convert input to state ID")
	}

	if _, ok := s.Class[id]; !ok {
		return 0, errors.New("the input is not in the class/label list: " + input.(string))
	}

	return id, nil
}

// ----------------------------------------------------------------------------
//  Enum: DataType
// ----------------------------------------------------------------------------

const (
	Unknown DataType = iota
	FileCSV
	FileJSON
	SliceString
)

// ----------------------------------------------------------------------------
//  LoadSample()
// ----------------------------------------------------------------------------

// LoadSample returns a Sample object from the given data.
func LoadSample(dataType DataType, target any) (*Sample, error) {
	switch dataType {
	case SliceString:
		return LoadSampleFromSliceString(target)
	case FileCSV:
		return nil, nil
	case FileJSON:
		return nil, nil
	case Unknown:
		return nil, nil
	}

	return nil, nil
}

// ----------------------------------------------------------------------------
//  LoadSampleFromSliceString()
// ----------------------------------------------------------------------------

func loadSampleFromSliceString(sample *Sample, samples []string) error {
	rowData := make([]uint64, len(samples))

	for _, column := range samples {
		id, err := utils.AnyToStateID(column)
		if err != nil {
			return errors.Wrap(err, "failed to convert input to state ID")
		}

		sample.Class[id] = column
		rowData = append(rowData, id)
	}

	sample.Data = append(sample.Data, rowData)

	return nil
}

func LoadSampleFromSliceString(samples any) (*Sample, error) {
	result := &Sample{
		Class: make(map[uint64]any),
		Data:  [][]uint64{},
	}

	// Single row
	if v, ok := samples.([]string); ok {
		if err := loadSampleFromSliceString(result, v); err != nil {
			return nil, errors.Wrap(err, "failed to load sample from slice string")
		}

		return result, nil
	}

	// Multiple rows
	if v, ok := samples.([][]string); ok {
		for _, row := range v {
			if err := loadSampleFromSliceString(result, row); err != nil {
				return nil, errors.Wrap(err, "failed to load sample from slice string")
			}
		}

		return result, nil
	}

	return nil, errors.New("unsupported slice type")
}
