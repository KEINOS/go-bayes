/*
Package logmem is an implementation of bayes.NodeLogger for memory-based logging.
*/
package logmem

import (
	"fmt"

	"github.com/KEINOS/go-bayes/pkg/theorem"
)

// ----------------------------------------------------------------------------
//  Type: NodeLog
// ----------------------------------------------------------------------------

// NodeLog holds the records of a node. It is an implementation of bayes.NodeLogger
// for memory-based logging.
type NodeLog struct {
	// FromAtoB is the number of accesses from node A to node B as map[A]map[B].
	// A is the incoming access and B is the outgoing access.
	FromAToB map[uint64]map[uint64]int
	// FromA is the number of incoming accesses from node A as map[A].
	FromA map[uint64]int
	// ToB is the number of outgoing accesses to node B as map[B].
	ToB map[uint64]int
	// nodeID is the node ID of the current node.
	nodeID uint64
	// TotalAccesses is the total number of accesses to the node.
	TotalAccesses int
}

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

// New returns a new NodeLog instance.
func New(nodeID uint64) *NodeLog {
	return &NodeLog{
		nodeID:        nodeID,
		TotalAccesses: 0,
		FromAToB:      make(map[uint64]map[uint64]int),
		FromA:         make(map[uint64]int),
		ToB:           make(map[uint64]int),
	}
}

// ----------------------------------------------------------------------------
//  Methods
// ----------------------------------------------------------------------------

// ID returns the node ID of the current node.
func (n NodeLog) ID() uint64 {
	return n.nodeID
}

// Predict returns the probability of the next node to be toNodeB if the incoming
// node is fromNodeA.
func (n NodeLog) Predict(fromNodeA, toNodeB uint64) float64 {
	// Prior probability of the next node to be node B.
	PriorProbToB := n.PriorPtoB(toNodeB)
	// Prior probability of the incoming node to be node B if the previous node was A.
	PriorProbFromAtoB := n.PriorPfromAtoB(fromNodeA, toNodeB)
	// Prior probability of the incoming node not to be node B if the previous node was A.
	PriorProbNotFromAtoB := n.PriorPNotFromAtoB(fromNodeA, toNodeB)

	return theorem.Bayes(PriorProbToB, PriorProbFromAtoB, PriorProbNotFromAtoB)
}

// PriorPfromAtoB returns the prior probability of the node to be B if the
// previous node is A.
func (n NodeLog) PriorPfromAtoB(fromA, toB uint64) float64 {
	if n.TotalAccesses == 0 {
		return 0
	}

	return float64(n.FromAToB[fromA][toB]) / float64(n.TotalAccesses)
}

// PriorPNotFromAtoB returns the prior probability of the node not to be B
// if the previous node is A.
func (n NodeLog) PriorPNotFromAtoB(fromA, toB uint64) float64 {
	if n.TotalAccesses == 0 {
		return 0
	}

	notA := n.FromA[fromA] - n.FromAToB[fromA][toB]

	return float64(notA) / float64(n.TotalAccesses)
}

// PriorPtoB returns the prior probability of the outgoing node to be nodeB.
//
// Which is the number of outgoing accesses to the node B out of the total number
// of accesses of current node.
func (n NodeLog) PriorPtoB(nodeB uint64) float64 {
	if n.TotalAccesses == 0 {
		return 0
	}

	return float64(n.ToB[nodeB]) / float64(n.TotalAccesses)
}

// String returns a string representation of the NodeLog which is the node ID.
func (n NodeLog) String() string {
	return fmt.Sprintf("%d", n.nodeID)
}

// Update updates the records of a node.
// It must be called by the next node accessed.
func (n *NodeLog) Update(fromA, toB uint64) {
	if _, ok := n.FromAToB[fromA]; !ok {
		n.FromAToB[fromA] = make(map[uint64]int)
	}

	n.TotalAccesses++
	n.FromA[fromA]++
	n.ToB[toB]++
	n.FromAToB[fromA][toB]++
}
