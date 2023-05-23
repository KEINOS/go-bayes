/*
Package ksv defines the interface for a key-value store.
*/
package kvs

import "github.com/KEINOS/go-bayes/pkg/kvs/inmemory"

// ----------------------------------------------------------------------------
//  Type: KVS (interface)
// ----------------------------------------------------------------------------

// DB is the interface for a key-value store.
// This interface is also compatible with iterator.Iterator interface as well.
type DB interface {
	// Delete deletes the value for the given key.
	Delete(key uint64) error
	// GetKeys returns the keys in the KVS.
	GetKeys() []uint64
	// GetValue returns the value for the given key.
	GetValue(key uint64) (any, error)
	// HasNext returns true if there is a next element.
	HasNext() bool
	// Load loads the KVS data from a file.
	Load(filePath string) error
	// Next returns the next key-value pair.
	Next() (uint64, any, error)
	// RegistGob registers the given struct to the gob registry.
	RegistGob(value any)
	// Reset resets the iterator to the first element.
	Reset()
	// Save saves the KVS data to a file.
	Save(filePath string) error
	// Set sets the value for the given key.
	Set(key uint64, value any) error
}

// ----------------------------------------------------------------------------
//  Type: DBType (Enum for DB type)
// ----------------------------------------------------------------------------

// DBType is the type of the key-value store.
type DBType int

// Enum of the key-value store types.
const (
	// UnknownDB represents an unknown DB type.
	UnknownDB DBType = iota
	// InMemoryDB represents an in-memory DB type.
	InMemoryDB
)

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

// New returns a new DB instance of the given DB type. If unknown DB type is
// given, nil is returned.
func New(dbType DBType) DB { //nolint:ireturn // ignore returning an interface
	var obj DB // Ensure if obj implements the DB interface

	switch dbType {
	case InMemoryDB:
		obj = inmemory.New()
	case UnknownDB:
		fallthrough
	default:
		return nil
	}

	return obj
}
