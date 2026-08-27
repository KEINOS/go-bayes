//go:build cgo

package bayes

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/KEINOS/go-bayes/bayes/internal/modelstores/sqlitestore"
	"github.com/KEINOS/go-bayes/bayes/modelstore"
	"github.com/stretchr/testify/require"
)

var errTestSaveFailed = errors.New("save failed")

const (
	saveCallChmod     = "chmod"
	saveCallClose     = "close"
	saveCallImport    = "import"
	saveCallSync      = "sync"
	saveCallValidate  = "validate"
	saveSuccessCase   = "success"
	saveTestModelPath = "/models/model.db"
	saveTestTemporary = "/models/temporary.db"
)

type exportTemporaryModelTest struct {
	configure func(*saveDependencies, *temporaryModelStoreStub, *Predictor)
	wantCalls []string
}

type saveCloserStub struct {
	closeCalls int
	closeErr   error
}

func (s *saveCloserStub) Close() error {
	s.closeCalls++

	return s.closeErr
}

type saveFileInfoStub struct {
	mode os.FileMode
}

func (s saveFileInfoStub) IsDir() bool        { return false }
func (s saveFileInfoStub) ModTime() time.Time { return time.Time{} }
func (s saveFileInfoStub) Mode() os.FileMode  { return s.mode }
func (s saveFileInfoStub) Name() string       { return "model.db" }
func (s saveFileInfoStub) Size() int64        { return 0 }
func (s saveFileInfoStub) Sys() any           { return nil }

type saveFileStub struct {
	name      string
	calls     []string
	chmodErr  error
	closeErr  error
	syncErr   error
	chmodMode os.FileMode
}

func (s *saveFileStub) Chmod(mode os.FileMode) error {
	s.calls = append(s.calls, saveCallChmod)
	s.chmodMode = mode

	return s.chmodErr
}

func (s *saveFileStub) Close() error {
	s.calls = append(s.calls, saveCallClose)

	return s.closeErr
}

func (s *saveFileStub) Name() string { return s.name }

func (s *saveFileStub) Sync() error {
	s.calls = append(s.calls, saveCallSync)

	return s.syncErr
}

type temporaryModelStoreStub struct {
	calls       []string
	closeErr    error
	importErr   error
	validateErr error
}

func (s *temporaryModelStoreStub) Close() error {
	s.calls = append(s.calls, saveCallClose)

	return s.closeErr
}

func (s *temporaryModelStoreStub) Import(
	context.Context,
	[]modelstore.Class,
	modelstore.ModelStore,
) error {
	s.calls = append(s.calls, saveCallImport)

	return s.importErr
}

func (s *temporaryModelStoreStub) Validate(context.Context) (sqlitestore.Metadata, error) {
	s.calls = append(s.calls, saveCallValidate)

	return sqlitestore.Metadata{}, s.validateErr
}

func TestResolveModelHasher(t *testing.T) {
	t.Parallel()

	dummyMetaData := new(sqlitestore.Metadata)
	dummyMetaData.CodecVersion = 999

	hasher, err := resolveModelHasher(*dummyMetaData, nil)
	require.Error(t, err,
		"undefined codec version should error")
	require.Nil(t, hasher,
		"returned hasher should be nil on error")
}

