package bayes

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"iter"
	"maps"
	"reflect"
	"slices"

	"github.com/KEINOS/go-bayes/bayes/internal/theorem"
	"github.com/KEINOS/go-bayes/bayes/modelstore"
)

var (
	errPredictorHasherUninit    = errors.New("hasher is not initialized")
	errPredictorHasherNameEmpty = errors.New("hasher name must not be empty")
	errPredictorNotInitialized  = errors.New("predictor is not initialized")
	errPredictorItemsNil        = errors.New("items must not be nil")
	errPredictorItemsNotSlice   = errors.New("items must be a slice")
	errUnsupportedValueType     = errors.New("unsupported value type")
)

// PredictorConfig defines one Predictor and its immutable model store.
// A nil Hasher selects NewDefaultHasher.
type PredictorConfig struct {
	Storage           Storage
	ScopeID           uint64
	Hasher            Hasher
	ModelStore        ModelStore
	SQLitePath        string
	SQLiteSynchronous SQLiteSynchronous
	SQLiteCacheKiB    int
}

// Predictor learns transitions from folded ordered contexts to possible next
// values. Each instance owns its store and class cache.
//
// Predictor is not safe for concurrent use from multiple goroutines.
type Predictor struct {
	store      modelstore.ModelStore
	classes    map[uint64]classEntry
	storage    Storage
	scopeID    uint64
	hasher     Hasher
	scratch    codecScratch
	sqlitePath string
	closed     bool
}

type classEntry struct {
	Raw    any
	Stored modelstore.Class
}

type codecScratch struct {
	bytes []byte
	ids   []uint64
}

// NewPredictor creates a Predictor from config.
func NewPredictor(ctx context.Context, config PredictorConfig) (*Predictor, error) {
	hasher := config.Hasher
	if hasher == nil {
		hasher = NewDefaultHasher()
	}

	if hasher.Name() == "" {
		return nil, errPredictorHasherNameEmpty
	}

	config.Hasher = hasher

	store, err := newModelStore(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create model store: %w", err)
	}

	predictor := &Predictor{
		store:      store,
		classes:    make(map[uint64]classEntry),
		storage:    config.Storage,
		scopeID:    config.ScopeID,
		hasher:     hasher,
		scratch:    codecScratch{bytes: nil, ids: nil},
		sqlitePath: config.SQLitePath,
		closed:     false,
	}

	err = predictor.loadClassCache(ctx)
	if err != nil {
		if config.ModelStore == nil {
			_ = store.Close()
		}

		return nil, fmt.Errorf("failed to load model classes: %w", err)
	}

	return predictor, nil
}

// Close releases the model store. It is safe to call more than once.
func (p *Predictor) Close() error {
	if p.closed {
		return nil
	}

	p.closed = true
	if p.store == nil {
		return nil
	}

	err := p.store.Close()
	if err != nil {
		return fmt.Errorf("failed to close model store: %w", err)
	}

	return nil
}

// GetClass resolves classID to the original value recorded during training.
// It returns nil when classID is unknown.
func (p *Predictor) GetClass(classID uint64) any {
	return p.classes[classID].Raw
}

// HashTrans converts supported values to item IDs and folds their ordered
// sequence into one deterministic context ID.
func (p *Predictor) HashTrans(transitions ...any) (uint64, error) {
	if p.hasher == nil {
		return 0, errPredictorHasherUninit
	}

	p.scratch.ids = ensureUint64Capacity(p.scratch.ids, len(transitions))

	itemIDs := p.scratch.ids[:0]
	for _, transition := range transitions {
		itemID, err := p.itemID(transition)
		if err != nil {
			return 0, fmt.Errorf("failed to encode transition: %w", err)
		}

		itemIDs = append(itemIDs, itemID)
	}

	return p.contextID(itemIDs), nil
}

// ID returns the model scope ID.
func (p *Predictor) ID() uint64 {
	if p.store != nil {
		return p.store.ScopeID()
	}

	return p.scopeID
}

// MarshalJSON prevents accidental use of the removed JSON model format.
func (p *Predictor) MarshalJSON() ([]byte, error) {
	return nil, ErrJSONPersistenceUnsupported
}

// Predict scores learned candidates for one ordered context.
//
//nolint:cyclop // scoring keeps error, empty-model, and deterministic tie behavior explicit.
func (p *Predictor) Predict(ctx context.Context, items any) (uint64, error) {
	if p.closed {
		return 0, ErrPredictorClosed
	}

	if p.store == nil {
		return 0, errPredictorNotInitialized
	}

	itemsAny, err := normalizeItems(items)
	if err != nil {
		return 0, err
	}

	flowID, err := p.HashTrans(itemsAny...)
	if err != nil {
		return 0, fmt.Errorf("failed to hash the flow: %w", err)
	}

	stats, err := p.store.Stats(ctx, flowID)
	if err != nil {
		return 0, fmt.Errorf("failed to read model statistics: %w", err)
	}

	if stats.Total == 0 {
		return 0, nil
	}

	var (
		bestID    uint64
		bestScore float64
	)

	found := false

	for _, candidate := range stats.Candidates {
		priorTo := float64(candidate.ToCount) / float64(stats.Total)
		priorPair := float64(candidate.PairCount) / float64(stats.Total)
		priorNotPair := float64(stats.FromCount-candidate.PairCount) / float64(stats.Total)
		score := theorem.Bayes(priorTo, priorPair, priorNotPair)

		if score > 0 && (!found || score > bestScore || score == bestScore && candidate.ClassID < bestID) {
			bestID = candidate.ClassID
			bestScore = score
			found = true
		}
	}

	return bestID, nil
}

