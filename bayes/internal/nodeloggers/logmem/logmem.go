/*
Package logmem is an implementation of bayes.NodeLogger for memory-based logging.
*/
package logmem

import (
	"strconv"

	"github.com/KEINOS/go-bayes/bayes/internal/theorem"
	"github.com/KEINOS/go-bayes/bayes/nodelogger"
)

// Compile-time check that *NodeLog satisfies nodelogger.NodeLogger.
var _ nodelogger.NodeLogger = (*NodeLog)(nil)

// ============================================================================
//  Type: NodeLog
// ============================================================================

// NodeLog holds the records of a node. It implements nodelogger.NodeLogger for
// memory-based logging.
//
//nolint:recvcheck // allow pointer receiver for setter method
type NodeLog struct {
	// fromAToB is the number of accesses from node A to node B as map[A]map[B].
	// A is the incoming access and B is the outgoing access.
	fromAToB map[uint64]map[uint64]int
	// fromA is the number of incoming accesses from node A as map[A].
	fromA map[uint64]int
	// toB is the number of outgoing accesses to node B as map[B].
	toB map[uint64]int
	// nodeID is the node ID of the current node.
	nodeID uint64
	// totalAccesses is the total number of accesses to the node.
	totalAccesses int
}

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

// New returns a new NodeLog instance.
func New(nodeID uint64) *NodeLog {
	return &NodeLog{
		nodeID:        nodeID,
		totalAccesses: 0,
		fromAToB:      make(map[uint64]map[uint64]int),
		fromA:         make(map[uint64]int),
		toB:           make(map[uint64]int),
	}
}

// ----------------------------------------------------------------------------
//  Methods
// ----------------------------------------------------------------------------

// ID returns the node ID of the current node.
func (n NodeLog) ID() uint64 {
	return n.nodeID
}

// Predict returns the score for toNodeB after fromNodeA.
func (n NodeLog) Predict(fromNodeA, toNodeB uint64) float64 {
	// Base frequency of going to node B.
	PriorProbToB := n.PriorProbTo(toNodeB)
	// Frequency of the A-to-B pair.
	PriorProbFromAtoB := n.PriorProbFromTo(fromNodeA, toNodeB)
	// Frequency of A followed by a value other than B.
	PriorProbNotFromAtoB := n.PriorProbNotFromTo(fromNodeA, toNodeB)

	return theorem.Bayes(PriorProbToB, PriorProbFromAtoB, PriorProbNotFromAtoB)
}

// PriorProbFromTo returns the frequency of the fromA-toB pair among all updates.
func (n NodeLog) PriorProbFromTo(fromA, toB uint64) float64 {
	if n.totalAccesses == 0 {
		return 0
	}

	return float64(n.fromAToB[fromA][toB]) / float64(n.totalAccesses)
}

// PriorProbNotFromTo returns the frequency of fromA followed by a value other
// than toB among all updates.
func (n NodeLog) PriorProbNotFromTo(fromA, toB uint64) float64 {
	if n.totalAccesses == 0 {
		return 0
	}

	notA := n.fromA[fromA] - n.fromAToB[fromA][toB]

	return float64(notA) / float64(n.totalAccesses)
}

// PriorProbTo returns the frequency of nodeB among all outgoing updates.
func (n NodeLog) PriorProbTo(nodeB uint64) float64 {
	if n.totalAccesses == 0 {
		return 0
	}

	return float64(n.toB[nodeB]) / float64(n.totalAccesses)
}

// TotalAccesses returns the total number of updates recorded.
func (n NodeLog) TotalAccesses() int {
	return n.totalAccesses
}

// FromCount returns the number of incoming accesses from a given node.
func (n NodeLog) FromCount(fromA uint64) int {
	return n.fromA[fromA]
}

// ToCount returns the number of outgoing accesses to a given node.
func (n NodeLog) ToCount(toB uint64) int {
	return n.toB[toB]
}

// String returns a string representation of the NodeLog which is the node ID.
func (n NodeLog) String() string {
	return strconv.FormatUint(n.nodeID, 10)
}

// Update updates the records of a node.
// It must be called by the next node accessed.
func (n *NodeLog) Update(fromA, toB uint64) {
	if _, ok := n.fromAToB[fromA]; !ok {
		n.fromAToB[fromA] = make(map[uint64]int)
	}

	n.totalAccesses++
	n.fromA[fromA]++
	n.toB[toB]++
	n.fromAToB[fromA][toB]++
}
