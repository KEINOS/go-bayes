package bayes

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/KEINOS/go-bayes/bayes/modelstore"
)

const (
	itemDomain           = byte(0x01)
	contextDomain        = byte(0x02)
	itemHeaderLength     = 2
	bytesPerID           = 8
	fixedItemLength      = 10
	payloadFloat32Length = 4
	payloadUint64Length  = 8
	tagBool              = byte(0x01)
	tagString            = byte(0x02)
	tagInt               = byte(0x03)
	tagInt16             = byte(0x04)
	tagInt32             = byte(0x05)
	tagInt64             = byte(0x06)
	tagUint              = byte(0x07)
	tagUint16            = byte(0x08)
	tagUint32            = byte(0x09)
	tagUint64            = byte(0x0a)
	tagFloat32           = byte(0x0b)
	tagFloat64           = byte(0x0c)
)

var errInvalidStoredClass = errors.New("invalid stored class")

//nolint:cyclop,funlen // the type switch defines the persisted value contract.
func (p *Predictor) classRecord(id uint64, value any) (modelstore.Class, error) {
	record := modelstore.Class{ID: id, TypeTag: 0, Payload: nil}

	var bits [8]byte

	switch typed := value.(type) {
	case bool:
		record.TypeTag = tagBool

		record.Payload = []byte{0}
		if typed {
			record.Payload[0] = 1
		}
	case string:
		record.TypeTag = tagString
		record.Payload = []byte(typed)
	case int:
		record.TypeTag = tagInt

		binary.BigEndian.PutUint64(bits[:], uint64(int64(typed))) // #nosec G115 -- preserve bits.
		record.Payload = append([]byte(nil), bits[:]...)
	case int16:
		record.TypeTag = tagInt16

		binary.BigEndian.PutUint64(bits[:], uint64(int64(typed))) // #nosec G115 -- preserve bits.
		record.Payload = append([]byte(nil), bits[:]...)
	case int32:
		record.TypeTag = tagInt32

		binary.BigEndian.PutUint64(bits[:], uint64(int64(typed))) // #nosec G115 -- preserve bits.
		record.Payload = append([]byte(nil), bits[:]...)
	case int64:
		record.TypeTag = tagInt64

		binary.BigEndian.PutUint64(bits[:], uint64(typed)) // #nosec G115 -- preserve bits.
		record.Payload = append([]byte(nil), bits[:]...)
	case uint:
		record.TypeTag = tagUint

		binary.BigEndian.PutUint64(bits[:], uint64(typed))
		record.Payload = append([]byte(nil), bits[:]...)
	case uint16:
		record.TypeTag = tagUint16

		binary.BigEndian.PutUint64(bits[:], uint64(typed))
		record.Payload = append([]byte(nil), bits[:]...)
	case uint32:
		record.TypeTag = tagUint32

		binary.BigEndian.PutUint64(bits[:], uint64(typed))
		record.Payload = append([]byte(nil), bits[:]...)
	case uint64:
		record.TypeTag = tagUint64

		binary.BigEndian.PutUint64(bits[:], typed)
		record.Payload = append([]byte(nil), bits[:]...)
	case float32:
		record.TypeTag = tagFloat32

		var short [4]byte
		binary.BigEndian.PutUint32(short[:], math.Float32bits(typed))
		record.Payload = append([]byte(nil), short[:]...)
	case float64:
		record.TypeTag = tagFloat64

		binary.BigEndian.PutUint64(bits[:], math.Float64bits(typed))
		record.Payload = append([]byte(nil), bits[:]...)
	default:
		return modelstore.Class{}, fmt.Errorf("%w: %v", errUnsupportedValueType, reflect.TypeOf(value))
	}

	return record, nil
}