// Reset atomically clears learned transitions and classes in the same store.
func (p *Predictor) Reset(ctx context.Context) error {
	if p.closed {
		return ErrPredictorClosed
	}

	if p.store == nil {
		return errPredictorNotInitialized
	}

	err := p.store.Reset(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset predictor: %w", err)
	}

	p.classes = make(map[uint64]classEntry)

	return nil
}

// Save writes a complete portable SQLite model file.
func (p *Predictor) Save(ctx context.Context, path string) error {
	if p.closed {
		return ErrPredictorClosed
	}

	return saveModel(ctx, p, path)
}

// Train learns each next value in an observed sequence and every folded suffix
// of its preceding ordered context.
//
//nolint:cyclop,funlen,gocognit // training pre-encodes values before one atomic store call.
func (p *Predictor) Train(ctx context.Context, items any) error {
	if p.closed {
		return ErrPredictorClosed
	}

	if p.store == nil {
		return errPredictorNotInitialized
	}

	if p.hasher == nil {
		return errPredictorHasherUninit
	}

	itemsAny, err := normalizeItems(items)
	if err != nil {
		return err
	}

	err = ctx.Err()
	if err != nil {
		return fmt.Errorf("training canceled: %w", err)
	}

	itemIDs := make([]uint64, len(itemsAny))
	newClasses := make(map[uint64]classEntry)

	for index, raw := range itemsAny {
		itemID, itemErr := p.itemID(raw)
		if itemErr != nil {
			return fmt.Errorf("failed during training iteration: %w", itemErr)
		}

		itemIDs[index] = itemID

		if index > 0 {
			stored, classErr := p.classRecord(itemID, raw)
			if classErr != nil {
				return fmt.Errorf("failed to encode class: %w", classErr)
			}

			if existing, exists := newClasses[itemID]; exists && !sameStoredClass(existing.Stored, stored) {
				return fmt.Errorf("%w: values in one training sequence share class ID %d", ErrHashCollision, itemID)
			}

			if existing, exists := p.classes[itemID]; exists && !sameStoredClass(existing.Stored, stored) {
				return fmt.Errorf("%w: class ID %d is already assigned", ErrHashCollision, itemID)
			}

			newClasses[itemID] = classEntry{Raw: raw, Stored: stored}
		}
	}

	classes := make([]modelstore.Class, 0, len(newClasses))
	for _, class := range newClasses {
		classes = append(classes, class.Stored)
	}

	slices.SortFunc(classes, func(left, right modelstore.Class) int {
		return cmp.Compare(left.ID, right.ID)
	})

	batch := modelstore.TrainingBatch{
		Classes: classes,
		Transitions: func() iter.Seq[modelstore.TransitionDelta] {
			return func(yield func(modelstore.TransitionDelta) bool) {
				for nextIndex := 1; nextIndex < len(itemIDs); nextIndex++ {
					for suffixStart := range nextIndex {
						fromID := p.contextID(itemIDs[suffixStart:nextIndex])
						if !yield(modelstore.TransitionDelta{FromID: fromID, ToID: itemIDs[nextIndex], Count: 1}) {
							return
						}
					}
				}
			}
		},
	}

	err = p.store.Apply(ctx, batch)
	if err != nil {
		return fmt.Errorf("failed to apply training batch: %w", err)
	}

	maps.Copy(p.classes, newClasses)

	return nil
}

// UnmarshalJSON prevents accidental use of the removed JSON model format.
func (p *Predictor) UnmarshalJSON([]byte) error {
	return ErrJSONPersistenceUnsupported
}

func (p *Predictor) loadClassCache(ctx context.Context) error {
	classes, err := p.store.Classes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list stored classes: %w", err)
	}

	for _, stored := range classes {
		raw, decodeErr := p.decodeClass(stored)
		if decodeErr != nil {
			return decodeErr
		}

		p.classes[stored.ID] = classEntry{Raw: raw, Stored: cloneStoredClass(stored)}
	}

	return nil
}

func normalizeItems(items any) ([]any, error) {
	if items == nil {
		return nil, errPredictorItemsNil
	}

	if values, ok := items.([]any); ok {
		return values, nil
	}

	value := reflect.ValueOf(items)
	if value.Kind() != reflect.Slice {
		return nil, errPredictorItemsNotSlice
	}

	normalized := make([]any, 0, value.Len())
	for index := range value.Len() {
		normalized = append(normalized, value.Index(index).Interface())
	}

	return normalized, nil
}

func cloneStoredClass(class modelstore.Class) modelstore.Class {
	class.Payload = append([]byte(nil), class.Payload...)

	return class
}

func sameStoredClass(left, right modelstore.Class) bool {
	return left.TypeTag == right.TypeTag && bytes.Equal(left.Payload, right.Payload)
}
