// Package modelstore defines storage records used by bayes.Predictor.
package modelstore

import (
	"context"
	"errors"
	"iter"
)

var (
	// ErrClosed means that a ModelStore operation was requested after Close.
	ErrClosed = errors.New("model store is closed")
	// ErrClassConflict means that one class ID has two different values.
	ErrClassConflict = errors.New("model store class ID conflicts with an existing value")
	// ErrCommitIndeterminate means that a durable commit result is unknown.
	ErrCommitIndeterminate = errors.New("model store commit result is indeterminate")
	// ErrCountOverflow means that applying a batch would exceed int64 counts.
	ErrCountOverflow = errors.New("model store count overflow")
	// ErrInvalidBatch means that a training batch violates the store contract.
	ErrInvalidBatch = errors.New("invalid model store training batch")
	// ErrPoisoned means that a store cannot continue after an uncertain commit.
	ErrPoisoned = errors.New("model store is poisoned")
)

// CandidateStats contains the counts needed to score one possible class.
type CandidateStats struct {
	ClassID   uint64
	ToCount   int64
	PairCount int64
}

// Class stores one reversible class value.
// Payload belongs to the value that contains it and must not be retained.
type Class struct {
	ID      uint64
	TypeTag byte
	Payload []byte
}

// ModelStore keeps exact transition counts and reversible class records.
// Implementations are used sequentially by Predictor.
type ModelStore interface {
	ScopeID() uint64
	Apply(ctx context.Context, batch TrainingBatch) error
	Stats(ctx context.Context, fromID uint64) (Stats, error)
	Classes(ctx context.Context) ([]Class, error)
	ExportTransitions(ctx context.Context, sink func(TransitionCount) error) error
	Reset(ctx context.Context) error
	Close() error
}

// Stats contains all counts needed to score candidates for one input ID.
type Stats struct {
	Total      int64
	FromCount  int64
	Candidates []CandidateStats
}

// TrainingBatch is one atomic model update.
// Transitions must return the same sequence every time it is called.
type TrainingBatch struct {
	Classes     []Class
	Transitions TransitionIterator
}

// TransitionCount contains one exact stored pair count.
type TransitionCount struct {
	FromID uint64
	ToID   uint64
	Count  int64
}

// TransitionDelta contains one positive pair-count update.
type TransitionDelta struct {
	FromID uint64
	ToID   uint64
	Count  int64
}

// TransitionIterator returns a fresh, deterministic transition sequence.
type TransitionIterator func() iter.Seq[TransitionDelta]
