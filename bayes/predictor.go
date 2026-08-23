package bayes

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/KEINOS/go-bayes/bayes/internal/nodeloggers/logmem"
	"github.com/KEINOS/go-bayes/bayes/nodelogger"
)

const (
	predictorJSONSchemaVersion = 1
	itemDomain                 = byte(0x01)
	contextDomain              = byte(0x02)
	itemHeaderLength           = 2
	bytesPerID                 = 8
	fixedItemLength            = 10
	tagBool                    = byte(0x01)
	tagString                  = byte(0x02)
	tagInt                     = byte(0x03)
	tagInt16                   = byte(0x04)
	tagInt32                   = byte(0x05)
	tagInt64                   = byte(0x06)
	tagUint                    = byte(0x07)
	tagUint16                  = byte(0x08)
	tagUint32                  = byte(0x09)
	tagUint64                  = byte(0x0a)
	tagFloat32                 = byte(0x0b)
	tagFloat64                 = byte(0x0c)
)

// classEntry holds the class ID and the original value.
// This type is used internally by Predictor to map class IDs to their
// original raw values (used for GetClass and JSON serialization).
type classEntry struct {
	Raw any    `json:"raw"`
	ID  uint64 `json:"id"`
}

var (
	errPredictorHasherUninit      = errors.New("hasher is not initialized")
	errPredictorHasherNameEmpty   = errors.New("hasher name must not be empty")
	errPredictorHasherMismatch    = errors.New("snapshot hasher does not match predictor hasher")
	errPredictorSchemaUnsupported = errors.New("unsupported predictor JSON schema version")
	errPredictorNotInitialized    = errors.New("predictor is not initialized")
	errPredictorItemsNil          = errors.New("items must not be nil")
	errPredictorItemsNotSlice     = errors.New("items must be a slice")
	errUnsupportedPredictorImpl   = errors.New("unsupported predictor implementation")
	errUnsupportedValueType       = errors.New("unsupported value type")
)

// PredictorConfig defines the storage, scope, and value/context hasher used by a
// Predictor. A nil Hasher selects [NewDefaultHasher].
type PredictorConfig struct {
	Storage Storage
	ScopeID uint64
	Hasher  Hasher
}

// Predictor learns transitions from folded ordered contexts to possible next
// values. Each instance owns its transition statistics and class map.
//
// Predictor is not safe for concurrent use from multiple goroutines.
// Use external synchronization if needed.
type Predictor struct {
	predictor nodelogger.NodeLogger
	classes   map[uint64]classEntry
	storage   Storage
	scopeID   uint64
	hasher    Hasher
	scratch   codecScratch
}

type codecScratch struct {
	bytes []byte
	ids   []uint64
}

// NewPredictor creates a Predictor from config. A nil config Hasher selects
// [NewDefaultHasher], which currently returns the xxHash3 implementation.
func NewPredictor(config PredictorConfig) (*Predictor, error) {
	hasher := config.Hasher
	if hasher == nil {
		hasher = NewDefaultHasher()
	}

	if hasher.Name() == "" {
		return nil, errPredictorHasherNameEmpty
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
		scratch:   codecScratch{bytes: nil, ids: nil},
	}, nil
}

// ID returns the current scope ID.
func (p *Predictor) ID() uint64 {
	if p.predictor != nil {
		return p.predictor.ID()
	}

	return p.scopeID
}

// SetStorage selects the storage backend that the next [Predictor.Reset] uses.
// It does not move or modify the predictor's current learned state.
func (p *Predictor) SetStorage(storage Storage) {
	p.storage = storage
}

// Reset recreates the storage backend and clears all learned transitions and
// class values. It preserves the current scope ID and hasher.
func (p *Predictor) Reset() error {
	predictor, err := newNodeLogger(p.storage, p.scopeID)
	if err != nil {
		return fmt.Errorf("failed to reset predictor: %w", err)
	}

	p.predictor = predictor
	p.classes = make(map[uint64]classEntry)

	return nil
}

// GetClass resolves classID to the original value recorded during training. It
// returns nil when classID is unknown.
func (p *Predictor) GetClass(classID uint64) any {
	return p.classes[classID].Raw
}

// HashTrans converts supported values to item IDs and folds their ordered
// sequence into one deterministic context ID. The result is fixed-width but is
// not guaranteed to be collision-free or reversible.
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

// Predict folds items as one ordered context, scores every learned candidate,
// and returns the highest-scoring class ID. Use [Predictor.GetClass] to resolve
// the ID to its original value.
//
// Predict does not perform similarity matching or automatic suffix backoff. It
// returns zero when no candidate receives a positive score. Zero can also be a
// valid class ID. If candidates have the same highest score, either candidate
// can be returned.
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

// Train learns each next value in an observed sequence. For every value after
// the first, it records transitions from every folded suffix of the preceding
// context. Each observed next value also becomes a candidate class that
// [Predictor.Predict] can return.
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

	drill := []uint64{}

	for index, itemRaw := range itemsAny {
		item, err := p.itemID(itemRaw)
		if err != nil {
			return fmt.Errorf("failed during training iteration: %w", err)
		}

		if index == 0 {
			drill = append(drill, item)

			continue
		}

		for i := range drill {
			flowID := p.contextID(drill[i:])
			p.predictor.Update(flowID, item)
		}

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
	SchemaVersion int                   `json:"schemaVersion"`
	Hasher        string                `json:"hasher"`
	Storage       Storage               `json:"storage"`
	ScopeID       uint64                `json:"scopeId"`
	Classes       map[uint64]classEntry `json:"classes"`
	NodeLog       *logmem.Snapshot      `json:"nodeLog,omitempty"`
}

