package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math/bits"
	"reflect"

	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
)

func AnyToUint64(value any) (uint64, error) {
	// Any type to hashed byte slice
	b, err := anyToHash(value)
	if err != nil {
		return 0, errors.Wrap(err, "failed to convert the value to uint64")
	}

	// Compress hashed bytes to uint64
	return bytesToUint64(b), nil
}

func anyToBytes(value any) ([]byte, error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	refVal := reflect.ValueOf(value)

	err := enc.Encode(value)

	// Filter unsupported types.
	// pointer and invalid reflect values are not supported.
	if err == nil {
		if refVal.Kind() == reflect.Ptr || !refVal.IsValid() {
			err = errors.Errorf("invalid or unsupported value. Input type: %T", value)
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to gob encode the value")
	}

	return buf.Bytes(), nil
}

func anyToHash(value any) ([]byte, error) {
	b, err := anyToBytes(value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the value to bytes")
	}

	// Use type name as salt for the hash
	salt := fmt.Sprintf("%T", value)

	b = append(b, []byte(salt)...)

	// Hash the bytes
	return bytesToHash(b)
}

func bytesToHash(value []byte) ([]byte, error) {
	h := blake3.New()

	if _, err := h.Write(value); err != nil || value == nil {
		if err == nil {
			err = errors.New("nil value given")
		}

		return nil, errors.Wrap(err, "failed to write the value to the hasher")
	}

	return h.Sum(nil), nil
}

// bytesToUint64 returns the unique ID from the given byte slice.
//
// It compressses the byte slice to an 8 bytes length slice. The first 7 bytes
// stays the same and the last byte is the xor sum of the hole byte slice.
func bytesToUint64(bytes []byte) uint64 {
	var (
		compHash [8]byte
		sum      byte = 0
	)

	copy(compHash[:], bytes[:])

	for _, byteData := range bytes {
		sum = bits.RotateLeft8(sum, 1)
		sum ^= byteData
	}

	compHash[7] = sum // set the last byte to the xor sum of the whole byte slice

	return binary.BigEndian.Uint64(compHash[:])
}