//nolint:cyclop,funlen,gocognit,gocyclo // the switch is the persisted value contract.
func (p *Predictor) decodeClass(class modelstore.Class) (any, error) {
	invalidLength := func(want int) error {
		return fmt.Errorf(
			"%w: tag %d payload length %d, want %d",
			errInvalidStoredClass,
			class.TypeTag,
			len(class.Payload),
			want,
		)
	}
	read64 := func() (uint64, error) {
		if len(class.Payload) != payloadUint64Length {
			return 0, invalidLength(payloadUint64Length)
		}

		return binary.BigEndian.Uint64(class.Payload), nil
	}

	var (
		raw any
		err error
	)

	switch class.TypeTag {
	case tagBool:
		if len(class.Payload) != 1 {
			return nil, invalidLength(1)
		}

		if class.Payload[0] > 1 {
			return nil, fmt.Errorf("%w: invalid Boolean byte %d", errInvalidStoredClass, class.Payload[0])
		}

		raw = class.Payload[0] == 1
	case tagString:
		raw = string(class.Payload)
	case tagInt:
		var bits uint64

		bits, err = read64()

		value := int64(bits) // #nosec G115 -- restore persisted two's-complement bits.
		if err == nil && strconv.IntSize == 32 && (value < math.MinInt32 || value > math.MaxInt32) {
			err = fmt.Errorf("%w: int value does not fit this architecture", errInvalidStoredClass)
		}

		if err == nil {
			raw = int(value)
		}
	case tagInt16:
		var bits uint64

		bits, err = read64()

		value := int64(bits) // #nosec G115 -- restore persisted two's-complement bits.
		if err == nil && (value < math.MinInt16 || value > math.MaxInt16) {
			err = fmt.Errorf("%w: noncanonical int16", errInvalidStoredClass)
		}

		if err == nil {
			raw = int16(value) // #nosec G115 -- range checked above.
		}
	case tagInt32:
		var bits uint64

		bits, err = read64()

		value := int64(bits) // #nosec G115 -- restore persisted two's-complement bits.
		if err == nil && (value < math.MinInt32 || value > math.MaxInt32) {
			err = fmt.Errorf("%w: noncanonical int32", errInvalidStoredClass)
		}

		if err == nil {
			raw = int32(value) // #nosec G115 -- range checked above.
		}
	case tagInt64:
		var bits uint64

		bits, err = read64()
		if err == nil {
			raw = int64(bits) // #nosec G115 -- restore persisted two's-complement bits.
		}
	case tagUint:
		var bits uint64

		bits, err = read64()
		if err == nil && strconv.IntSize == 32 && bits > math.MaxUint32 {
			err = fmt.Errorf("%w: uint value does not fit this architecture", errInvalidStoredClass)
		}

		if err == nil {
			raw = uint(bits)
		}
	case tagUint16:
		var bits uint64

		bits, err = read64()
		if err == nil && bits > math.MaxUint16 {
			err = fmt.Errorf("%w: noncanonical uint16", errInvalidStoredClass)
		}

		if err == nil {
			raw = uint16(bits) // #nosec G115 -- range checked above.
		}
	case tagUint32:
		var bits uint64

		bits, err = read64()
		if err == nil && bits > math.MaxUint32 {
			err = fmt.Errorf("%w: noncanonical uint32", errInvalidStoredClass)
		}

		if err == nil {
			raw = uint32(bits) // #nosec G115 -- range checked above.
		}
	case tagUint64:
		var bits uint64

		bits, err = read64()
		if err == nil {
			raw = bits
		}
	case tagFloat32:
		if len(class.Payload) != payloadFloat32Length {
			return nil, invalidLength(payloadFloat32Length)
		}

		raw = math.Float32frombits(binary.BigEndian.Uint32(class.Payload))
	case tagFloat64:
		var bits uint64

		bits, err = read64()
		if err == nil {
			raw = math.Float64frombits(bits)
		}
	default:
		return nil, fmt.Errorf("%w: unknown tag %d", errInvalidStoredClass, class.TypeTag)
	}

	if err != nil {
		return nil, err
	}

	id, hashErr := p.itemID(raw)
	if hashErr != nil {
		return nil, fmt.Errorf("%w: failed to hash class: %w", errInvalidStoredClass, hashErr)
	}

	if id != class.ID {
		return nil, fmt.Errorf("%w: class ID mismatch", errInvalidStoredClass)
	}

	return raw, nil
}

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
		return encode64(tagInt, uint64(int64(typed))), nil // #nosec G115 -- preserve bits.
	case int16:
		return encode64(tagInt16, uint64(int64(typed))), nil // #nosec G115 -- preserve bits.
	case int32:
		return encode64(tagInt32, uint64(int64(typed))), nil // #nosec G115 -- preserve bits.
	case int64:
		return encode64(tagInt64, uint64(typed)), nil // #nosec G115 -- preserve bits.
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
