package class

// Map is the interface to hold a map of Class objects.
//
// It is used as a key-value store or hash-table-like usage. Each nodelogger
// must implement this interface.
type Map interface {
	// AddAny adds any type of value supported as the Class object to the list.
	// If the item is already in the list, it should be ignored.
	AddAny(items ...any) error
	// GetClassID returns the class ID of the given value. It should return an
	// error if the value is not in the list.
	GetClassID(item any) (classID uint64, err error)
	// GetKeys returns the list of class IDs in the Map object.
	GetKeys() []uint64
}