func TestExportTemporaryModel(t *testing.T) {
	t.Parallel()

	predictor := &Predictor{hasher: NewDefaultHasher(), store: &fakeStore{scope: 1}}

	tests := map[string]exportTemporaryModelTest{
		"create": {
			configure: func(deps *saveDependencies, _ *temporaryModelStoreStub, _ *Predictor) {
				deps.createStore = func(
					context.Context, string, sqlitestore.Metadata, sqlitestore.OpenConfig,
				) (temporaryModelStore, error) {
					return nil, errTestSaveFailed
				}
			},
		},
		"classes": {
			configure: func(_ *saveDependencies, _ *temporaryModelStoreStub, target *Predictor) {
				target.store = &fakeStore{scope: 1, classesErr: errTestSaveFailed}
			},
			wantCalls: []string{saveCallClose},
		},
		saveCallImport: {
			configure: func(_ *saveDependencies, store *temporaryModelStoreStub, _ *Predictor) {
				store.importErr = errTestSaveFailed
			},
			wantCalls: []string{saveCallImport, saveCallClose},
		},
		saveCallValidate: {
			configure: func(_ *saveDependencies, store *temporaryModelStoreStub, _ *Predictor) {
				store.validateErr = errTestSaveFailed
			},
			wantCalls: []string{saveCallImport, saveCallValidate, saveCallClose},
		},
		saveCallClose: {
			configure: func(_ *saveDependencies, store *temporaryModelStoreStub, _ *Predictor) {
				store.closeErr = errTestSaveFailed
			},
			wantCalls: []string{saveCallImport, saveCallValidate, saveCallClose},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertExportTemporaryModelFailure(t, predictor, test)
		})
	}
}

func TestPrepareSaveDestination(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	lock := new(saveCloserStub)
	deps := defaultSaveDependencies()
	deps.canonicalPath = func(string) (string, error) { return saveTestModelPath, nil }
	deps.acquirePathLock = func(context.Context, string) (io.Closer, error) { return lock, nil }
	deps.isOpenAlias = func(string) (bool, error) { return false, nil }
	deps.stat = func(string) (os.FileInfo, error) { return saveFileInfoStub{mode: 0o400}, nil }

	destination, err := prepareSaveDestination(ctx, "model.db", deps)
	require.NoError(t, err)
	require.Equal(t, saveTestModelPath, destination.path)
	require.Equal(t, filepath.FromSlash("/models"), destination.directory)
	require.Equal(t, os.FileMode(0o400), destination.mode)
	require.Same(t, lock, destination.lock)
	require.Zero(t, lock.closeCalls)
}

func TestPrepareSaveDestinationFailures(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*saveDependencies){
		"canonical path": func(deps *saveDependencies) {
			deps.canonicalPath = func(string) (string, error) { return "", errTestSaveFailed }
		},
		"path lock": func(deps *saveDependencies) {
			deps.acquirePathLock = func(context.Context, string) (io.Closer, error) {
				return nil, errTestSaveFailed
			}
		},
		"active lookup": func(deps *saveDependencies) {
			deps.isOpenAlias = func(string) (bool, error) { return false, errTestSaveFailed }
		},
		"active model": func(deps *saveDependencies) {
			deps.isOpenAlias = func(string) (bool, error) { return true, nil }
		},
		"destination stat": func(deps *saveDependencies) {
			deps.stat = func(string) (os.FileInfo, error) { return nil, errTestSaveFailed }
		},
	}

	for name, configure := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lock := new(saveCloserStub)
			deps := stubSaveDependencies(lock)
			configure(&deps)

			_, err := prepareSaveDestination(context.Background(), "model.db", deps)
			require.Error(t, err)
		})
	}
}

func TestPrepareTemporaryModelPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configure func(*saveDependencies, *saveFileStub)
		wantPath  string
	}{
		saveSuccessCase: {wantPath: filepath.FromSlash(saveTestTemporary)},
		"create": {
			configure: func(deps *saveDependencies, _ *saveFileStub) {
				deps.createTemp = func(string, string) (saveFile, error) { return nil, errTestSaveFailed }
			},
		},
		"close": {
			configure: func(_ *saveDependencies, file *saveFileStub) { file.closeErr = errTestSaveFailed },
		},
		"remove": {
			configure: func(deps *saveDependencies, _ *saveFileStub) {
				deps.remove = func(string) error { return errTestSaveFailed }
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			file := &saveFileStub{name: saveTestTemporary}
			deps := defaultSaveDependencies()
			deps.createTemp = func(string, string) (saveFile, error) { return file, nil }
			deps.remove = func(string) error { return nil }
			if test.configure != nil {
				test.configure(&deps, file)
			}

			path, err := prepareTemporaryModelPath("/models", deps)
			if test.wantPath == "" {
				require.ErrorIs(t, err, errTestSaveFailed)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.wantPath, path)
		})
	}
}

