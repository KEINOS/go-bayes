package bayes

import (
	"encoding/binary"
	"hash/crc32"
	"os"
	"unsafe"

	"github.com/KEINOS/go-bayes/pkg/dumptype"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
)

// ============================================================================
//  Shorthand functions.
// ============================================================================
//  This file contains functions for easy-to-use purposes.
//
//  Convenient Functions:
//    - SetStorage() - Sets the storage used by the predictor.
//    - Reset() - Resets the trained data of the predictor.
//    - Train() - Trains the predictor with the given items.
//    - Predict() - Predicts the next item from the given items.
//    - HashTrans() - Returns a unique hash from the input items.
//    - GetClass() - Returns the original item value of the given class ID.
// ============================================================================

const (
	// StorageDefault is the default storage used by the predictor. Currently, it
	// is an in-memory log (logmem package).
	StorageDefault = MemoryStorage
	// ScopeIDDefault is the default scope ID on creating an instance of the
	// predictor.
	ScopeIDDefault = uint64(0)
)

var (
	_classes   map[uint64]_Class
	_predictor NodeLogger
	_storage   = StorageDefault
	OldClasses map[uint64]_Class
)

func init() {
	Reset()
}

// ----------------------------------------------------------------------------
//  Type: _Class (private)
// ----------------------------------------------------------------------------

// _Class holds the class ID and the original value.
type _Class struct {
	Raw any
	ID  uint64
}

// ----------------------------------------------------------------------------
//  Public functions
// ----------------------------------------------------------------------------

// GetClass returns the original value of the given class ID.
func GetClass(classID uint64) any {
	return _classes[classID].Raw
}

// HashTrans returns a unique hash from the input transitions. Note that the hash
// is not cryptographically secure.
func HashTrans[T any](transitions ...T) (uint64, error) {
	// Calculate the hash of the transition.
	hashed, err := getBlake3(transitions...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get the hash of the transition")
	}

	// Calculate the CRC32C of the hash value.
	chksum := getCRC32C(hashed)

	return chopAndMergeBytes(hashed, chksum)
}

// Predict returns the next class ID inferred from the given items.
//
// To get the original value of the class, use `GetClass()`.
func Predict[T any](items []T) (classID uint64, err error) {
	if _predictor == nil {
		return 0, errors.New("predictor is not initialized")
	}

	biggest := struct {
		Probability float64
		Class       uint64
	}{
		Probability: 0,
		Class:       0,
	}

	flowID, err := HashTrans(items...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to hash the flow")
	}

	for classID := range _classes {
		probability := _predictor.Predict(flowID, classID)

		if biggest.Probability < probability {
			biggest.Probability = probability
			biggest.Class = classID
		}
	}

	return biggest.Class, nil
}

// Reset resets the train object.
func Reset() error {
	var err error

	_predictor, err = New(_storage, ScopeIDDefault)
	if err != nil {
		return errors.Wrap(err, "failed to set the predictor")
	}

	OldClasses = _classes
	_classes = make(map[uint64]_Class)

	return nil
}

// Restore loads the data of the predictor from the given filePath.
func Restore(pathFile string) error {
	// Open the file.
	fp, err := os.Open(pathFile)
	if err != nil {
		return errors.Wrap(err, "failed to open the file")
	}
	defer fp.Close()

	_predictor, _ = New(_storage, ScopeIDDefault)
	_classes = OldClasses

	return errors.Wrap(_predictor.Restore(dumptype.GOB, fp), "failed to restore the predictor")
}

// SetStorage sets the storage used by the predictor. This won't affect the
// predictors created via `New()`.
//
// Do not forget to `Reset()` the predictor after changing the storage.
func SetStorage(storageType Storage) error {
	if storageType.IsUnknown() {
		return errors.New("unknown storage type")
	}

	_storage = storageType

	return nil
}

// Store saves the data of the predictor to the given filePath.
func Store(pathFile string) error {
	// Create the file.
	fp, err := os.Create(pathFile)
	if err != nil {
		return errors.Wrap(err, "failed to open the file")
	}
	defer fp.Close()

	// HERE THE CLASS MAP SHOULD NOT BE NIL
	return errors.Wrap(_predictor.Store(dumptype.GOB, nil, fp), "failed to store the predictor")
}

