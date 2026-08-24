// Package mapstore provides the in-memory ModelStore implementation.
package mapstore

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/KEINOS/go-bayes/bayes/modelstore"
)

var _ modelstore.ModelStore = (*Store)(nil)

type pair struct {
	from uint64
	to   uint64
}

// Store keeps a complete model in Go maps.
type Store struct {
	scopeID uint64
	total   int64
	classes map[uint64]modelstore.Class
	from    map[uint64]int64
	to      map[uint64]int64
	pairs   map[pair]int64
	closed  bool
}

// New returns an empty in-memory store.
func New(scopeID uint64) *Store {
	return &Store{
		scopeID: scopeID,
		total:   0,
		classes: make(map[uint64]modelstore.Class),
		from:    make(map[uint64]int64),
		to:      make(map[uint64]int64),
		pairs:   make(map[pair]int64),
		closed:  false,
	}
}

// Apply applies one batch atomically.
//
//nolint:cyclop,funlen,gocognit // atomic preflight validates every related count before mutation.
func (s *Store) Apply(ctx context.Context, batch modelstore.TrainingBatch) error {
	if s.closed {
		return modelstore.ErrClosed
	}

	if batch.Transitions == nil {
		return fmt.Errorf("%w: nil transition iterator", modelstore.ErrInvalidBatch)
	}

	classes := make(map[uint64]modelstore.Class, len(batch.Classes))
	for _, class := range batch.Classes {
		if existing, exists := classes[class.ID]; exists && !sameClass(existing, class) {
			return fmt.Errorf("%w: class %d conflicts within batch", modelstore.ErrClassConflict, class.ID)
		}

		if existing, exists := s.classes[class.ID]; exists && !sameClass(existing, class) {
			return fmt.Errorf("%w: class %d conflicts with stored value", modelstore.ErrClassConflict, class.ID)
		}

		classes[class.ID] = cloneClass(class)
	}

	deltas := make(map[pair]int64)
	fromDeltas := make(map[uint64]int64)
	toDeltas := make(map[uint64]int64)

	var (
		totalDelta    int64
		validationErr error
	)

	for delta := range batch.Transitions() {
		err := ctx.Err()
		if err != nil {
			return fmt.Errorf("map-store batch canceled: %w", err)
		}

		if delta.Count <= 0 {
			return fmt.Errorf("%w: transition count must be positive", modelstore.ErrInvalidBatch)
		}

		key := pair{from: delta.FromID, to: delta.ToID}

		validationErr = addChecked(deltas, key, delta.Count)
		if validationErr != nil {
			break
		}

		validationErr = addChecked(fromDeltas, delta.FromID, delta.Count)
		if validationErr != nil {
			break
		}

		validationErr = addChecked(toDeltas, delta.ToID, delta.Count)
		if validationErr != nil {
			break
		}

		if totalDelta > math.MaxInt64-delta.Count {
			validationErr = modelstore.ErrCountOverflow

			break
		}

		totalDelta += delta.Count
	}

	if validationErr != nil {
		return validationErr
	}

	if s.total > math.MaxInt64-totalDelta {
		return modelstore.ErrCountOverflow
	}

	for id, delta := range fromDeltas {
		if s.from[id] > math.MaxInt64-delta {
			return modelstore.ErrCountOverflow
		}
	}

	for classID, delta := range toDeltas {
		if s.to[classID] > math.MaxInt64-delta {
			return modelstore.ErrCountOverflow
		}

		if _, exists := classes[classID]; !exists {
			if _, exists = s.classes[classID]; !exists {
				return fmt.Errorf("%w: transition class %d is missing", modelstore.ErrInvalidBatch, classID)
			}
		}
	}

	for key, delta := range deltas {
		if s.pairs[key] > math.MaxInt64-delta {
			return modelstore.ErrCountOverflow
		}
	}

	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("map-store commit canceled: %w", err)
	}

	s.total += totalDelta
	maps.Copy(s.classes, classes)

	for id, delta := range fromDeltas {
		s.from[id] += delta
	}

	for id, delta := range toDeltas {
		s.to[id] += delta
	}

	for key, delta := range deltas {
		s.pairs[key] += delta
	}

	return nil
}

