package bayes

import (
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

// Type returns the type name of the storage.
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
