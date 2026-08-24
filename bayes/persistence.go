package bayes

import (
	"context"
	"errors"
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/modelstore"
)

var (
	// ErrCommitIndeterminate means SQLite could not confirm whether a commit succeeded.
	ErrCommitIndeterminate = modelstore.ErrCommitIndeterminate
	// ErrInvalidModel means a file is not a valid model for this package.
	ErrInvalidModel = errors.New("invalid go-bayes model")
	// ErrModelLocked means another cooperating process owns the model path.
	ErrModelLocked = errors.New("go-bayes model path is locked")
	// ErrSaveDurabilityUnknown means replacement succeeded but its durable sync did not.
	ErrSaveDurabilityUnknown = errors.New("saved model durability is unknown")
	// ErrStorePoisoned means a store cannot be used after an uncertain commit.
	ErrStorePoisoned = modelstore.ErrPoisoned
)

// Load reads a SQLite model into a new in-memory Predictor.
func Load(ctx context.Context, path string, options ...Option) (*Predictor, error) {
	config := PredictorConfig{
		Storage:           MemoryStorage,
		ScopeID:           0,
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

	if config.SQLitePath != "" || config.ModelStore != nil || config.SQLiteCacheKiB != 0 ||
		config.SQLiteSynchronous != SQLiteSynchronousFull {
		return nil, fmt.Errorf("%w: Load accepts only Hasher selection", errStorageConfig)
	}

	return loadSQLiteModel(ctx, path, config.Hasher)
}

// Open opens an existing SQLite model for direct file-backed operation.
func Open(ctx context.Context, path string, options ...Option) (*Predictor, error) {
	config := PredictorConfig{
		Storage:           SQLiteStorage,
		ScopeID:           0,
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

	if config.SQLitePath != "" || config.ModelStore != nil {
		return nil, fmt.Errorf("%w: Open uses its positional path", errStorageConfig)
	}

	return openSQLiteModel(ctx, path, config)
}
