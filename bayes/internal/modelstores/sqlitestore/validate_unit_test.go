//go:build cgo

//nolint:exhaustruct // tests set only fields relevant to each case.
package sqlitestore

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate_returnsMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadata := testMetadata()
	store, err := Create(ctx, filepath.Join(t.TempDir(), "valid.db"), metadata, OpenConfig{})
	require.NoError(t, err)
	defer func() { require.NoError(t, store.Close()) }()

	validated, err := store.Validate(ctx)
	require.NoError(t, err)
	require.Equal(t, metadata, validated)
}

func TestValidate_reportsConnectionLoss(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Create(ctx, filepath.Join(t.TempDir(), "closed.db"), testMetadata(), OpenConfig{})
	require.NoError(t, err)
	require.NoError(t, store.conn.Close())

	_, err = store.Validate(ctx)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, store.Close())
}

func TestValidateRejectsCorruptSchemaAndMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := map[string][]string{
		"application id": {"PRAGMA application_id = 1"},
		"schema version": {"PRAGMA user_version = 99"},
		"extra table":    {"CREATE TABLE extra (id INTEGER) STRICT"},
		"extra index":    {"CREATE INDEX extra_index ON metadata(total_count)"},
		"extra trigger":  {"CREATE TRIGGER extra_trigger AFTER UPDATE ON metadata BEGIN SELECT 1; END"},
		"extra view":     {"CREATE VIEW extra_view AS SELECT * FROM metadata"},
		"counts":         {"UPDATE metadata SET total_count = 1"},
		"foreign key": {
			"PRAGMA foreign_keys = OFF",
			"INSERT INTO to_b (id, count) VALUES (99, 1)",
		},
		"non-STRICT transition table": {
			"DROP TABLE from_a_to_b",
			"CREATE TABLE from_a_to_b (from_id INTEGER, to_id INTEGER, count INTEGER, PRIMARY KEY (from_id, to_id))",
		},
		"multiple metadata rows": {
			"DROP TABLE metadata",
			`CREATE TABLE metadata (
				singleton INTEGER,
				codec_version INTEGER,
				hasher_name TEXT,
				item_probe INTEGER,
				context_probe INTEGER,
				scope_id INTEGER,
				total_count INTEGER
			) STRICT`,
			"INSERT INTO metadata VALUES (1, 1, 'test', 1, 2, 3, 0)",
			"INSERT INTO metadata VALUES (2, 1, 'test', 1, 2, 3, 0)",
		},
	}

	for name, mutations := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "invalid.db")
			store, err := Create(ctx, path, testMetadata(), OpenConfig{Portable: true})
			require.NoError(t, err)
			require.NoError(t, store.Close())

			database, err := sql.Open("sqlite3", path)
			require.NoError(t, err)
			for _, mutation := range mutations {
				_, err = database.ExecContext(ctx, mutation)
				require.NoError(t, err)
			}
			require.NoError(t, database.Close())

			_, err = Open(ctx, path, OpenConfig{Portable: true})
			require.ErrorIs(t, err, ErrInvalidModel)
		})
	}
}
