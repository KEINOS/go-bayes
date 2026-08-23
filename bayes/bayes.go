// Package bayes provides Bayesian inference with a Folded Context Transition
// Predictor (FCTP).
//
// A Predictor converts supported values to fixed-width IDs, folds an ordered
// context into one context ID, and learns transitions from that ID to possible
// next-value class IDs. It uses Bayes' theorem to estimate the most likely next
// value for an observed context. It is not a Naive Bayes classifier.
package bayes

import (
	"errors"

	"github.com/KEINOS/go-bayes/bayes/internal/nodeloggers/logmem"
	"github.com/KEINOS/go-bayes/bayes/nodelogger"
)

var (
	errNewOptionNil             = errors.New("new option must not be nil")
	errUnknownStorageEngineType = errors.New("unknown storage engine type")
)

// ----------------------------------------------------------------------------
//  Type: NodeLogger
// ----------------------------------------------------------------------------

// NodeLogger is the node logging interface used by Predictor.
type NodeLogger = nodelogger.NodeLogger

// ----------------------------------------------------------------------------
//  Type: Storage
// ----------------------------------------------------------------------------

// Storage is the type of storage to log the accesses.
type Storage int

const (
	// UnknownStorage represents the unknown storage.
	UnknownStorage Storage = iota
	// MemoryStorage represents the in-memory storage.
	MemoryStorage
)

// Type returns the type name of the storage.
func (s Storage) Type() string {
	switch s {
	case MemoryStorage:
		return "in-memory"
	case UnknownStorage:
	}

	return "unknown"
}

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

// New returns an isolated Predictor using the requested storage and scope. With
// no options, it uses xxHash3 context folding.
func New(engine Storage, scopeID uint64, options ...Option) (*Predictor, error) {
	config := PredictorConfig{
		Storage: engine,
		ScopeID: scopeID,
		Hasher:  nil,
	}

	for _, option := range options {
		if option == nil {
			return nil, errNewOptionNil
		}

		err := option(&config)
		if err != nil {
			return nil, err
		}
	}

	return NewPredictor(config)
}

// newNodeLogger creates a new NodeLogger instance based on the storage engine type.
//
//nolint:ireturn // NodeLogger is the internal storage/logger abstraction.
func newNodeLogger(engine Storage, scopeID uint64) (NodeLogger, error) {
	// Currently only MemoryStorage is supported. So disable switch statement
	//
	// switch engine {
	// case MemoryStorage:
	// 	return logmem.New(scopeID), nil
	// }
	if engine == MemoryStorage {
		return logmem.New(scopeID), nil
	}

	return nil, errUnknownStorageEngineType
}
