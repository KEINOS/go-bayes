package class

import "github.com/pkg/errors"

// Salt is a salt used to generate the class ID. It depends on the type of the
// input items.
type Salt int

const (
	SaltUnknown Salt = iota
	SaltINT
	SaltINT16
	SaltINT32
	SaltINT64
	SaltUINT
	SaltUINT16
	SaltUINT32
	SaltUINT64
	SaltFLOAT32
	SaltFLOAT64
	SaltSTRING
	SaltBYTES
	SaltBOOL
)

// GetSalt returns the salt value according to the type of the input item.
func GetSalt(input any) (byte, error) {
	switch input.(type) {
	case int:
		return byte(SaltINT), nil
	case int16:
		return byte(SaltINT16), nil
	case int32:
		return byte(SaltINT32), nil
	case int64:
		return byte(SaltINT64), nil
	case uint:
		return byte(SaltUINT), nil
	case uint16:
		return byte(SaltUINT16), nil
	case uint32:
		return byte(SaltUINT32), nil
	case uint64:
		return byte(SaltUINT64), nil
	case float32:
		return byte(SaltFLOAT32), nil
	case float64:
		return byte(SaltFLOAT64), nil
	case string:
		return byte(SaltSTRING), nil
	case []byte:
		return byte(SaltBYTES), nil
	case bool:
		return byte(SaltBOOL), nil
	}

	return byte(SaltUnknown), errors.Errorf("unsupported type: %T", input)
}
