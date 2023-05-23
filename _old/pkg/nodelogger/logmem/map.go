package logmem

import (
	"sort"

	"github.com/KEINOS/go-bayes/pkg/class"
	"github.com/pkg/errors"
)

// ----------------------------------------------------------------------------
//  Type: Map
// ----------------------------------------------------------------------------

// Map is the map of the Class type.
//
// It is an implementation of the class.Map interface. It is used as a key-value
// store or hash-table-like usage.
type Map map[uint64]class.Class

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

// NewMap creates a new Map instance.
func NewMap(IDs ...any) (Map, error) {
	lenIDs := len(IDs)
	mapClasses := make(Map, lenIDs)

	if lenIDs > 0 && IDs[0] != nil {
		if err := mapClasses.AddAny(IDs...); err != nil {
			return nil, errors.Wrap(err, "failed to add Class objects")
		}
	}

	return mapClasses, nil
}

// ----------------------------------------------------------------------------
//  Methods
// ----------------------------------------------------------------------------

// AddAny adds any type value as a Class object to the Map object.
func (m *Map) AddAny(items ...any) error {
	for _, itemRaw := range items {
		item, err := class.New(itemRaw)
		if err != nil {
			return errors.Wrap(err, "failed to instantiate Map object")
		}

		// Is existing class ID?
		if _, ok := (*m)[item.ID]; ok {
			if (*m)[item.ID].Raw == item.Raw {
				continue
			}

			return errors.Errorf(
				"duplicate class ID with different type of value. Original: %T, New: %T, Value: %v",
				(*m)[item.ID].Raw,
				item.Raw,
				item,
			)
		}

		(*m)[item.ID] = *item
	}

	return nil
}

func (m *Map) GetClass(classID uint64) (item class.Class, err error) {
	if _, ok := (*m)[classID]; !ok {
		return class.Class{}, errors.Errorf("class ID %v is not in the map", classID)
	}

	return (*m)[classID], nil
}

// GetClassID returns the class ID of the given value.
// It will return an error if the item is not in the map.
func (m *Map) GetClassID(item any) (classID uint64, err error) {
	c, err := class.New(item)
	if err != nil {
		return 0, errors.Wrap(err, "failed to instantiate Class object")
	}

	if _, ok := (*m)[c.ID]; !ok {
		return 0, errors.Errorf("item %v is not in the map", c)
	}

	return c.ID, nil
}

// GetKeys returns the sorted list of class IDs in the Map object.
func (m Map) GetKeys() []uint64 {
	keys := make([]uint64, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}
