package utils

import (
	"encoding/binary"
	"math"

	"github.com/pkg/errors"
)

// AnyToBytes converts an any type value to a byte slice.
//
// The minimum length of the returned byte slice is [8]byte.
//
//nolint: funlen,gocyclo,cyclop // ignore function length and complexity here
func AnyToBytes(a any) ([]byte, error) {
	switch v := a.(type) {
	case int:
		return IntToBytes(int64(v)), nil
	case int8:
		return IntToBytes(int64(v)), nil
	case int16:
		return IntToBytes(int64(v)), nil
	case int32:
		return IntToBytes(int64(v)), nil
	case int64:
		return IntToBytes(v), nil
	case uint:
		return IntToBytes(int64(v)), nil
	case uint8:
		return IntToBytes(int64(v)), nil
	case uint16:
		return IntToBytes(int64(v)), nil
	case uint32:
		return IntToBytes(int64(v)), nil
	case uint64:
		return IntToBytes(int64(v)), nil
	case float32:
		var buf [8]byte

		binary.BigEndian.PutUint64(buf[:], math.Float64bits(float64(v)))

		return buf[:], nil
	case float64:
		var buf [8]byte

		binary.BigEndian.PutUint64(buf[:], math.Float64bits(v))

		return buf[:], nil
	case string:
		return []byte(v), nil
	case []string:
		buf := []byte{}

		for _, vv := range v {
			buf = append(buf, []byte(vv)...)
		}

		return buf, nil
	case []int:
		buf := []byte{}

		for _, vv := range v {
			b := IntToBytes(int64(vv))

			buf = append(buf, b...)
		}

		return buf, nil
	case []byte:
		return v, nil
	case bool:
		if v {
			vv := ^uint64(0)

			return IntToBytes(int64(vv)), nil
		}

		return IntToBytes(int64(0)), nil
	}

	return []byte{}, errors.Errorf("failed to convert to byte slice. Unsupported type: %T", a)
}

// func isSlice(a any) bool {
// 	return reflect.TypeOf(a).Kind() == reflect.Slice
// }
