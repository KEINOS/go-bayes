//go:build cgo

// Package sqlitestore provides the SQLite ModelStore implementation.
package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/KEINOS/go-bayes/bayes/modelstore"
	"github.com/gofrs/flock"
	// Register the cgo SQLite driver used by this build-tagged package.
	_ "github.com/mattn/go-sqlite3"
)

const (
	applicationID = 0x47424159
	schemaVersion = 1
)

var (
	// ErrInvalidModel means that a database is not a valid go-bayes model.
	ErrInvalidModel = errors.New("invalid go-bayes SQLite model")
	// ErrLocked means that another cooperating process owns the model path.
	ErrLocked = errors.New("go-bayes SQLite model is locked")
	//nolint:gochecknoglobals // direct-open ownership needs one process-wide alias registry.
	openRegistry = struct {
		sync.Mutex

		files map[*Store]registeredFile
	}{Mutex: sync.Mutex{}, files: make(map[*Store]registeredFile)}
)

// Metadata identifies the value codec and model scope.
type Metadata struct {
	CodecVersion int64
	HasherName   string
	ItemProbe    uint64
	ContextProbe uint64
	ScopeID      uint64
}

// OpenConfig controls a SQLite connection.
type OpenConfig struct {
	SynchronousNormal bool
	CacheKiB          int
	Portable          bool
}

// PathLock is a cross-process advisory lock for one canonical model path.
type PathLock struct {
	canonical string
	flock     *flock.Flock
}

// Store operates directly on one SQLite model.
type Store struct {
	db       *sql.DB
	conn     *sql.Conn
	commit   func(*sql.Tx) error
	pathLock *PathLock
	path     string
	fileInfo os.FileInfo
	metadata Metadata
	closed   bool
	poisoned bool
}

type registeredFile struct {
	path string
	info os.FileInfo
}

var _ modelstore.ModelStore = (*Store)(nil)

const schemaSQL = `
CREATE TABLE metadata (
    singleton INTEGER PRIMARY KEY CHECK (singleton = 1),
    codec_version INTEGER NOT NULL CHECK (codec_version > 0),
    hasher_name TEXT NOT NULL CHECK (length(hasher_name) > 0),
    item_probe INTEGER NOT NULL,
    context_probe INTEGER NOT NULL,
    scope_id INTEGER NOT NULL,
    total_count INTEGER NOT NULL CHECK (total_count >= 0)
) STRICT;
CREATE TABLE classes (
    id INTEGER PRIMARY KEY,
    type_tag INTEGER NOT NULL CHECK (type_tag BETWEEN 1 AND 255),
    payload BLOB NOT NULL
) STRICT;
CREATE TABLE from_a (
    id INTEGER PRIMARY KEY,
    count INTEGER NOT NULL CHECK (count > 0)
) STRICT;
CREATE TABLE to_b (
    id INTEGER PRIMARY KEY REFERENCES classes(id),
    count INTEGER NOT NULL CHECK (count > 0)
) STRICT;
CREATE TABLE from_a_to_b (
    from_id INTEGER NOT NULL REFERENCES from_a(id),
    to_id INTEGER NOT NULL REFERENCES to_b(id),
    count INTEGER NOT NULL CHECK (count > 0),
    PRIMARY KEY (from_id, to_id)
) STRICT, WITHOUT ROWID;
`

// AcquirePathLock obtains the stable path lock used by Open, Load, and Save.
func AcquirePathLock(ctx context.Context, path string) (*PathLock, error) {
	err := ctx.Err()
	if err != nil {
		return nil, fmt.Errorf("model-path lock canceled: %w", err)
	}

	canonical, err := CanonicalPath(path)
	if err != nil {
		return nil, err
	}

	pathLock := &PathLock{
		canonical: canonical,
		flock:     flock.New(canonical + ".go-bayes.lock"),
	}
	locked, err := pathLock.flock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to lock model path: %w", err)
	}

	if !locked {
		return nil, fmt.Errorf("%w: %s", ErrLocked, canonical)
	}

	return pathLock, nil
}

// CanonicalPath resolves an absolute model path and existing symlinks.
func CanonicalPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: path must not be empty", ErrInvalidModel)
	}

	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("failed to resolve model path: %w", err)
	}

	resolved, resolveErr := filepath.EvalSymlinks(absPath)
	if resolveErr == nil {
		return resolved, nil
	}

	parent, err := filepath.EvalSymlinks(filepath.Dir(absPath))
	if err != nil {
		return "", fmt.Errorf("failed to resolve model directory: %w", err)
	}

	return filepath.Join(parent, filepath.Base(absPath)), nil
}

