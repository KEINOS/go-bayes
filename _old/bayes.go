package bayes

import (
	"io"

	"github.com/KEINOS/go-bayes/pkg/class"
	"github.com/KEINOS/go-bayes/pkg/dumptype"
	"github.com/KEINOS/go-bayes/pkg/nodelogger/logmem"
	"github.com/pkg/errors"
)

// ----------------------------------------------------------------------------
//  Type: NodeLogger
// ----------------------------------------------------------------------------

// NodeLogger is an interface to log the node's state. Here, node is something
// in between the predecessor and the successor.
//
// Each uint64 argument of the method is a node ID. Consider node IDs to be
// equivalent to item IDs.
type NodeLogger interface {
	// ID returns the ID of the logger.
	ID() uint64
	// Predict returns the probability of the next node to be toNodeB if the incoming
	// node is fromNodeA.
	Predict(fromNodeA, toNodeB uint64) float64
	// PriorPtoB returns the prior probability of the node to be B.
	// Which is the number of accesses to the node B divided by the total number
	// of accesses of current node.
	PriorPtoB(nodeB uint64) float64
	// PriorPfromAtoB returns the prior probability of the node to be B if the
	// previous node is A.
	PriorPfromAtoB(fromA, toB uint64) float64
	// PriorPNotFromAtoB returns the prior probability of the node not to be B
	// if the previous node is A.
	PriorPNotFromAtoB(fromA, toB uint64) float64
	// Restore restores the records of the logger in dt format from r.
	Restore(dt dumptype.DumpType, r io.Reader) error
	// Store stores the records of the logger and the class.Map in dt format to w.
	Store(dt dumptype.DumpType, classMap class.Map, w io.Writer) error
	// Update updates the records of a node. It must be called by the next node
	// accessed.
	Update(fromA, toB uint64)
}

// ----------------------------------------------------------------------------
//  Type: Storage
// ----------------------------------------------------------------------------

// Storage is the type of storage to log the accesses.
type Storage int

const (
	// UnknwonStorage represents the unknown storage.
	UnknwonStorage Storage = iota
	// MemoryStorage represents the in-memory storage.
	MemoryStorage
	// SQLite3Storage represents the SQLite3 as a storage.
	SQLite3Storage
)

// IsUnknown returns true if the storage type is unknown.
func (s Storage) IsUnknown() bool {
	switch s {
	case MemoryStorage, SQLite3Storage:
		return false
	case UnknwonStorage:
		fallthrough
	default:
		return true
	}
}

// String is the implementation of the Stringer interface.
// It returns the type name of the storage. Alias of Type().
func (s Storage) String() string {
	return s.Type()
}

// Type returns the type name of the storage. Equivalent to String().
func (s Storage) Type() string {
	switch s {
	case MemoryStorage:
		return "in-memory"
	case SQLite3Storage:
		return "SQLite3"
	}

	return "unknown"
}

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

// New returns a new NodeLogger instance.
//
// Use this function if you want to have more control over the NodeLogger
// instance rather than using the convenient functions.
func New(engine Storage, scopeID uint64) (NodeLogger, error) {
	switch engine {
	case MemoryStorage:
		return logmem.New(scopeID), nil
	}

	return nil, errors.New("unknown storage engine type")
}
