package bayes

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/KEINOS/go-bayes/bayes/internal/nodeloggers/logmem"
	"github.com/KEINOS/go-bayes/bayes/nodelogger"
	"github.com/zeebo/blake3"
)

// classEntry holds the class ID and the original value.
// This type is used internally by Predictor to map class IDs to their
// original raw values (used for GetClass and JSON serialization).
type classEntry struct {
	Raw any    `json:"raw"`
	ID  uint64 `json:"id"`
}

var (
	errPredictorHasherNil         = errors.New("hasher must not be nil")
	errPredictorHasherUninit      = errors.New("hasher is not initialized")
	errPredictorNotInitialized    = errors.New("predictor is not initialized")
	errPredictorItemsNil          = errors.New("items must not be nil")
	errPredictorItemsNotSlice     = errors.New("items must be a slice")
	errUnsupportedPredictorImpl   = errors.New("unsupported predictor implementation")
	errConvAnyToUint64Unsupported = errors.New("unsupported type for conversion")
)

// PredictorConfig defines dependencies and initial settings for Predictor.
type PredictorConfig struct {
	Storage Storage
	ScopeID uint64
	Hasher  Hasher
}

// Predictor provides an instance-based API without package-level singleton
// coupling.
//
// Predictor is not safe for concurrent use from multiple goroutines.
// Use external synchronization if needed.
type Predictor struct {
	predictor nodelogger.NodeLogger
	classes   map[uint64]classEntry
	storage   Storage
	scopeID   uint64
	hasher    Hasher
}

// NewPredictor creates a new Predictor instance from config.
// A nil config Hasher selects xxHash3.
func NewPredictor(config PredictorConfig) (*Predictor, error) {
	hasher := config.Hasher
	if hasher == nil {
		hasher = NewDefaultHasher()
	}

	predictor, err := newNodeLogger(config.Storage, config.ScopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create predictor: %w", err)
	}

	return &Predictor{
		predictor: predictor,
		classes:   make(map[uint64]classEntry),
		storage:   config.Storage,
		scopeID:   config.ScopeID,
		hasher:    hasher,
	}, nil
}

// ID returns the current scope ID.
func (p *Predictor) ID() uint64 {
	if p.predictor != nil {
		return p.predictor.ID()
	}

	return p.scopeID
}

// SetStorage sets storage for next Reset.
func (p *Predictor) SetStorage(storage Storage) {
	p.storage = storage
}

// SetHasher sets hasher used by HashTrans.
func (p *Predictor) SetHasher(hasher Hasher) error {
	if hasher == nil {
		return errPredictorHasherNil
	}

	p.hasher = hasher

	return nil
}

// Reset recreates predictor state with current storage and scope settings.
func (p *Predictor) Reset() error {
	predictor, err := newNodeLogger(p.storage, p.scopeID)
	if err != nil {
		return fmt.Errorf("failed to reset predictor: %w", err)
	}

	p.predictor = predictor
	p.classes = make(map[uint64]classEntry)

	return nil
}

// GetClass returns original value for classID.
func (p *Predictor) GetClass(classID uint64) any {
	return p.classes[classID].Raw
}

// HashTrans returns a flow ID for transitions.
func (p *Predictor) HashTrans(transitions ...any) (uint64, error) {
	if p.hasher == nil {
		return 0, errPredictorHasherUninit
	}

	hashed := make([]uint64, 0, len(transitions))

	for _, transition := range transitions {
		value, err := convAnyToUint64(transition)
		if err != nil {
			return 0, fmt.Errorf("failed to convert transitions to uint64: %w", err)
		}

		hashed = append(hashed, value)
	}

	flowID, err := p.hasher.HashTrans(hashed...)
	if err != nil {
		return 0, fmt.Errorf("failed to hash transitions: %w", err)
	}

	return flowID, nil
}

// Predict infers next class ID from items.
func (p *Predictor) Predict(items any) (uint64, error) {
	if p.predictor == nil {
		return 0, errPredictorNotInitialized
	}

	itemsAny, err := normalizeItems(items)
	if err != nil {
		return 0, err
	}

	biggest := struct {
		Probability float64
		Class       uint64
	}{
		Probability: 0,
		Class:       0,
	}

	flowID, err := p.HashTrans(itemsAny...)
	if err != nil {
		return 0, fmt.Errorf("failed to hash the flow: %w", err)
	}

	for classID := range p.classes {
		probability := p.predictor.Predict(flowID, classID)

		if biggest.Probability < probability {
			biggest.Probability = probability
			biggest.Class = classID
		}
	}

	return biggest.Class, nil
}