// MarshalJSON exports MemoryStorage transition state, scope, storage type, and
// class values. JSON does not preserve the exact Go type of numeric class
// values. The schema version and selected hasher name are also stored.
func (p *Predictor) MarshalJSON() ([]byte, error) {
	if p.hasher == nil {
		return nil, errPredictorHasherUninit
	}

	if p.hasher.Name() == "" {
		return nil, errPredictorHasherNameEmpty
	}

	payload := predictorJSON{
		SchemaVersion: predictorJSONSchemaVersion,
		Hasher:        p.hasher.Name(),
		Storage:       p.storage,
		ScopeID:       p.scopeID,
		Classes:       p.classes,
		NodeLog:       nil,
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

// UnmarshalJSON imports state previously exported by [Predictor.MarshalJSON].
// JSON numeric class values are restored as float64. Built-in hashers are
// restored by name. A custom hasher must be injected into the receiver before
// unmarshaling, and its name must match the snapshot.
func (p *Predictor) UnmarshalJSON(data []byte) error {
	var payload predictorJSON

	err := json.Unmarshal(data, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal predictor JSON: %w", err)
	}

	if payload.SchemaVersion != predictorJSONSchemaVersion {
		return fmt.Errorf("%w: %d", errPredictorSchemaUnsupported, payload.SchemaVersion)
	}

	hasher := p.hasher
	if hasher == nil || hasher.Name() != payload.Hasher {
		var builtIn bool

		hasher, builtIn = newBuiltInHasher(payload.Hasher)
		if !builtIn {
			return fmt.Errorf("%w: %q", errPredictorHasherMismatch, payload.Hasher)
		}
	}

	p.scopeID = payload.ScopeID
	p.storage = payload.Storage

	p.hasher = hasher
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

// contextID hashes the canonical representation of an ordered item ID list.
func (p *Predictor) contextID(itemIDs []uint64) uint64 {
	maxLength := 1 + binary.MaxVarintLen64 + len(itemIDs)*bytesPerID
	p.scratch.bytes = ensureByteCapacity(p.scratch.bytes, maxLength)
	encoded := p.scratch.bytes[:maxLength]
	encoded[0] = contextDomain
	countLength := binary.PutUvarint(encoded[1:], uint64(len(itemIDs)))
	encoded = encoded[:1+countLength+len(itemIDs)*bytesPerID]

	offset := 1 + countLength
	for _, itemID := range itemIDs {
		binary.BigEndian.PutUint64(encoded[offset:], itemID)
		offset += bytesPerID
	}

	return p.hasher.Hash(encoded)
}

// itemID hashes the type-preserving canonical representation of a value.
//
//nolint:cyclop,funlen // the type switch defines the supported value contract.
func (p *Predictor) itemID(value any) (uint64, error) {
	p.scratch.bytes = ensureByteCapacity(p.scratch.bytes, fixedItemLength)
	fixed := p.scratch.bytes[:fixedItemLength]
	fixed[0] = itemDomain

	encode64 := func(tag byte, payload uint64) uint64 {
		fixed[1] = tag
		binary.BigEndian.PutUint64(fixed[2:], payload)

		return p.hasher.Hash(fixed)
	}

	switch typed := value.(type) {
	case bool:
		fixed[1] = tagBool

		fixed[2] = 0
		if typed {
			fixed[2] = 1
		}

		return p.hasher.Hash(fixed[:3]), nil
	case string:
		encodedLength := itemHeaderLength + len(typed)
		p.scratch.bytes = ensureByteCapacity(p.scratch.bytes, encodedLength)
		encoded := p.scratch.bytes[:encodedLength]
		encoded[0] = itemDomain
		encoded[1] = tagString
		copy(encoded[2:], typed)

		return p.hasher.Hash(encoded), nil
	case int:
		return encode64(tagInt, uint64(int64(typed))), nil // #nosec G115 -- preserve two's complement bits.
	case int16:
		return encode64(tagInt16, uint64(int64(typed))), nil // #nosec G115 -- preserve two's complement bits.
	case int32:
		return encode64(tagInt32, uint64(int64(typed))), nil // #nosec G115 -- preserve two's complement bits.
	case int64:
		return encode64(tagInt64, uint64(typed)), nil // #nosec G115 -- preserve two's complement bits.
	case uint:
		return encode64(tagUint, uint64(typed)), nil
	case uint16:
		return encode64(tagUint16, uint64(typed)), nil
	case uint32:
		return encode64(tagUint32, uint64(typed)), nil
	case uint64:
		return encode64(tagUint64, typed), nil
	case float32:
		return encode64(tagFloat32, uint64(math.Float32bits(typed))), nil
	case float64:
		return encode64(tagFloat64, math.Float64bits(typed)), nil
	default:
		return 0, fmt.Errorf("%w: %v", errUnsupportedValueType, reflect.TypeOf(value))
	}
}

func ensureByteCapacity(buffer []byte, length int) []byte {
	if cap(buffer) < length {
		return make([]byte, length)
	}

	return buffer[:length]
}

func ensureUint64Capacity(buffer []uint64, length int) []uint64 {
	if cap(buffer) < length {
		return make([]uint64, 0, length)
	}

	return buffer[:0]
}