func TestSyncSavedModelFile(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configure func(*saveDependencies, *saveFileStub)
		wantCalls []string
	}{
		saveSuccessCase: {wantCalls: []string{saveCallChmod, saveCallSync, saveCallClose}},
		"open": {
			configure: func(deps *saveDependencies, _ *saveFileStub) {
				deps.openFile = func(string, int, os.FileMode) (saveFile, error) {
					return nil, errTestSaveFailed
				}
			},
		},
		"chmod": {
			configure: func(_ *saveDependencies, file *saveFileStub) { file.chmodErr = errTestSaveFailed },
			wantCalls: []string{saveCallChmod, saveCallClose},
		},
		"sync": {
			configure: func(_ *saveDependencies, file *saveFileStub) { file.syncErr = errTestSaveFailed },
			wantCalls: []string{saveCallChmod, saveCallSync, saveCallClose},
		},
		"close": {
			configure: func(_ *saveDependencies, file *saveFileStub) { file.closeErr = errTestSaveFailed },
			wantCalls: []string{saveCallChmod, saveCallSync, saveCallClose},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			file := new(saveFileStub)
			deps := defaultSaveDependencies()
			deps.openFile = func(string, int, os.FileMode) (saveFile, error) { return file, nil }
			if test.configure != nil {
				test.configure(&deps, file)
			}

			err := syncSavedModelFile("temporary.db", 0o400, deps)
			if name == saveSuccessCase {
				require.NoError(t, err)
				require.Equal(t, os.FileMode(0o400), file.chmodMode)
			} else {
				require.ErrorIs(t, err, errTestSaveFailed)
			}
			require.Equal(t, test.wantCalls, file.calls)
		})
	}
}

func TestReplaceSavedModel(t *testing.T) {
	t.Parallel()

	destination := saveDestination{directory: "/models", path: saveTestModelPath}
	deps := defaultSaveDependencies()
	deps.rename = func(string, string) error { return errTestSaveFailed }
	require.ErrorIs(t, replaceSavedModel("temporary.db", destination, deps), errTestSaveFailed)

	deps.rename = func(string, string) error { return nil }
	deps.syncDirectory = func(string) error { return errTestSaveFailed }
	err := replaceSavedModel("temporary.db", destination, deps)
	require.ErrorIs(t, err, ErrSaveDurabilityUnknown)
	require.ErrorIs(t, err, errTestSaveFailed)
}

func stubSaveDependencies(lock io.Closer) saveDependencies {
	deps := defaultSaveDependencies()
	deps.canonicalPath = func(string) (string, error) { return saveTestModelPath, nil }
	deps.acquirePathLock = func(context.Context, string) (io.Closer, error) { return lock, nil }
	deps.isOpenAlias = func(string) (bool, error) { return false, nil }
	deps.stat = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }

	return deps
}

func assertExportTemporaryModelFailure(
	t *testing.T,
	predictor *Predictor,
	test exportTemporaryModelTest,
) {
	t.Helper()

	store := new(temporaryModelStoreStub)
	deps := defaultSaveDependencies()
	deps.createStore = func(
		context.Context, string, sqlitestore.Metadata, sqlitestore.OpenConfig,
	) (temporaryModelStore, error) {
		return store, nil
	}
	target := *predictor
	test.configure(&deps, store, &target)

	err := exportTemporaryModel(context.Background(), &target, "temporary.db", deps)
	require.ErrorIs(t, err, errTestSaveFailed)
	require.Equal(t, test.wantCalls, store.calls)
}
