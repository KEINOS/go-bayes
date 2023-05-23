package kvs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew_unknown_type(t *testing.T) {
	{
		db := New(UnknownDB)

		require.Nil(t, db, "unknown DBType should return nil")
	}
	{
		unknown := DBType(9999)
		db := New(unknown)

		require.Nil(t, db, "unregistered DBType value should return nil")
	}
}

func TestDB_Set_struct_then_save_and_load(t *testing.T) {
	pathDirTmp := t.TempDir()
	pathFileTmp := filepath.Join(pathDirTmp, "tmp.db")

	// Create DB1 and save
	db1 := New(InMemoryDB)
	require.NotNil(t, db1, "valid DBType should not be nil")

	type Dummy struct {
		A string
		B int
	}

	db1.RegistGob(*new(Dummy))

	foo1 := Dummy{A: "foo", B: 100}
	bar1 := Dummy{A: "bar", B: 200}

	require.NoError(t, db1.Set(1, foo1))
	require.NoError(t, db1.Set(2, bar1))

	require.NoError(t, db1.Save(pathFileTmp))

	// Create DB2 and load
	db2 := New(InMemoryDB)
	require.NotNil(t, db2, "valid DBType should not be nil")

	db2.RegistGob(*new(Dummy))

	require.NoError(t, db2.Load(pathFileTmp))

	// Check DB2
	anyValue, err := db2.GetValue(1)
	require.NoError(t, err)

	foo2, ok := anyValue.(Dummy) // type assertion
	require.True(t, ok)

	require.Equal(t, foo1.A, foo2.A)
}
