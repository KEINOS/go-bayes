// Package nodelogger defines the transition-statistics storage interface.
package nodelogger

// NodeLogger records transition counts and provides the probabilities used by
// Bayesian inference to score possible next-value IDs.
//
// Each uint64 argument is an item ID or a folded context ID.
type NodeLogger interface {
	// ID returns the application-defined ID of the logger.
	ID() uint64
	// Predict returns the score for toNodeB after fromNodeA.
	Predict(fromNodeA, toNodeB uint64) float64
	// PriorProbTo returns the frequency of toNodeB among all recorded updates.
	PriorProbTo(nodeB uint64) float64
	// PriorProbFromTo returns the frequency of the fromA-toB pair among all
	// recorded updates.
	PriorProbFromTo(fromA, toB uint64) float64
	// PriorProbNotFromTo returns the frequency of fromA followed by a value other
	// than toB among all recorded updates.
	PriorProbNotFromTo(fromA, toB uint64) float64
	// Update updates the records of a node. It must be called by the next node
	// accessed.
	Update(fromA, toB uint64)
}
