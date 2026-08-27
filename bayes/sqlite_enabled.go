//go:build cgo

package bayes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"

	"github.com/KEINOS/go-bayes/bayes/internal/modelstores/mapstore"
	"github.com/KEINOS/go-bayes/bayes/internal/modelstores/sqlitestore"
	"github.com/KEINOS/go-bayes/bayes/modelstore"
)

const modelCodecVersion = 1

type saveDependencies struct {
	acquirePathLock func(context.Context, string) (io.Closer, error)
	canonicalPath   func(string) (string, error)
	createStore     func(
		context.Context,
		string,
		sqlitestore.Metadata,
		sqlitestore.OpenConfig,
	) (temporaryModelStore, error)
	createTemp    func(string, string) (saveFile, error)
	isOpenAlias   func(string) (bool, error)
	openFile      func(string, int, os.FileMode) (saveFile, error)
	remove        func(string) error
	rename        func(string, string) error
	stat          func(string) (os.FileInfo, error)
	syncDirectory func(string) error
}

type saveDestination struct {
	lock      io.Closer
	directory string
	path      string
	mode      os.FileMode
}

type saveFile interface {
	io.Closer
	Chmod(mode os.FileMode) error
	Name() string
	Sync() error
}

type temporaryModelStore interface {
	Close() error
	Import(ctx context.Context, classes []modelstore.Class, source modelstore.ModelStore) error
	Validate(ctx context.Context) (sqlitestore.Metadata, error)
}

//nolint:funlen // load validates and transfers one complete model with explicit cleanup.
func loadSQLiteModel(ctx context.Context, path string, injected Hasher) (*Predictor, error) {
	source, err := sqlitestore.Open(ctx, path, sqlitestore.OpenConfig{
		SynchronousNormal: false,
		CacheKiB:          0,
		Portable:          true,
	})
	if err != nil {
		return nil, wrapSQLiteOpenError(err)
	}
	defer func() { _ = source.Close() }()

	hasher, err := resolveModelHasher(source.Metadata(), injected)
	if err != nil {
		return nil, err
	}

	classes, err := source.Classes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read model classes: %w", err)
	}

	records := []modelstore.TransitionCount{}

	err = source.ExportTransitions(ctx, func(record modelstore.TransitionCount) error {
		records = append(records, record)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read model transitions: %w", err)
	}

	memoryStore := mapstore.New(source.ScopeID())
	batch := modelstore.TrainingBatch{
		Classes: classes,
		Transitions: func() iter.Seq[modelstore.TransitionDelta] {
			return func(yield func(modelstore.TransitionDelta) bool) {
				for _, record := range records {
					if !yield(modelstore.TransitionDelta(record)) {
						return
					}
				}
			}
		},
	}

	err = memoryStore.Apply(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("failed to restore model in memory: %w", err)
	}

	predictor, err := NewPredictor(ctx, PredictorConfig{
		Storage:           UnknownStorage,
		ScopeID:           source.ScopeID(),
		Hasher:            hasher,
		ModelStore:        memoryStore,
		SQLitePath:        "",
		SQLiteSynchronous: SQLiteSynchronousFull,
		SQLiteCacheKiB:    0,
	})
	if err != nil {
		_ = memoryStore.Close()

		return nil, err
	}

	predictor.storage = MemoryStorage

	return predictor, nil
}

//nolint:ireturn // the build-tagged factory implements the public storage extension point.
func newSQLiteStore(ctx context.Context, config PredictorConfig) (ModelStore, error) {
	if config.SQLitePath == "" {
		return nil, fmt.Errorf("%w: SQLiteStorage requires WithSQLitePath", errStorageConfig)
	}

	_, err := os.Stat(config.SQLitePath)
	if err == nil {
		return nil, fmt.Errorf("%w: SQLite model already exists", errStorageConfig)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to inspect SQLite model path: %w", err)
	}

	metadata := metadataFor(config.Hasher, config.ScopeID)

	store, err := sqlitestore.Create(ctx, config.SQLitePath, metadata, sqliteOpenConfig(config, false))
	if err != nil {
		return nil, wrapSQLiteOpenError(err)
	}

	return store, nil
}

