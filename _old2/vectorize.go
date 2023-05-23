package bayes

import (
	"sort"

	"github.com/KEINOS/go-bayes/pkg/utils"
	"github.com/pkg/errors"
)

type (
	// Label is the actual value of the class.
	Label any
	// Scalar is a unique and discrete-valued data in uint64. It is a hash value
	// of the label and its type. Which is used to identify the Label from the Class.
	Scalar uint64
	// Class is the map of Label to Scalar.
	Class map[Scalar]Label
	// Vector is the slice of Scalar/uint64.
	// It represents a set of data to be trained. It is a row in the 2D sample.
	Vector []Scalar
	// Matrix2D is the slice of Vector. Which is the two dimensional matrix that
	// represents the data in row and column. In other words, it is a uint64 table.
	Matrix2D []Vector
)

func (c Class) Label(scalar Scalar) (Label, error) {
	if label, ok := c[scalar]; ok {
		return label, nil
	}

	return nil, errors.Errorf("scalar %v is not registered in the class list", scalar)
}

func (c Class) Scalar(label Label) (Scalar, error) {
	scalar, err := utils.AnyToStateID(label)
	if err != nil {
		return 0, err
	}

	if _, ok := c[Scalar(scalar)]; ok {
		return Scalar(scalar), nil
	}

	return 0, errors.Errorf("label %v is not registered in the class list", label)
}

// Keys returns the sorted slice of Scalar. Which is the list of the keys of the map.
func (c Class) Keys() []Scalar {
	keys := []Scalar{}

	for key := range c {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}

// ToClass returns a Class object from the given input.
//
// The input should be a slice or two dimensional slice.
// E.g.) []any and [][]any.
func ToClass(input any) (Class, error) {
	switch v := input.(type) {
	case []any:
		c := make(Class, len(v))

		for _, label := range v {
			id, err := utils.AnyToStateID(label)
			if err != nil {
				return nil, err
			}
			scalar := Scalar(id)

			c[scalar] = label
		}

		return c, nil
	case [][]any:
		c := make(Class, len(v))

		for _, row := range v {
			for _, column := range row {
				x := column
				id, err := utils.AnyToStateID(x)
				if err != nil {
					return nil, err
				}
				scalar := Scalar(id)

				c[scalar] = column
			}
		}

		return c, nil
	}

	return nil, nil
}

// ----------------------------------------------------------------------------
//  Vectorize functions. ToVector() and ToMatrix2D()
// ----------------------------------------------------------------------------

// ToVector returns a Vector object from the given input.
//
// The input should be a slice. Use this function to convert a row of the table.
// For conversion of a table, use ToMatrix2D.
func ToVector(input []any) (Vector, error) {
	result := make([]Scalar, len(input))

	for i, item := range input {
		id, err := utils.AnyToStateID(item)
		if err != nil {
			return nil, err
		}

		result[i] = Scalar(id)
	}

	return result, nil
}

func ToMatrix2D(input [][]any) (Matrix2D, error) {
	result := make(Matrix2D, len(input))

	for i, row := range input {
		vector, err := ToVector(row)
		if err != nil {
			return nil, err
		}

		result[i] = vector
	}

	return result, nil
}

func Vectorize(input any) ([]uint64, error) {
	switch v := input.(type) {
	case []string:
		return vectorizeSliceString(v)
	case [][]string:
		result := []uint64{}

		for _, row := range v {
			stateIDs, _ := vectorizeSliceString(row)

			result = append(result, stateIDs...)
		}

		return result, nil
	default:
		return nil, errors.New("Unsupported type")
	}
}

func vectorizeSliceString(input []string) ([]uint64, error) {
	result := []uint64{}

	for _, item := range input {
		stateID, err := utils.AnyToStateID(item)
		if err != nil {
			return nil, err
		}

		result = append(result, stateID)
	}

	return result, nil
}
