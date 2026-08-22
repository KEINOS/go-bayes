// Package bayes provides Bayesian sequence prediction utilities.
package bayes

import (
	"errors"

	"github.com/KEINOS/go-bayes/bayes/internal/nodeloggers/logmem"
	"github.com/KEINOS/go-bayes/bayes/nodelogger"
)

var errUnknownStorageEngineType = errors.New("unknown storage engine type")

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

// New returns a new Predictor instance.
func New(engine Storage, scopeID uint64) (*Predictor, error) {
	return NewPredictor(PredictorConfig{
		Storage: engine,
		ScopeID: scopeID,
		Hasher:  nil,
	})
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
