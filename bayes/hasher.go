package bayes

import (
	"errors"
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/internal/hashers/blake3base"
	"github.com/KEINOS/go-bayes/bayes/internal/hashers/xxHash3base"
)

var errUnknownHasher = errors.New("unknown hasher")

var (
	_ Hasher = blake3base.New()
	_ Hasher = xxhash3base.New()
)

// Hasher converts canonical bytes into one deterministic, fixed-width ID.
// Name must return a stable, non-empty name. Hash must not retain or mutate
// data.
type Hasher interface {
	Name() string
	Hash(data []byte) uint64
}

// Option configures a Predictor created by New.
type Option func(*PredictorConfig) error

// NewBlake3Hasher returns the BLAKE3 value and context hasher.
//
//nolint:ireturn // returning the public extension interface hides internal implementations.
func NewBlake3Hasher() Hasher {
	return blake3base.New()
}

// NewDefaultHasher returns the xxHash3 value and context hasher.
//
//nolint:ireturn // returning interface is intentional for algorithm swapping.
func NewDefaultHasher() Hasher {
	return NewXXHash3Hasher()
}

// NewXXHash3Hasher returns the xxHash3 value and context hasher.
//
//nolint:ireturn // returning the public extension interface hides internal implementations.
func NewXXHash3Hasher() Hasher {
	return xxhash3base.New()
}

// WithHasher selects the value and context hasher used by a Predictor.
// Supported names are "blake3" and "xxhash3".
func WithHasher(name string) Option {
	return func(config *PredictorConfig) error {
		switch name {
		case "blake3":
			config.Hasher = NewBlake3Hasher()
		case "xxhash3":
			config.Hasher = NewXXHash3Hasher()
		default:
			return fmt.Errorf("%w: %q", errUnknownHasher, name)
		}

		return nil
	}
}

// newBuiltInHasher returns a built-in hasher for a persistence name.
//
//nolint:ireturn // the public interface keeps implementations internal.
func newBuiltInHasher(name string) (Hasher, bool) {
	switch name {
	case "blake3":
		return NewBlake3Hasher(), true
	case "xxhash3":
		return NewXXHash3Hasher(), true
	default:
		return nil, false
	}
}