// Classes returns copied class records in ascending ID order.
func (s *Store) Classes(ctx context.Context) ([]modelstore.Class, error) {
	if s.closed {
		return nil, modelstore.ErrClosed
	}

	err := ctx.Err()
	if err != nil {
		return nil, fmt.Errorf("class listing canceled: %w", err)
	}

	classes := make([]modelstore.Class, 0, len(s.classes))
	for _, class := range s.classes {
		classes = append(classes, cloneClass(class))
	}

	slices.SortFunc(classes, func(left, right modelstore.Class) int {
		return compareID(left.ID, right.ID)
	})

	return classes, nil
}

// Close releases this store. It is safe to call more than once.
func (s *Store) Close() error {
	s.closed = true

	return nil
}

// ExportTransitions sends copied transition counts in ascending ID order.
func (s *Store) ExportTransitions(ctx context.Context, sink func(modelstore.TransitionCount) error) error {
	if s.closed {
		return modelstore.ErrClosed
	}

	if sink == nil {
		return fmt.Errorf("%w: nil export sink", modelstore.ErrInvalidBatch)
	}

	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("transition export canceled: %w", err)
	}

	records := make([]modelstore.TransitionCount, 0, len(s.pairs))
	for key, count := range s.pairs {
		records = append(records, modelstore.TransitionCount{FromID: key.from, ToID: key.to, Count: count})
	}

	slices.SortFunc(records, func(left, right modelstore.TransitionCount) int {
		if result := compareID(left.FromID, right.FromID); result != 0 {
			return result
		}

		return compareID(left.ToID, right.ToID)
	})

	for _, record := range records {
		err := ctx.Err()
		if err != nil {
			return fmt.Errorf("transition export canceled: %w", err)
		}

		err = sink(record)
		if err != nil {
			return err
		}
	}

	return nil
}

// Reset clears all learned state.
func (s *Store) Reset(ctx context.Context) error {
	if s.closed {
		return modelstore.ErrClosed
	}

	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("map-store reset canceled: %w", err)
	}

	s.total = 0
	s.classes = make(map[uint64]modelstore.Class)
	s.from = make(map[uint64]int64)
	s.to = make(map[uint64]int64)
	s.pairs = make(map[pair]int64)

	return nil
}

// ScopeID returns the model scope ID.
func (s *Store) ScopeID() uint64 {
	return s.scopeID
}

// Stats returns one consistent count view for an input ID.
func (s *Store) Stats(ctx context.Context, fromID uint64) (modelstore.Stats, error) {
	if s.closed {
		return modelstore.Stats{}, modelstore.ErrClosed
	}

	err := ctx.Err()
	if err != nil {
		return modelstore.Stats{}, fmt.Errorf("statistics query canceled: %w", err)
	}

	stats := modelstore.Stats{
		Total:      s.total,
		FromCount:  s.from[fromID],
		Candidates: make([]modelstore.CandidateStats, 0, len(s.classes)),
	}
	for classID := range s.classes {
		stats.Candidates = append(stats.Candidates, modelstore.CandidateStats{
			ClassID:   classID,
			ToCount:   s.to[classID],
			PairCount: s.pairs[pair{from: fromID, to: classID}],
		})
	}

	slices.SortFunc(stats.Candidates, func(left, right modelstore.CandidateStats) int {
		return compareID(left.ClassID, right.ClassID)
	})

	return stats, nil
}

func addChecked[Key comparable](target map[Key]int64, key Key, delta int64) error {
	if target[key] > math.MaxInt64-delta {
		return modelstore.ErrCountOverflow
	}

	target[key] += delta

	return nil
}

func cloneClass(class modelstore.Class) modelstore.Class {
	class.Payload = append([]byte(nil), class.Payload...)

	return class
}

func sameClass(left, right modelstore.Class) bool {
	return left.TypeTag == right.TypeTag && bytes.Equal(left.Payload, right.Payload)
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
