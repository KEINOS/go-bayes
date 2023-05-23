package utils

import "encoding/binary"

// BytesToStateID returns the state ID from the given byte slice.
//
// It compressses the byte slice to a 8 byte length slice. The first 7 bytes
// stays the same and the last byte is the xor sum of the hole byte slice.
func BytesToStateID(bytes []byte) uint64 {
	var (
		compHash [8]byte
		sum      byte = 0
	)

	copy(compHash[:], bytes[:])

	for _, byteData := range bytes {
		sum ^= byteData
	}

	compHash[7] = sum // set the last byte to the xor sum of the whole byte slice

	return binary.BigEndian.Uint64(compHash[:])
}
