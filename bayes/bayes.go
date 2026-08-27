// Package bayes provides Bayesian inference with a Folded Context Transition
// Predictor (FCTP).
//
// A Predictor converts supported values to fixed-width IDs, folds an ordered
// context into one context ID, and learns transitions from that ID to possible
// next-value class IDs. It uses Bayes' theorem to estimate the most likely next
// value for an observed context. It is not a Naive Bayes classifier.
package bayes

import (
	"context"
	"errors"
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/internal/modelstores/mapstore"
	"github.com/KEINOS/go-bayes/bayes/modelstore"
)

var (
	// ErrJSONPersistenceUnsupported tells callers to use Save, Load, or Open.
	ErrJSONPersistenceUnsupported = errors.New("JSON model persistence is unsupported; use Save, Load, or Open")
	// ErrHashCollision means one class ID was produced for two different values.
	ErrHashCollision = modelstore.ErrClassConflict
	// ErrPredictorClosed means that an I/O method was called after Close.
	ErrPredictorClosed = errors.New("predictor is closed")
	// ErrSQLiteUnavailable means that this build does not include cgo SQLite support.
	ErrSQLiteUnavailable = errors.New("SQLite model storage is unavailable in this build")
	errNewOptionNil      = errors.New("new option must not be nil")
	errStorageConfig     = errors.New("invalid storage configuration")
)

const (
	fileModePrivate = 0o600
)

// ModelStore keeps exact model counts and reversible class records.
type ModelStore = modelstore.ModelStore

// SQLiteSynchronous controls SQLite's durability and write latency.
type SQLiteSynchronous int

const (
	// SQLiteSynchronousFull syncs each committed transaction for durability.
	SQLiteSynchronousFull SQLiteSynchronous = iota
	// SQLiteSynchronousNormal reduces sync work and can lose recent commits after power loss.
	SQLiteSynchronousNormal
)

// Storage selects a built-in model store.
type Storage int

const (
	// UnknownStorage is used when a caller injects a ModelStore.
	UnknownStorage Storage = iota
	// MemoryStorage keeps the complete model in Go maps.
	MemoryStorage
	// SQLiteStorage keeps the model in a SQLite file.
	SQLiteStorage
)

// New returns an isolated Predictor using the requested storage and scope.
// With no Hasher option, it uses xxHash3 for value and context IDs.
func New(ctx context.Context, engine Storage, scopeID uint64, options ...Option) (*Predictor, error) {
	config := PredictorConfig{
		Storage:           engine,
		ScopeID:           scopeID,
		Hasher:            nil,
		ModelStore:        nil,
		SQLitePath:        "",
		SQLiteSynchronous: SQLiteSynchronousFull,
		SQLiteCacheKiB:    0,
	}

	err := applyOptions(&config, options)
	if err != nil {
		return nil, err
	}

	return NewPredictor(ctx, config)
}

// Type returns a short storage name.
func (s Storage) Type() string {
	switch s {
	case MemoryStorage:
		return "in-memory"
	case SQLiteStorage:
		return "sqlite"
	case UnknownStorage:
	}

	return "unknown"
}

func applyOptions(config *PredictorConfig, options []Option) error {
	for _, option := range options {
		if option == nil {
			return errNewOptionNil
		}

		err := option(config)
		if err != nil {
			return err
		}
	}

	return nil
}

// newModelStore creates a configured built-in store or accepts an injected one.
//
//nolint:cyclop,ireturn // constructor validation is explicit; the interface is the extension boundary.
func newModelStore(ctx context.Context, config PredictorConfig) (modelstore.ModelStore, error) {
	err := ctx.Err()
	if err != nil {
		return nil, fmt.Errorf("model-store construction canceled: %w", err)
	}

	if config.ModelStore != nil {
		if config.Storage != UnknownStorage || config.SQLitePath != "" {
			return nil, fmt.Errorf("%w: injected ModelStore requires UnknownStorage", errStorageConfig)
		}

		if config.ModelStore.ScopeID() != config.ScopeID {
			return nil, fmt.Errorf("%w: ModelStore scope does not match config", errStorageConfig)
		}

		return config.ModelStore, nil
	}

	switch config.Storage {
	case MemoryStorage:
		if config.SQLitePath != "" || config.SQLiteSynchronous != SQLiteSynchronousFull || config.SQLiteCacheKiB != 0 {
			return nil, fmt.Errorf("%w: memory storage rejects SQLite options", errStorageConfig)
		}

		return mapstore.New(config.ScopeID), nil
	case SQLiteStorage:
		return newSQLiteStore(ctx, config)
	case UnknownStorage:
		return nil, fmt.Errorf("%w: no ModelStore was supplied", errStorageConfig)
	default:
		return nil, fmt.Errorf("%w: unknown storage %d", errStorageConfig, config.Storage)
	}
}
