//go:build cgo

package sqlitestore

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"math"
	"slices"

	"github.com/KEINOS/go-bayes/bayes/modelstore"
)

const upsertCountSQL = `
INSERT INTO %s (id, count) VALUES (?, ?)
ON CONFLICT(id) DO UPDATE SET count = count + excluded.count
WHERE count <= 9223372036854775807 - excluded.count`

// Apply writes a complete training call as one transaction.
func (s *Store) Apply(ctx context.Context, batch modelstore.TrainingBatch) error {
	if batch.Transitions == nil {
		return fmt.Errorf("%w: nil transition iterator", modelstore.ErrInvalidBatch)
	}

	return s.applyStream(ctx, batch.Classes, func(consume func(modelstore.TransitionDelta) error) error {
		for delta := range batch.Transitions() {
			err := consume(delta)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Import copies a consistent source-store snapshot without materializing all transitions.
func (s *Store) Import(ctx context.Context, classes []modelstore.Class, source modelstore.ModelStore) error {
	if source == nil {
		return fmt.Errorf("%w: nil import source", modelstore.ErrInvalidBatch)
	}

	return s.applyStream(ctx, classes, func(consume func(modelstore.TransitionDelta) error) error {
		return source.ExportTransitions(ctx, func(record modelstore.TransitionCount) error {
			return consume(modelstore.TransitionDelta(record))
		})
	})
}

//nolint:cyclop,funlen,gocognit,varnamelen,wrapcheck // this is the transaction boundary for driver operations.
func (s *Store) applyStream(
	ctx context.Context,
	classes []modelstore.Class,
	stream func(func(modelstore.TransitionDelta) error) error,
) error {
	err := s.checkUsable()
	if err != nil {
		return err
	}

	tx, err := s.conn.BeginTx(context.WithoutCancel(ctx), nil)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	classStmt, err := tx.PrepareContext(ctx, `
INSERT INTO classes (id, type_tag, payload) VALUES (?, ?, ?)
ON CONFLICT(id) DO UPDATE SET type_tag = excluded.type_tag, payload = excluded.payload
WHERE classes.type_tag = excluded.type_tag AND classes.payload = excluded.payload`)
	if err != nil {
		return err
	}
	defer func() { _ = classStmt.Close() }()

	for _, class := range sortedClasses(classes) {
		result, execErr := classStmt.ExecContext(ctx, idToSQL(class.ID), int64(class.TypeTag), class.Payload)
		err = execErr
		if err != nil {
			return err
		}

		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return rowsErr
		}

		if rowsAffected != 1 {
			return fmt.Errorf("%w: class %d conflicts with stored value", modelstore.ErrClassConflict, class.ID)
		}
	}

	fromStmt, err := tx.PrepareContext(ctx, fmt.Sprintf(upsertCountSQL, "from_a"))
	if err != nil {
		return err
	}
	defer func() { _ = fromStmt.Close() }()

	toStmt, err := tx.PrepareContext(ctx, fmt.Sprintf(upsertCountSQL, "to_b"))
	if err != nil {
		return err
	}

	defer func() { _ = toStmt.Close() }()

	pairStmt, err := tx.PrepareContext(ctx, `
INSERT INTO from_a_to_b (from_id, to_id, count) VALUES (?, ?, ?)
ON CONFLICT(from_id, to_id) DO UPDATE SET count = count + excluded.count
WHERE count <= 9223372036854775807 - excluded.count`)
	if err != nil {
		return err
	}

	defer func() { _ = pairStmt.Close() }()

	var totalDelta int64

	err = stream(func(delta modelstore.TransitionDelta) error {
		err = ctx.Err()
		if err != nil {
			return err
		}

		if delta.Count <= 0 {
			return fmt.Errorf("%w: transition count must be positive", modelstore.ErrInvalidBatch)
		}

		if totalDelta > math.MaxInt64-delta.Count {
			return modelstore.ErrCountOverflow
		}

		totalDelta += delta.Count

		result, execErr := fromStmt.ExecContext(ctx, idToSQL(delta.FromID), delta.Count)
		if execErr != nil {
			return execErr
		}

		err = overflowResult(result)
		if err != nil {
			return err
		}

		result, execErr = toStmt.ExecContext(ctx, idToSQL(delta.ToID), delta.Count)
		if execErr != nil {
			return execErr
		}

		err = overflowResult(result)
		if err != nil {
			return err
		}

		result, execErr = pairStmt.ExecContext(ctx, idToSQL(delta.FromID), idToSQL(delta.ToID), delta.Count)
		if execErr != nil {
			return execErr
		}

		err = overflowResult(result)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
UPDATE metadata SET total_count = total_count + ?
WHERE singleton = 1 AND total_count <= 9223372036854775807 - ?`, totalDelta, totalDelta)
	if err != nil {
		return err
	}

	err = overflowResult(result)
	if err != nil {
		return err
	}

	valid, err := transactionInvariants(ctx, tx)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("%w: batch breaks count or class invariants", modelstore.ErrInvalidBatch)
	}

	err = s.commit(tx)
	if err != nil {
		s.poisoned = true

		return fmt.Errorf("%w: %w", modelstore.ErrCommitIndeterminate, err)
	}

	committed = true

	return nil
}

// Classes returns copied class records in unsigned ID order.
//
//nolint:varnamelen,wrapcheck // driver errors retain their original SQLite details.
func (s *Store) Classes(ctx context.Context) ([]modelstore.Class, error) {
	err := s.checkUsable()
	if err != nil {
		return nil, err
	}

	rows, err := s.conn.QueryContext(ctx, `SELECT id, type_tag, payload FROM classes`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	classes := []modelstore.Class{}

	for rows.Next() {
		var (
			id      int64
			tag     int64
			payload []byte
		)

		err = rows.Scan(&id, &tag, &payload)
		if err != nil {
			return nil, err
		}

		classes = append(classes, modelstore.Class{
			ID:      idFromSQL(id),
			TypeTag: byte(tag), // #nosec G115 -- schema validation constrains the tag to one byte.
			Payload: append([]byte(nil), payload...),
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	slices.SortFunc(classes, func(left, right modelstore.Class) int {
		return cmp.Compare(left.ID, right.ID)
	})

	return classes, nil
}

// ExportTransitions streams exact pair counts in unsigned ID order.
//
//nolint:cyclop,wrapcheck // query and callback failures are returned at their exact boundary.
func (s *Store) ExportTransitions(ctx context.Context, sink func(modelstore.TransitionCount) error) error {
	err := s.checkUsable()
	if err != nil {
		return err
	}

	if sink == nil {
		return fmt.Errorf("%w: nil export sink", modelstore.ErrInvalidBatch)
	}

	rows, err := s.conn.QueryContext(ctx, `SELECT from_id, to_id, count FROM from_a_to_b`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	records := []modelstore.TransitionCount{}

	for rows.Next() {
		var fromID, toID, count int64

		err = rows.Scan(&fromID, &toID, &count)
		if err != nil {
			return err
		}

		records = append(records, modelstore.TransitionCount{FromID: idFromSQL(fromID), ToID: idFromSQL(toID), Count: count})
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	slices.SortFunc(records, func(left, right modelstore.TransitionCount) int {
		if result := cmp.Compare(left.FromID, right.FromID); result != 0 {
			return result
		}

		return cmp.Compare(left.ToID, right.ToID)
	})

	for _, record := range records {
		err = ctx.Err()
		if err != nil {
			return err
		}

		err = sink(record)
		if err != nil {
			return err
		}
	}

	return nil
}

// Reset atomically clears model rows while preserving metadata configuration.
//
//nolint:varnamelen,wrapcheck // driver errors retain their original SQLite details.
func (s *Store) Reset(ctx context.Context) error {
	err := s.checkUsable()
	if err != nil {
		return err
	}

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	for _, statement := range []string{
		"DELETE FROM from_a_to_b",
		"DELETE FROM from_a",
		"DELETE FROM to_b",
		"DELETE FROM classes",
		"UPDATE metadata SET total_count = 0 WHERE singleton = 1",
	} {
		_, err = tx.ExecContext(ctx, statement)
		if err != nil {
			return err
		}
	}

	err = s.commit(tx)
	if err != nil {
		s.poisoned = true

		return fmt.Errorf("%w: %w", modelstore.ErrCommitIndeterminate, err)
	}

	committed = true

	return nil
}

// Stats reads all candidate counts for one input in one query.
//
//nolint:wrapcheck // driver errors retain their original SQLite details.
func (s *Store) Stats(ctx context.Context, fromID uint64) (modelstore.Stats, error) {
	err := s.checkUsable()
	if err != nil {
		return modelstore.Stats{}, err
	}

	rows, err := s.conn.QueryContext(ctx, `
SELECT m.total_count, COALESCE(f.count, 0), c.id, t.count, COALESCE(p.count, 0)
FROM metadata AS m
LEFT JOIN classes AS c ON 1 = 1
LEFT JOIN to_b AS t ON t.id = c.id
LEFT JOIN from_a AS f ON f.id = ?
LEFT JOIN from_a_to_b AS p ON p.from_id = f.id AND p.to_id = c.id
WHERE m.singleton = 1`, idToSQL(fromID))
	if err != nil {
		return modelstore.Stats{}, err
	}
	defer func() { _ = rows.Close() }()

	stats := modelstore.Stats{Total: 0, FromCount: 0, Candidates: nil}

	for rows.Next() {
		var (
			classID   sql.NullInt64
			toCount   sql.NullInt64
			pairCount int64
		)

		err = rows.Scan(&stats.Total, &stats.FromCount, &classID, &toCount, &pairCount)
		if err != nil {
			return modelstore.Stats{}, err
		}

		if classID.Valid {
			stats.Candidates = append(stats.Candidates, modelstore.CandidateStats{
				ClassID:   idFromSQL(classID.Int64),
				ToCount:   toCount.Int64,
				PairCount: pairCount,
			})
		}
	}

	err = rows.Err()
	if err != nil {
		return modelstore.Stats{}, err
	}

	slices.SortFunc(stats.Candidates, func(left, right modelstore.CandidateStats) int {
		return cmp.Compare(left.ClassID, right.ClassID)
	})

	return stats, nil
}

//nolint:wrapcheck // the query adapter returns the underlying database error.
func transactionInvariants(ctx context.Context, queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}) (bool, error) {
	var valid int

	err := queryer.QueryRowContext(ctx, `
SELECT
    (SELECT total_count FROM metadata WHERE singleton = 1) = COALESCE((SELECT SUM(count) FROM from_a), 0)
AND (SELECT total_count FROM metadata WHERE singleton = 1) = COALESCE((SELECT SUM(count) FROM to_b), 0)
AND (SELECT total_count FROM metadata WHERE singleton = 1) = COALESCE((SELECT SUM(count) FROM from_a_to_b), 0)
AND NOT EXISTS (
    SELECT 1 FROM from_a AS f
    WHERE f.count != (SELECT COALESCE(SUM(p.count), 0) FROM from_a_to_b AS p WHERE p.from_id = f.id)
)
AND NOT EXISTS (
    SELECT 1 FROM to_b AS t
    WHERE t.count != (SELECT COALESCE(SUM(p.count), 0) FROM from_a_to_b AS p WHERE p.to_id = t.id)
)
AND NOT EXISTS (SELECT id FROM classes EXCEPT SELECT id FROM to_b)
AND NOT EXISTS (SELECT id FROM to_b EXCEPT SELECT id FROM classes)`).Scan(&valid)

	return valid == 1, err
}
