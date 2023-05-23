/*
Package class defines the Class type and its methods.

Here, Class is equivalent to the item
*/
package class

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
)

// ----------------------------------------------------------------------------
//  Type: Class
// ----------------------------------------------------------------------------

// Class holds the class ID and the original value. It is some what similar to
// key-value pair with any type of value.
type Class struct {
	Raw any    // Raw is the value of the class ID.
	ID  uint64 // ID is the class ID that is used to identify the Raw value.
}

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

// New creates a new Class instance.
func New(id any) (*Class, error) {
	uint64ID, err := AnyToClassID(id)

	if err == nil {
		err = errors.New("unsupported type")

		switch v := id.(type) {
		case int:
			return &Class{ID: uint64ID, Raw: v}, nil
		case int16:
			return &Class{ID: uint64ID, Raw: v}, nil
		case int32:
			return &Class{ID: uint64ID, Raw: v}, nil
		case int64:
			return &Class{ID: uint64ID, Raw: v}, nil
		case uint:
			return &Class{ID: uint64ID, Raw: v}, nil
		case uint16:
			return &Class{ID: uint64ID, Raw: v}, nil
		case uint32:
			return &Class{ID: uint64ID, Raw: v}, nil
		case uint64:
			return &Class{ID: uint64ID, Raw: v}, nil
		case float32:
			return &Class{ID: uint64ID, Raw: v}, nil
		case float64:
			return &Class{ID: uint64ID, Raw: v}, nil
		case string:
			return &Class{ID: uint64ID, Raw: v}, nil
		case bool:
			return &Class{ID: uint64ID, Raw: v}, nil
		case []byte:
			return &Class{ID: uint64ID, Raw: v}, nil
		}
	}

	return nil, errors.Wrap(err, "failed to instantiate Class object")
}

// ----------------------------------------------------------------------------
//  Methods
// ----------------------------------------------------------------------------

// HexID returns the class ID in hexadecimal format.
func (c Class) HexID() string {
	return fmt.Sprintf("%016x", c.ID)
}

// String is the implementation of the Stringer interface.
// It retruns the raw value of the class in string format.
func (c Class) String() string {
	return fmt.Sprintf("%v", c.Raw)
}

// ----------------------------------------------------------------------------
//  Functions
// ----------------------------------------------------------------------------

// BytesToClassID returns the class ID from the given byte slice.
//
// It compressses the byte slice to a 8 byte length slice. The first 7 bytes
// stays the same and the last byte is the xor sum of the hole byte slice.
func BytesToClassID(b []byte) uint64 {
	var (
		compHash [8]byte
		sum      byte = 0
	)

	for i, byteData := range b {
		sum ^= byteData
		if i < 7 {
			compHash[i] = byteData

			continue
		}
	}

	compHash[7] = sum

	return binary.BigEndian.Uint64(compHash[:])
}

func AnyToBytes(a any) ([]byte, error) {
	switch v := a.(type) {
	case int:
		return IntToBytes(int64(v)), nil
	case int16:
		return IntToBytes(int64(v)), nil
	case int32:
		return IntToBytes(int64(v)), nil
	case int64:
		return IntToBytes(int64(v)), nil
	case uint:
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

// AnyToClassID returns the class ID of the item given in uint64.
func AnyToClassID(item any) (uint64, error) {
	bytesItem, err := AnyToHash(item)
	if err != nil {
		return 0, err
	}

	return BytesToClassID(bytesItem), nil
}

// AnyToHash returns the hash from the value given in byte slice.
func AnyToHash(a any) (hashed []byte, err error) {
	b, err1 := AnyToBytes(a)
	salt, err2 := GetSalt(a)

	if err1 == nil && err2 == nil {
		b = append(b, salt)

		h := blake3.Sum512(b)

		return h[:], nil
	}

	if err1 != nil {
		err = err1
	}

	if err2 != nil {
		err = errors.Wrap(err2, err.Error())
	}

	return []byte{}, errors.Wrap(err, "failed to hash from any type")
}

// IntToBytes converts an integer to a byte slice in big endian format.
func IntToBytes(num int64) []byte {
	size := int(unsafe.Sizeof(num))
	arr := make([]byte, size)

	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
		arr[size-i-1] = byt
	}

	return arr
}