func openSQLiteModel(ctx context.Context, path string, config PredictorConfig) (*Predictor, error) {
	store, err := sqlitestore.Open(ctx, path, sqliteOpenConfig(config, false))
	if err != nil {
		return nil, wrapSQLiteOpenError(err)
	}

	hasher, err := resolveModelHasher(store.Metadata(), config.Hasher)
	if err != nil {
		_ = store.Close()

		return nil, err
	}

	predictor, err := NewPredictor(ctx, PredictorConfig{
		Storage:           UnknownStorage,
		ScopeID:           store.ScopeID(),
		Hasher:            hasher,
		ModelStore:        store,
		SQLitePath:        "",
		SQLiteSynchronous: SQLiteSynchronousFull,
		SQLiteCacheKiB:    0,
	})
	if err != nil {
		_ = store.Close()

		return nil, err
	}

	predictor.storage = SQLiteStorage
	predictor.sqlitePath = store.Path()

	return predictor, nil
}

func defaultSaveDependencies() saveDependencies {
	return saveDependencies{
		acquirePathLock: func(ctx context.Context, path string) (io.Closer, error) {
			return sqlitestore.AcquirePathLock(ctx, path)
		},
		canonicalPath: sqlitestore.CanonicalPath,
		createStore: func(
			ctx context.Context,
			path string,
			metadata sqlitestore.Metadata,
			config sqlitestore.OpenConfig,
		) (temporaryModelStore, error) {
			return sqlitestore.Create(ctx, path, metadata, config)
		},
		createTemp: func(directory, pattern string) (saveFile, error) {
			return os.CreateTemp(directory, pattern)
		},
		isOpenAlias: sqlitestore.IsOpenAlias,
		openFile: func(path string, flag int, mode os.FileMode) (saveFile, error) {
			return os.OpenFile(path, flag, mode) // #nosec G304 -- path comes from CreateTemp above.
		},
		remove:        os.Remove,
		rename:        os.Rename,
		stat:          os.Stat,
		syncDirectory: syncDirectory,
	}
}

func exportTemporaryModel(
	ctx context.Context,
	predictor *Predictor,
	temporaryPath string,
	deps saveDependencies,
) error {
	temporaryStore, err := deps.createStore(
		ctx,
		temporaryPath,
		metadataFor(predictor.hasher, predictor.ID()),
		sqlitestore.OpenConfig{SynchronousNormal: false, CacheKiB: 0, Portable: true},
	)
	if err != nil {
		return fmt.Errorf("failed to create temporary model database: %w", err)
	}

	classes, err := predictor.store.Classes(ctx)
	if err == nil {
		err = temporaryStore.Import(ctx, classes, predictor.store)
	}
	if err == nil {
		_, err = temporaryStore.Validate(ctx)
	}

	closeErr := temporaryStore.Close()
	if err != nil {
		return fmt.Errorf("failed to export model: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close temporary model: %w", closeErr)
	}

	return nil
}

func prepareSaveDestination(
	ctx context.Context,
	path string,
	deps saveDependencies,
) (saveDestination, error) {
	if path == "" {
		return saveDestination{}, fmt.Errorf("%w: save path must not be empty", errStorageConfig)
	}

	destinationPath, err := deps.canonicalPath(path)
	if err != nil {
		return saveDestination{}, fmt.Errorf("failed to resolve save destination: %w", err)
	}

	destinationLock, err := deps.acquirePathLock(ctx, destinationPath)
	if err != nil {
		return saveDestination{}, wrapSQLiteOpenError(err)
	}

	active, err := deps.isOpenAlias(destinationPath)
	if err != nil {
		_ = destinationLock.Close()

		return saveDestination{}, fmt.Errorf("failed to inspect active model paths: %w", err)
	}
	if active {
		_ = destinationLock.Close()

		return saveDestination{}, fmt.Errorf("%w: destination is an active model", ErrModelLocked)
	}

	mode := os.FileMode(fileModePrivate)
	info, statErr := deps.stat(destinationPath)
	if statErr == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(statErr, os.ErrNotExist) {
		_ = destinationLock.Close()

		return saveDestination{}, fmt.Errorf("failed to inspect save destination: %w", statErr)
	}

	return saveDestination{
		lock:      destinationLock,
		directory: filepath.Dir(destinationPath),
		path:      destinationPath,
		mode:      mode,
	}, nil
}

func prepareTemporaryModelPath(directory string, deps saveDependencies) (string, error) {
	temporary, err := deps.createTemp(directory, ".go-bayes-*.db")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary model: %w", err)
	}

	temporaryPath := filepath.Clean(temporary.Name())

	err = temporary.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close temporary model placeholder: %w", err)
	}

	err = deps.remove(temporaryPath)
	if err != nil {
		return "", fmt.Errorf("failed to prepare temporary model path: %w", err)
	}

	return temporaryPath, nil
}