// Train updates predictor with observed item sequence.
func (p *Predictor) Train(items any) error {
	if p.predictor == nil {
		err := p.Reset()
		if err != nil {
			return fmt.Errorf("failed to initialize predictor: %w", err)
		}
	}

	if p.hasher == nil {
		return errPredictorHasherUninit
	}

	if p.classes == nil {
		p.classes = make(map[uint64]classEntry)
	}

	itemsAny, err := normalizeItems(items)
	if err != nil {
		return err
	}

	prevItem := uint64(0)
	drill := []uint64{}

	for index, itemRaw := range itemsAny {
		item, err := convAnyToUint64(itemRaw)
		if err != nil {
			return fmt.Errorf("failed during training iteration: %w", err)
		}

		if index == 0 {
			prevItem = item
			drill = append(drill, item)

			continue
		}

		p.predictor.Update(prevItem, item)

		for i := range drill {
			flowID, _ := p.hasher.HashTrans(drill[i:]...)
			p.predictor.Update(flowID, item)
		}

		prevItem = item
		drill = append(drill, item)
		p.addClass(item, itemRaw)
	}

	return nil
}

// normalizeItems converts any slice input into []any for Predict and Train.
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

type predictorJSON struct {
	Storage Storage               `json:"storage"`
	ScopeID uint64                `json:"scopeId"`
	Classes map[uint64]classEntry `json:"classes"`
	NodeLog *logmem.Snapshot      `json:"nodeLog,omitempty"`
}

// MarshalJSON exports Predictor state for MemoryStorage.
func (p *Predictor) MarshalJSON() ([]byte, error) {
	payload := predictorJSON{
		Storage: p.storage,
		ScopeID: p.scopeID,
		Classes: p.classes,
		NodeLog: nil,
	}

	nodeLog, ok := p.predictor.(*logmem.NodeLog)
	if !ok {
		if p.predictor != nil {
			return nil, errUnsupportedPredictorImpl
		}

		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal predictor JSON: %w", err)
		}

		return raw, nil
	}

	snapshot := nodeLog.Snapshot()
	payload.NodeLog = &snapshot

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal predictor JSON: %w", err)
	}

	return raw, nil
}

// UnmarshalJSON imports Predictor state previously exported by MarshalJSON.
func (p *Predictor) UnmarshalJSON(data []byte) error {
	var payload predictorJSON

	err := json.Unmarshal(data, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal predictor JSON: %w", err)
	}

	p.scopeID = payload.ScopeID
	p.storage = payload.Storage

	p.hasher = NewDefaultHasher()
	if payload.Classes == nil {
		p.classes = make(map[uint64]classEntry)
	} else {
		p.classes = payload.Classes
	}

	if payload.NodeLog != nil {
		p.predictor = logmem.NewFromSnapshot(*payload.NodeLog)

		return nil
	}

	predictor, err := newNodeLogger(p.storage, p.scopeID)
	if err != nil {
		return fmt.Errorf("failed to initialize predictor from JSON: %w", err)
	}

	p.predictor = predictor

	return nil
}

func (p *Predictor) addClass(class uint64, raw any) {
	p.classes[class] = classEntry{ID: class, Raw: raw}
}

// signedToUint64 converts signed integers to unsigned by preserving their
// two's complement bit pattern for deterministic ID generation.
func signedToUint64[T ~int | ~int16 | ~int32 | ~int64](v T) uint64 {
	return uint64(v) // #nosec
}

// convAnyToUint64 converts any type to a uint64 using appropriate casting rules.
//
//nolint:cyclop,varnamelen // necessary case complexity and switch variable name
func convAnyToUint64(i any) (uint64, error) {
	switch v := i.(type) {
	case uint64:
		return v, nil
	case uint32:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case int64:
		return signedToUint64(v), nil
	case int32:
		return signedToUint64(v), nil
	case int16:
		return signedToUint64(v), nil
	case int:
		return signedToUint64(v), nil
	case float64:
		return uint64(v), nil
	case float32:
		return uint64(v), nil
	case string:
		h := blake3.Sum512([]byte(v))

		return binary.BigEndian.Uint64(h[:]), nil
	case bool:
		if v {
			return 1, nil
		}

		return 0, nil
	default:
		return 0, fmt.Errorf("%w: %v", errConvAnyToUint64Unsupported, reflect.TypeOf(i))
	}
}
