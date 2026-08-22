package bayes

import (
	"errors"
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/hasher"
	"github.com/KEINOS/go-bayes/bayes/internal/hashers/blake3base"
	"github.com/KEINOS/go-bayes/bayes/internal/hashers/xxHash3base"
)

var errUnknownHasher = errors.New("unknown hasher")

// Hasher hashes transition IDs into a flow ID.
type Hasher = hasher.TransitionHasher

// Option configures a Predictor created by New.
type Option func(*PredictorConfig) error

// NewBlake3Hasher returns the BLAKE3 transition hasher.
//
//nolint:ireturn // returning the public extension interface hides internal implementations.
func NewBlake3Hasher() Hasher {
	return blake3base.New()
}

// NewDefaultHasher returns the current default hasher implementation.
//
//nolint:ireturn // returning interface is intentional for algorithm swapping.
func NewDefaultHasher() Hasher {
	return NewXXHash3Hasher()
}

// NewXXHash3Hasher returns the xxHash3 transition hasher.
//
//nolint:ireturn // returning the public extension interface hides internal implementations.
func NewXXHash3Hasher() Hasher {
	return xxhash3base.New()
}

// WithHasher selects the transition hasher used by a Predictor.
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