// Close releases the SQLite connection and path lock.
func (s *Store) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true
	unregister(s)

	var closeErr error
	if s.conn != nil {
		closeErr = errors.Join(closeErr, s.conn.Close())
	}

	if s.db != nil {
		closeErr = errors.Join(closeErr, s.db.Close())
	}

	if s.pathLock != nil {
		closeErr = errors.Join(closeErr, s.pathLock.Close())
	}

	return closeErr
}

// Close releases a path lock.
//
//nolint:wrapcheck // preserve the advisory-lock implementation error.
func (l *PathLock) Close() error {
	if l == nil || l.flock == nil {
		return nil
	}

	return l.flock.Unlock()
}

// Create creates a new empty SQLite model. The target must not exist.
//
//nolint:cyclop,funlen,varnamelen // schema creation keeps cleanup next to each failure boundary.
func Create(ctx context.Context, path string, metadata Metadata, config OpenConfig) (*Store, error) {
	pathLock, err := AcquirePathLock(ctx, path)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(pathLock.canonical)
	if err == nil {
		_ = pathLock.Close()

		return nil, fmt.Errorf("%w: model path already exists", ErrInvalidModel)
	}

	if !errors.Is(err, os.ErrNotExist) {
		_ = pathLock.Close()

		return nil, fmt.Errorf("failed to inspect model path: %w", err)
	}

	store, err := openConnection(ctx, pathLock, true, config)
	if err != nil {
		_ = pathLock.Close()

		return nil, err
	}

	store.metadata = metadata

	tx, err := store.conn.BeginTx(ctx, nil)
	if err != nil {
		_ = store.Close()

		return nil, fmt.Errorf("failed to begin model schema transaction: %w", err)
	}

	rollback := true
	defer func() {
		if rollback {
			_ = tx.Rollback()
		}
	}()

	statements := []string{
		fmt.Sprintf("PRAGMA application_id = %d", applicationID),
		fmt.Sprintf("PRAGMA user_version = %d", schemaVersion),
		schemaSQL,
	}
	for _, statement := range statements {
		_, err = tx.ExecContext(ctx, statement)
		if err != nil {
			_ = store.Close()

			return nil, fmt.Errorf("failed to create model schema: %w", err)
		}
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO metadata (
    singleton, codec_version, hasher_name, item_probe, context_probe, scope_id, total_count
) VALUES (1, ?, ?, ?, ?, ?, 0)`,
		metadata.CodecVersion,
		metadata.HasherName,
		idToSQL(metadata.ItemProbe),
		idToSQL(metadata.ContextProbe),
		idToSQL(metadata.ScopeID),
	)
	if err != nil {
		_ = store.Close()

		return nil, fmt.Errorf("failed to create model metadata: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		store.poisoned = true
		_ = store.Close()

		return nil, fmt.Errorf("%w: %w", modelstore.ErrCommitIndeterminate, err)
	}

	rollback = false

	err = store.refreshFileInfo()
	if err != nil {
		_ = store.Close()

		return nil, err
	}

	register(store)

	return store, nil
}

// IsOpenAlias reports whether path names any file open by this process.
func IsOpenAlias(path string) (bool, error) {
	canonical, err := CanonicalPath(path)
	if err != nil {
		return false, err
	}

	info, statErr := os.Stat(canonical)

	openRegistry.Lock()
	defer openRegistry.Unlock()

	for _, registered := range openRegistry.files {
		if canonical == registered.path {
			return true, nil
		}

		if statErr == nil && registered.info != nil && os.SameFile(info, registered.info) {
			return true, nil
		}
	}

	return false, nil
}

// Metadata returns immutable model metadata.
func (s *Store) Metadata() Metadata {
	return s.metadata
}

// Open opens and validates an existing model with exclusive lifetime ownership.
//
//nolint:varnamelen // tx is the conventional short name for the lock transaction.
func Open(ctx context.Context, path string, config OpenConfig) (*Store, error) {
	pathLock, err := AcquirePathLock(ctx, path)
	if err != nil {
		return nil, err
	}

	store, err := openConnection(ctx, pathLock, false, config)
	if err != nil {
		_ = pathLock.Close()

		return nil, err
	}

	// Force recovery and obtain SQLite's lifetime EXCLUSIVE lock.
	tx, err := store.conn.BeginTx(ctx, nil)
	if err != nil {
		_ = store.Close()

		return nil, fmt.Errorf("failed to lock model database: %w", err)
	}

	err = tx.Rollback()
	if err != nil {
		_ = store.Close()

		return nil, fmt.Errorf("failed to finish model lock transaction: %w", err)
	}

	metadata, err := store.Validate(ctx)
	if err != nil {
		_ = store.Close()

		return nil, err
	}

	store.metadata = metadata

	err = store.refreshFileInfo()
	if err != nil {
		_ = store.Close()

		return nil, err
	}

	register(store)

	return store, nil
}

// Path returns the canonical active database path.
func (s *Store) Path() string {
	return s.path
}

// ScopeID returns the model scope ID.
func (s *Store) ScopeID() uint64 {
	return s.metadata.ScopeID
}

//nolint:funlen,varnamelen // connection configuration keeps SQLite settings in one place.
func openConnection(ctx context.Context, pathLock *PathLock, create bool, config OpenConfig) (*Store, error) {
	mode := "rw"
	if create {
		mode = "rwc"
	}

	journal := "WAL"
	if config.Portable {
		journal = "DELETE"
	}

	synchronous := "FULL"
	if config.SynchronousNormal {
		synchronous = "NORMAL"
	}

	//nolint:exhaustruct // only local-file URI fields are required.
	uri := url.URL{
		Scheme: "file",
		Path:   pathLock.canonical,
	}
	query := uri.Query()
	query.Set("mode", mode)
	query.Set("cache", "private")
	query.Set("_busy_timeout", "0")
	query.Set("_foreign_keys", "on")
	query.Set("_journal_mode", journal)
	query.Set("_locking_mode", "EXCLUSIVE")
	query.Set("_sync", synchronous)
	query.Set("_txlock", "immediate")
	uri.RawQuery = query.Encode()

	db, err := sql.Open("sqlite3", uri.String())
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite driver: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	conn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("failed to open SQLite model: %w", err)
	}

	if config.CacheKiB > 0 {
		_, err = conn.ExecContext(ctx, fmt.Sprintf("PRAGMA cache_size = -%d", config.CacheKiB))
		if err != nil {
			_ = conn.Close()
			_ = db.Close()

			return nil, fmt.Errorf("failed to set SQLite cache size: %w", err)
		}
	}

	return &Store{
		db:       db,
		conn:     conn,
		commit:   (*sql.Tx).Commit,
		pathLock: pathLock,
		path:     pathLock.canonical,
		fileInfo: nil,
		metadata: Metadata{CodecVersion: 0, HasherName: "", ItemProbe: 0, ContextProbe: 0, ScopeID: 0},
		closed:   false,
		poisoned: false,
	}, nil
}

func (s *Store) checkUsable() error {
	if s.closed {
		return modelstore.ErrClosed
	}

	if s.poisoned {
		return modelstore.ErrPoisoned
	}

	return nil
}

func (s *Store) refreshFileInfo() error {
	info, err := os.Stat(s.path)
	if err != nil {
		return fmt.Errorf("failed to stat SQLite model: %w", err)
	}

	s.fileInfo = info

	return nil
}

func register(store *Store) {
	openRegistry.Lock()
	defer openRegistry.Unlock()

	openRegistry.files[store] = registeredFile{path: store.path, info: store.fileInfo}
}

func unregister(store *Store) {
	openRegistry.Lock()
	defer openRegistry.Unlock()

	delete(openRegistry.files, store)
}

func idToSQL(id uint64) int64 {
	return int64(id) // #nosec G115 -- SQLite stores the same 64 bits as signed INTEGER.
}

func idFromSQL(id int64) uint64 {
	return uint64(id) // #nosec G115 -- restore the original unsigned bits.
}

func compareID(left, right uint64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func sortedClasses(classes []modelstore.Class) []modelstore.Class {
	result := make([]modelstore.Class, len(classes))
	for index, class := range classes {
		class.Payload = append([]byte(nil), class.Payload...)
		result[index] = class
	}

	slices.SortFunc(result, func(left, right modelstore.Class) int {
		return compareID(left.ID, right.ID)
	})

	return result
}

//nolint:wrapcheck // preserve the sql.Result implementation error.
func overflowResult(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return modelstore.ErrCountOverflow
	}

	return nil
}
