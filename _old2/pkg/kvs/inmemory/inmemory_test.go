package inmemory

import (
	"encoding/gob"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
//  DB.Delete() method
// ----------------------------------------------------------------------------

func TestDB_Delete_non_existing_key(t *testing.T) {
	db := New()

	err := db.Delete(5963)

	require.NoError(t, err, "non existing key should not error")
}

// ----------------------------------------------------------------------------
//  DB.GetValue() method
// ----------------------------------------------------------------------------

func TestDB_Get_non_existing_key(t *testing.T) {
	db := New()

	v, err := db.GetValue(5963)

	require.Error(t, err, "non existing key should error")
	require.Contains(t, err.Error(), "key 5963 not found")
	require.Nil(t, v, "non existing key should return nil")
}

// ----------------------------------------------------------------------------
//  DB.Load() method
// ----------------------------------------------------------------------------

func TestDB_Load_non_existing_file(t *testing.T) {
	db := New()

	err := db.Load("non-existing-file.db")

	require.Error(t, err, "non existing file should error")
	require.Contains(t, err.Error(), "failed to read file")
}

func TestDB_Load_malformed_file(t *testing.T) {
	db := New()

	pathDirTmp := t.TempDir()
	pathFileTmp := filepath.Join(pathDirTmp, "malformed-file.db")

	malformedData := []byte("malformed data")

	perm := fs.FileMode(0o600)
	if err := os.WriteFile(pathFileTmp, malformedData, perm); err != nil {
		t.Fatal(err)
	}

	err := db.Load(pathFileTmp)

	require.Error(t, err, "malformed data should error")
	require.Contains(t, err.Error(), "can not decode gob data from:")
}

// ----------------------------------------------------------------------------
//  DB.Next() method
// ----------------------------------------------------------------------------

func TestDB_Next(t *testing.T) {
	db := New()

	require.NoError(t, db.Set(100, "foo"))
	require.NoError(t, db.Set(200, "bar"))
	require.NoError(t, db.Set(300, "baz"))

	expectKeys := []uint64{100, 200, 300}
	expectValues := []string{"foo", "bar", "baz"}

	actualKeys := []uint64{}
	actualValues := []string{}

	for db.HasNext() {
		key, anyValue, err := db.Next()
		require.NoError(t, err, "next should not error")

		value, ok := anyValue.(string)
		require.True(t, ok, "value should be a string")

		actualKeys = append(actualKeys, key)
		actualValues = append(actualValues, value)
	}

	require.Equal(t, expectKeys, actualKeys, "keys should be equal")
	require.Equal(t, expectValues, actualValues, "values should be equal")
}

func TestDB_Next_no_more_elements(t *testing.T) {
	db := New()

	key, v, err := db.Next()

	require.Error(t, err, "no more elements should error")
	require.Contains(t, err.Error(), "no more elements")
	require.Zero(t, key, "no more elements should return zero key")
	require.Nil(t, v, "no more elements should return nil value")
}

func TestDB_Next_key_not_found(t *testing.T) {
	db := New()
	err := db.Set(5963, "foo")

	require.NoError(t, err, "set should not error")

	db.Keys = []uint64{1234}

	key, v, err := db.Next()

	require.Error(t, err, "key not found should error")
	require.Contains(t, err.Error(), "key 1234 not found")
	require.Zero(t, key, "key not found should return zero key")
	require.Nil(t, v, "key not found should return nil value")
}

// ----------------------------------------------------------------------------
//  DB.Save() method
// ----------------------------------------------------------------------------

func TestDB_Save_invalid_file_path(t *testing.T) {
	db := New()

	err := db.Save(t.TempDir())

	require.Error(t, err, "non existing file should error")
	require.Contains(t, err.Error(), "failed to open file")
}

func TestDB_Save_fail_to_encode_gob(t *testing.T) {
	// Backup the gobNewEncoder and defer restore it
	OldGobNewEncoder := gobNewEncoder
	defer func() {
		gobNewEncoder = OldGobNewEncoder
	}()

	// Mock gob.Encoder to force fail encoding data to gob
	pathDirTmp := t.TempDir()

	dummyWriter, err := os.OpenFile(pathDirTmp, os.O_RDONLY, 0o644)

	require.NoError(t, err, "failed to open file")

	defer dummyWriter.Close()

	gobNewEncoder = func(w io.Writer) *gob.Encoder {
		// return with the bad file descriptor
		return gob.NewEncoder(dummyWriter)
	}

	// Test
	db := New()
	pathFile := filepath.Join(pathDirTmp, "fail-to-encode-gob.db")

	err = db.Save(pathFile)

	require.Error(t, err, "non existing file should error")
	require.Contains(t, err.Error(), "can not encode to gob")
}

// ----------------------------------------------------------------------------
//  DB.Set() method
// ----------------------------------------------------------------------------

func TestDB_Set_data_is_nil(t *testing.T) {
	db := new(DB)

	err := db.Set(5963, "foo")

	require.NoError(t, err, "data field is nil should not error")

	expect := "foo"
	actual, err := db.GetValue(5963)

	require.NoError(t, err, "existing key should not error")
	require.Equal(t, expect, actual, "data field is nil should not change value")
}