// Train trains the predictor with the given items.
//
// Once the item appears in the training set, the item is added to the class list.
func Train[T any](items []T) error {
	if _predictor == nil {
		Reset()
	}

	prevItem := uint64(0)
	drill := []uint64{}

	for i, itemRaw := range items {
		item, err := convAnyToUint64(itemRaw)
		if err != nil {
			return errors.Wrap(err, "failed during training iteration")
		}

		if i == 0 {
			prevItem = item
			drill = append(drill, item)

			continue
		}

		// 101 training. Trains only the predecessor and the successor item.
		// e.g.
		//   previous items --> [1, 2, 3, 4, 5]
		//   following item --> 6
		//   will train:
		//               [5] --> 6
		_predictor.Update(prevItem, item)

		// Drill.
		// Trains by repeating the flow of the previous items.
		// e.g.
		//   previous items --> [1, 2, 3, 4, 5]
		//   following item --> 6
		//   will train:
		//               [5] --> 6
		//            [4, 5] --> 6
		//         [3, 4, 5] --> 6
		//      [2, 3, 4, 5] --> 6
		//   [1, 2, 3, 4, 5] --> 6
		for i := 0; i < len(drill); i++ {
			flowID, _ := HashTrans(drill[i:]...)

			_predictor.Update(flowID, item)
		}

		prevItem = item
		drill = append(drill, item)
		addClass(item, itemRaw)
	}

	return nil
}

// ----------------------------------------------------------------------------
//  Private functions
// ----------------------------------------------------------------------------

func addClass(class uint64, raw any) {
	switch v := raw.(type) {
	case uint64:
		_classes[class] = _Class{ID: class, Raw: v}
	case uint32:
		_classes[class] = _Class{ID: class, Raw: v}
	case uint16:
		_classes[class] = _Class{ID: class, Raw: v}
	case uint:
		_classes[class] = _Class{ID: class, Raw: v}
	case int64:
		_classes[class] = _Class{ID: class, Raw: v}
	case int32:
		_classes[class] = _Class{ID: class, Raw: v}
	case int16:
		_classes[class] = _Class{ID: class, Raw: v}
	case int:
		_classes[class] = _Class{ID: class, Raw: v}
	case float64:
		_classes[class] = _Class{ID: class, Raw: v}
	case float32:
		_classes[class] = _Class{ID: class, Raw: v}
	case string:
		_classes[class] = _Class{ID: class, Raw: v}
	case bool:
		_classes[class] = _Class{ID: class, Raw: v}
	default:
		_classes[class] = _Class{ID: class, Raw: raw}
	}
}

// chopAndMergeBytes combines the two input as one in 8 byte length.
func chopAndMergeBytes(a, b []byte) (uint64, error) {
	if len(a) < 4 || len(b) < 4 {
		return 0, errors.New("failed to combine bytes. Both of the input must be 4byte or more")
	}

	rawid := make([]byte, 8)

	_ = copy(rawid, a)     // Upper half as hash
	_ = copy(rawid[4:], b) // Bottom half as checksum

	return binary.BigEndian.Uint64(rawid), nil
}

func convAnyToUint64(i interface{}) (uint64, error) {
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
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	case float32:
		return uint64(v), nil
	case string:
		h := blake3.Sum512([]byte(v))

		return binary.BigEndian.Uint64(h[:]), nil
	case bool:
		if v {
			return uint64(1), nil
		}

		return uint64(0), nil
	}

	return 0, errors.Errorf("failed to convert to uint64. Unsupported type: %T", i)
}

// getBlake3 returns the hash of the input to byte array.
func getBlake3[T any](inputs ...T) ([]byte, error) {
	hasher := blake3.New()

	for _, v := range inputs {
		vv, err := convAnyToUint64(v)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert to uint64")
		}

		// blake3.Hasher.Write() never returns an error.
		// https://github.com/zeebo/blake3/blob/master/api.go#L87-L91
		_, _ = hasher.Write(uint64ToByteArray(vv))
	}

	return hasher.Sum(nil), nil
}

// getCRC32C returns the CRC-32 with Castagnoli polynomial of the input.
func getCRC32C(input []byte) []byte {
	crcTable := crc32.MakeTable(crc32.Castagnoli)
	crc32C := crc32.New(crcTable)

	// crc32.digest.Write() never returns an error.
	// https://cs.opensource.google/go/go/+/master:src/hash/crc32/crc32.go;l=228-240
	_, _ = crc32C.Write(input)

	return crc32C.Sum(nil)
}

func isValidPath(path string) bool {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return true
	}

	var d []byte

	// Attempt to create it
	err := os.WriteFile(path, d, 0644)
	if err == nil {
		// And delete it
		err = os.Remove(path)
	}

	if err != nil {
		return false
	}

	return true
}

// uint64ToByteArray converts an unsigned integer to a byte array in little endian.
func uint64ToByteArray(num uint64) []byte {
	size := int(unsafe.Sizeof(num))
	arr := make([]byte, size)

	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
		arr[i] = byt
	}

	return arr
}
