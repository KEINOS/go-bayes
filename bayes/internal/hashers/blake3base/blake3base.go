// Package blake3base provides the BLAKE3 transition hasher implementation.
package blake3base

import (
	"encoding/binary"
	"errors"
	"hash/crc32"

	"github.com/KEINOS/go-bayes/bayes/hasher"
	"github.com/zeebo/blake3"
)

var errInvalidBytesLength = errors.New(
	"failed to combine bytes. Both of the input must be 4byte or more",
)

// Compile-time check that *Hasher satisfies hasher.TransitionHasher.
var _ hasher.TransitionHasher = (*Hasher)(nil)

// Hasher implements the BLAKE3 transition hashing algorithm.
type Hasher struct{}

// New returns a new BLAKE3 transition hasher.
func New() *Hasher {
	return &Hasher{}
}

// HashTrans returns a unique hash from the input transitions.
func (h *Hasher) HashTrans(transitions ...uint64) (uint64, error) {
	hashed := getBlake3(transitions...)

	chksum := getCRC32C(hashed)

	return chopAndMergeBytes(hashed, chksum)
}

func chopAndMergeBytes(bytesA, bytesB []byte) (uint64, error) {
	if len(bytesA) < 4 || len(bytesB) < 4 {
		return 0, errInvalidBytesLength
	}

	const lenOut = 8

	rawid := make([]byte, lenOut)

	_ = copy(rawid, bytesA)
	_ = copy(rawid[4:], bytesB)

	return binary.BigEndian.Uint64(rawid), nil
}

func getBlake3(inputs ...uint64) []byte {
	hasher := blake3.New()

	for _, v := range inputs {
		_, _ = hasher.Write(uint64ToByteArray(v))
	}

	return hasher.Sum(nil)
}

func getCRC32C(input []byte) []byte {
	crcTable := crc32.MakeTable(crc32.Castagnoli)
	crc32C := crc32.New(crcTable)

	_, _ = crc32C.Write(input)

	return crc32C.Sum(nil)
}

func uint64ToByteArray(num uint64) []byte {
	const lenUint64Bytes = 8

	arr := make([]byte, lenUint64Bytes)
	binary.LittleEndian.PutUint64(arr, num)

	return arr
}