func replaceSavedModel(temporaryPath string, destination saveDestination, deps saveDependencies) error {
	err := deps.rename(temporaryPath, destination.path)
	if err != nil {
		return fmt.Errorf("failed to replace saved model: %w", err)
	}

	err = deps.syncDirectory(destination.directory)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSaveDurabilityUnknown, err)
	}

	return nil
}

func saveModel(ctx context.Context, predictor *Predictor, path string) error {
	return saveModelWithDependencies(ctx, predictor, path, defaultSaveDependencies())
}

func saveModelWithDependencies(
	ctx context.Context,
	predictor *Predictor,
	path string,
	deps saveDependencies,
) error {
	if predictor.store == nil {
		return errPredictorNotInitialized
	}

	destination, err := prepareSaveDestination(ctx, path, deps)
	if err != nil {
		return err
	}
	defer func() { _ = destination.lock.Close() }()

	temporaryPath, err := prepareTemporaryModelPath(destination.directory, deps)
	if err != nil {
		return err
	}
	defer func() {
		_ = deps.remove(temporaryPath)
		_ = deps.remove(temporaryPath + ".go-bayes.lock")
	}()

	err = exportTemporaryModel(ctx, predictor, temporaryPath, deps)
	if err != nil {
		return err
	}

	err = syncSavedModelFile(temporaryPath, destination.mode, deps)
	if err != nil {
		return err
	}

	return replaceSavedModel(temporaryPath, destination, deps)
}

func syncSavedModelFile(path string, mode os.FileMode, deps saveDependencies) error {
	temporaryFile, err := deps.openFile(path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open temporary model for sync: %w", err)
	}

	chmodErr := temporaryFile.Chmod(mode)
	if chmodErr != nil {
		return fmt.Errorf("failed to set saved model permissions: %w", errors.Join(chmodErr, temporaryFile.Close()))
	}

	syncErr := temporaryFile.Sync()
	closeErr := temporaryFile.Close()
	if syncErr != nil || closeErr != nil {
		return fmt.Errorf("failed to sync temporary model: %w", errors.Join(syncErr, closeErr))
	}

	return nil
}

func metadataFor(hasher Hasher, scopeID uint64) sqlitestore.Metadata {
	return sqlitestore.Metadata{
		CodecVersion: modelCodecVersion,
		HasherName:   hasher.Name(),
		ItemProbe:    hasher.Hash(itemHasherProbeBytes()),
		ContextProbe: hasher.Hash(contextHasherProbeBytes()),
		ScopeID:      scopeID,
	}
}

//nolint:ireturn // caller-selected and built-in hashers share the public interface.
func resolveModelHasher(metadata sqlitestore.Metadata, injected Hasher) (Hasher, error) {
	if metadata.CodecVersion != modelCodecVersion {
		return nil, fmt.Errorf("%w: codec version %d", ErrInvalidModel, metadata.CodecVersion)
	}

	hasher := injected
	if hasher == nil {
		var builtIn bool

		hasher, builtIn = newBuiltInHasher(metadata.HasherName)
		if !builtIn {
			return nil, fmt.Errorf("%w: custom hasher %q must be injected", ErrInvalidModel, metadata.HasherName)
		}
	}

	if hasher.Name() != metadata.HasherName ||
		hasher.Hash(itemHasherProbeBytes()) != metadata.ItemProbe ||
		hasher.Hash(contextHasherProbeBytes()) != metadata.ContextProbe {
		return nil, fmt.Errorf("%w: model hasher does not match", ErrInvalidModel)
	}

	return hasher, nil
}

func contextHasherProbeBytes() []byte {
	return []byte{contextDomain, 0}
}

func itemHasherProbeBytes() []byte {
	return []byte{itemDomain, tagString, 'g', 'o', '-', 'b', 'a', 'y', 'e', 's'}
}

func sqliteOpenConfig(config PredictorConfig, portable bool) sqlitestore.OpenConfig {
	return sqlitestore.OpenConfig{
		SynchronousNormal: config.SQLiteSynchronous == SQLiteSynchronousNormal,
		CacheKiB:          config.SQLiteCacheKiB,
		Portable:          portable,
	}
}

func wrapSQLiteOpenError(err error) error {
	switch {
	case errors.Is(err, sqlitestore.ErrInvalidModel):
		return fmt.Errorf("%w: %w", ErrInvalidModel, err)
	case errors.Is(err, sqlitestore.ErrLocked):
		return fmt.Errorf("%w: %w", ErrModelLocked, err)
	default:
		return err
	}
}
