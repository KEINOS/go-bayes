package utils

import (
	"github.com/pkg/errors"
)

// anyToHash converts single any type value to a hash value.
//
// A salt value is added to the value to hash. This is to avoid collisions when
// hashing the same value with different types.
func anyToHash(item any) ([]byte, error) {
	if item == nil {
		return nil, errors.New("no items to hash. arg is nil")
	}

	salt := AnyToTypeID(item)

	bytes, err := AnyToBytes(item)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash any type value")
	}

	bytes = append(bytes, salt)

	return BytesToHash(bytes), nil
}

// AnyToHash converts any type values to a single hash value.
//
// A salt value is added to the value to hash. This is to avoid collisions when
// hashing the same value with different types.
func AnyToHash(items ...any) ([]byte, error) {
	if len(items) == 0 {
		return nil, errors.New("no items to hash. arg is empty")
	}

	if len(items) == 1 {
		return anyToHash(items[0])
	}

	var (
		bytesToHash = [][]byte{}
		salts       = []byte{}
		saltCurr    = byte(0)
	)

	for i, item := range items {
		if item == nil {
			return nil, errors.New("arg contains nil value")
		}

		salt := AnyToTypeID(item)

		// Convert item to byte slice.
		bytesItem, err := AnyToBytes(item)
		if err != nil {
			return nil, errors.Wrap(err, "failed to hash any type value")
		}

		// Update/append salt value if the type of the item has changed.
		if i == 0 || saltCurr != salt {
			saltCurr = salt
			salts = append(salts, salt)
		}

		bytesToHash = append(bytesToHash, bytesItem)
	}

	// Append salts to the byte slices to be hashed.
	bytesToHash = append(bytesToHash, salts)

	return BytesToHash(bytesToHash...), nil
}
