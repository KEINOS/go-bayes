// Package nodelogger defines interfaces for node logging implementations.
package nodelogger

// NodeLogger logs node transition statistics used by Bayesian prediction.
//
// Each uint64 argument is a node ID and is treated as an item ID.
type NodeLogger interface {
	// ID returns the ID of the logger.
	ID() uint64
	// Predict returns the probability of the next node to be toNodeB if the
	// incoming node is fromNodeA.
	Predict(fromNodeA, toNodeB uint64) float64
	// PriorProbTo returns the prior probability of the node to be B.
	// Which is the number of accesses to the node B divided by the total number
	// of accesses of current node.
	PriorProbTo(nodeB uint64) float64
	// PriorProbFromTo returns the prior probability of the node to be B if the
	// previous node is A.
	PriorProbFromTo(fromA, toB uint64) float64
	// PriorProbNotFromTo returns the prior probability of the node not to be B
	// if the previous node is A.
	PriorProbNotFromTo(fromA, toB uint64) float64
	// Update updates the records of a node. It must be called by the next node
	// accessed.
	Update(fromA, toB uint64)
}
