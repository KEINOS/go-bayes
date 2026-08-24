//go:build cgo

package sqlitestore

import (
	"context"
	"fmt"
	"slices"
)

const transitionTableName = "from_a_to_b"

// Validate verifies the SQLite container, schema, metadata, and count invariants.
//
//nolint:cyclop,funlen,sqlclosecheck,wrapcheck // validation reports the exact failed SQLite check.
func (s *Store) Validate(ctx context.Context) (Metadata, error) {
	err := s.checkUsable()
	if err != nil {
		return Metadata{}, err
	}

	var actualApplicationID, actualVersion int
	err = s.conn.QueryRowContext(ctx, "PRAGMA application_id").Scan(&actualApplicationID)
	if err != nil {
		return Metadata{}, err
	}

	if actualApplicationID != applicationID {
		return Metadata{}, fmt.Errorf("%w: application ID %d", ErrInvalidModel, actualApplicationID)
	}

	err = s.conn.QueryRowContext(ctx, "PRAGMA user_version").Scan(&actualVersion)
	if err != nil {
		return Metadata{}, err
	}

	if actualVersion != schemaVersion {
		return Metadata{}, fmt.Errorf("%w: schema version %d", ErrInvalidModel, actualVersion)
	}

	rows, err := s.conn.QueryContext(ctx, `
SELECT name FROM sqlite_schema
WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return Metadata{}, err
	}

	var tables []string

	for rows.Next() {
		var name string

		err = rows.Scan(&name)
		if err != nil {
			_ = rows.Close()

			return Metadata{}, err
		}

		tables = append(tables, name)
	}

	err = rows.Err()
	if err != nil {
		_ = rows.Close()

		return Metadata{}, err
	}

	err = rows.Close()
	if err != nil {
		return Metadata{}, err
	}

	slices.Sort(tables)

	wantTables := requiredTableNames()
	if !slices.Equal(tables, wantTables) {
		return Metadata{}, fmt.Errorf("%w: unexpected schema tables %v", ErrInvalidModel, tables)
	}

	err = validateNoExtraObjects(ctx, s)
	if err != nil {
		return Metadata{}, err
	}

	err = validateStrictTables(ctx, s)
	if err != nil {
		return Metadata{}, err
	}

	var integrity string

	err = s.conn.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&integrity)
	if err != nil {
		return Metadata{}, err
	}

	if integrity != "ok" {
		return Metadata{}, fmt.Errorf("%w: quick_check: %s", ErrInvalidModel, integrity)
	}

	foreignRows, err := s.conn.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return Metadata{}, err
	}

	foreignViolation := foreignRows.Next()

	err = foreignRows.Err()
	if err != nil {
		_ = foreignRows.Close()

		return Metadata{}, err
	}

	err = foreignRows.Close()
	if err != nil {
		return Metadata{}, err
	}

	if foreignViolation {
		return Metadata{}, fmt.Errorf("%w: foreign key violation", ErrInvalidModel)
	}

	var (
		metadata                         Metadata
		itemProbe, contextProbe, scopeID int64
		rowCount                         int
	)

	err = s.conn.QueryRowContext(ctx, `
SELECT COUNT(*), MIN(codec_version), MIN(hasher_name), MIN(item_probe), MIN(context_probe), MIN(scope_id)
FROM metadata`).Scan(
		&rowCount,
		&metadata.CodecVersion,
		&metadata.HasherName,
		&itemProbe,
		&contextProbe,
		&scopeID,
	)
	if err != nil {
		return Metadata{}, err
	}

	if rowCount != 1 {
		return Metadata{}, fmt.Errorf("%w: metadata row count %d", ErrInvalidModel, rowCount)
	}

	metadata.ItemProbe = idFromSQL(itemProbe)
	metadata.ContextProbe = idFromSQL(contextProbe)
	metadata.ScopeID = idFromSQL(scopeID)

	valid, err := transactionInvariants(ctx, s.conn)
	if err != nil {
		return Metadata{}, err
	}

	if !valid {
		return Metadata{}, fmt.Errorf("%w: inconsistent model counts", ErrInvalidModel)
	}

	return metadata, nil
}

//nolint:wrapcheck // validation reports the exact failed SQLite check.
func validateNoExtraObjects(ctx context.Context, store *Store) error {
	rows, err := store.conn.QueryContext(ctx, `
SELECT type, name FROM sqlite_schema
WHERE name NOT LIKE 'sqlite_%' AND type <> 'table'
ORDER BY type, name`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	var objects []string

	for rows.Next() {
		var objectType, name string

		err = rows.Scan(&objectType, &name)
		if err != nil {
			return err
		}

		objects = append(objects, objectType+":"+name)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	if len(objects) != 0 {
		return fmt.Errorf("%w: unexpected schema objects %v", ErrInvalidModel, objects)
	}

	return nil
}

//nolint:cyclop,wrapcheck // validation reports the exact failed SQLite check.
func validateStrictTables(ctx context.Context, store *Store) error {
	rows, err := store.conn.QueryContext(ctx, "PRAGMA table_list")
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	requiredTables := requiredTableNames()
	found := make(map[string]bool, len(requiredTables))

	for rows.Next() {
		var (
			schema, name, tableType       string
			columns, withoutRowID, strict int
		)

		err = rows.Scan(&schema, &name, &tableType, &columns, &withoutRowID, &strict)
		if err != nil {
			return err
		}

		if schema != "main" || !slices.Contains(requiredTables, name) {
			continue
		}

		wantWithoutRowID := 0
		if name == transitionTableName {
			wantWithoutRowID = 1
		}

		if tableType != "table" || strict != 1 || withoutRowID != wantWithoutRowID {
			return fmt.Errorf("%w: incompatible table %s", ErrInvalidModel, name)
		}

		found[name] = true
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	if len(found) != len(requiredTables) {
		return fmt.Errorf("%w: missing STRICT table", ErrInvalidModel)
	}

	return nil
}

func requiredTableNames() []string {
	return []string{"classes", "from_a", transitionTableName, "metadata", "to_b"}
}
