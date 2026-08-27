//go:build cgo

//nolint:exhaustruct // tests set only fields relevant to each failure.
package sqlitestore

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/KEINOS/go-bayes/bayes/modelstore"
	"github.com/stretchr/testify/require"
)

var errTestCreateFailed = errors.New("injected create failure")

type createTransactionStub struct {
	commitErr     error
	execCalls     int
	failExecAt    int
	rollbackCalls int
}

func (s *createTransactionStub) Commit() error {
	return s.commitErr
}

func (s *createTransactionStub) ExecContext(
	context.Context,
	string,
	...any,
) (sql.Result, error) {
	s.execCalls++
	if s.execCalls == s.failExecAt {
		return nil, errTestCreateFailed
	}

	return driver.ResultNoRows, nil
}

func (s *createTransactionStub) Rollback() error {
	s.rollbackCalls++

	return nil
}

//nolint:funlen // the table verifies every create-stage cleanup boundary.
func TestCreateWithDependencies_releasesResourcesAfterFailure(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configure     func(*createDependencies, *createTransactionStub)
		wantPoisoned  bool
		wantRollback  int
		wantStoreOpen bool
	}{
		"inspect path": {
			configure: func(deps *createDependencies, _ *createTransactionStub) {
				deps.stat = func(string) (os.FileInfo, error) { return nil, errTestCreateFailed }
			},
		},
		"open connection": {
			configure: func(deps *createDependencies, _ *createTransactionStub) {
				deps.openConnection = func(
					context.Context, *PathLock, bool, OpenConfig,
				) (*Store, error) {
					return nil, errTestCreateFailed
				}
			},
		},
		"begin transaction": {
			configure: func(deps *createDependencies, _ *createTransactionStub) {
				deps.beginTransaction = func(context.Context, *Store) (createTransaction, error) {
					var transaction *sql.Tx

					return transaction, errTestCreateFailed
				}
			},
			wantStoreOpen: true,
		},
		"create schema": {
			configure: func(_ *createDependencies, transaction *createTransactionStub) {
				transaction.failExecAt = 1
			},
			wantRollback:  1,
			wantStoreOpen: true,
		},
		"create metadata": {
			configure: func(_ *createDependencies, transaction *createTransactionStub) {
				transaction.failExecAt = 4
			},
			wantRollback:  1,
			wantStoreOpen: true,
		},
		"commit transaction": {
			configure: func(_ *createDependencies, transaction *createTransactionStub) {
				transaction.commitErr = errTestCreateFailed
			},
			wantPoisoned:  true,
			wantRollback:  1,
			wantStoreOpen: true,
		},
		"refresh file information": {
			configure: func(deps *createDependencies, _ *createTransactionStub) {
				deps.refreshFileInfo = func(*Store) error { return errTestCreateFailed }
			},
			wantStoreOpen: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "model.db")
			transaction := new(createTransactionStub)
			deps := defaultCreateDependencies()
			var createdStore *Store

			deps.openConnection = func(
				_ context.Context,
				pathLock *PathLock,
				_ bool,
				_ OpenConfig,
			) (*Store, error) {
				createdStore = &Store{pathLock: pathLock, path: pathLock.canonical}

				return createdStore, nil
			}
			deps.beginTransaction = func(context.Context, *Store) (createTransaction, error) {
				return transaction, nil
			}
			deps.refreshFileInfo = func(*Store) error { return nil }
			deps.register = func(*Store) { require.Fail(t, "failed create must not register a store") }
			test.configure(&deps, transaction)

			store, err := createWithDependencies(ctx, path, testMetadata(), OpenConfig{}, deps)
			require.ErrorIs(t, err, errTestCreateFailed)
			require.Nil(t, store)
			require.Equal(t, test.wantRollback, transaction.rollbackCalls)

			if test.wantStoreOpen {
				require.NotNil(t, createdStore)
				require.True(t, createdStore.closed)
				require.Equal(t, test.wantPoisoned, createdStore.poisoned)
			}

			lock, lockErr := AcquirePathLock(ctx, path)
			require.NoError(t, lockErr, "the failed create must release its path lock")
			require.NoError(t, lock.Close())
		})
	}
}

func TestDefaultCreateDependencies_returnsNilTransactionAfterBeginFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })

	connection, err := database.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, connection.Close())

	store := new(Store)
	store.conn = connection

	transaction, err := defaultCreateDependencies().beginTransaction(ctx, store)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.Nil(t, transaction)
}

func TestCreateWithDependencies_reportsIndeterminateCommit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "model.db")
	transaction := &createTransactionStub{commitErr: errTestCreateFailed}
	deps := defaultCreateDependencies()
	deps.openConnection = func(
		_ context.Context,
		pathLock *PathLock,
		_ bool,
		_ OpenConfig,
	) (*Store, error) {
		return &Store{pathLock: pathLock, path: pathLock.canonical}, nil
	}
	deps.beginTransaction = func(context.Context, *Store) (createTransaction, error) {
		return transaction, nil
	}

	_, err := createWithDependencies(ctx, path, testMetadata(), OpenConfig{}, deps)
	require.ErrorIs(t, err, modelstore.ErrCommitIndeterminate)
	require.ErrorIs(t, err, errTestCreateFailed)
}
