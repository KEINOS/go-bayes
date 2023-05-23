/*
Package logmem is an implementation of bayes.NodeLogger for memory-based logging.
*/
package logmem

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"

	"github.com/KEINOS/go-bayes/pkg/class"
	"github.com/KEINOS/go-bayes/pkg/dumptype"
	"github.com/KEINOS/go-bayes/pkg/theorem"
	"github.com/pkg/errors"
)

// ----------------------------------------------------------------------------
//  Type: NodeLog
// ----------------------------------------------------------------------------

// NodeLog holds the records of a node. It is an implementation of bayes.NodeLogger
// for memory-based logging.
type NodeLog struct {
	// ClassMap holds the class.Map which is the mapping table of node IDs and its
	// actual item.
	ClassMap class.Map
	// FromAtoB is the number of accesses from node A to node B as map[A]map[B].
	// A is the incoming access and B is the outgoing access.
	FromAToB map[uint64]map[uint64]int `json:"from_a_to_b"`
	// FromA is the number of incoming accesses from node A as map[A].
	FromA map[uint64]int `json:"from_a"`
	// ToB is the number of outgoing accesses to node B as map[B].
	ToB map[uint64]int `json:"to_b"`
	// NodeID is the node ID of the current node.
	NodeID uint64
	// TotalAccesses is the total number of accesses to the node.
	TotalAccesses int `json:"total_accesses"`
}

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

// New returns a new NodeLog instance.
func New(nodeID uint64) *NodeLog {
	registGob()

	return &NodeLog{
		NodeID:        nodeID,
		TotalAccesses: 0,
		FromAToB:      make(map[uint64]map[uint64]int),
		FromA:         make(map[uint64]int),
		ToB:           make(map[uint64]int),
	}
}

var isGobRegistered = false

func registGob() {
	if isGobRegistered {
		return
	}

	gob.Register(&Map{})

	isGobRegistered = true
}

// ----------------------------------------------------------------------------
//  Common Variables
// ----------------------------------------------------------------------------

var (
	marshalIndentPrefix = ""
	marshalIndentIndent = "\t"
)

// jsonMarshalIndent is a monkey patch of json.MarshalIndent. It is a copy of
// json.MarshalIndent to ease testing.
var jsonMarshalIndent = json.MarshalIndent

// ----------------------------------------------------------------------------
//  Methods
// ----------------------------------------------------------------------------

// ID returns the node ID of the current node.
func (node NodeLog) ID() uint64 {
	return node.NodeID
}

// Predict returns the probability of the next node to be toNodeB if the incoming
// node is fromNodeA.
func (node NodeLog) Predict(fromNodeA, toNodeB uint64) float64 {
	// Prior probability of the next node to be node B.
	PriorProbToB := node.PriorPtoB(toNodeB)
	// Prior probability of the incoming node to be node B if the previous node was A.
	PriorProbFromAtoB := node.PriorPfromAtoB(fromNodeA, toNodeB)
	// Prior probability of the incoming node not to be node B if the previous node was A.
	PriorProbNotFromAtoB := node.PriorPNotFromAtoB(fromNodeA, toNodeB)

	return theorem.Bayes(PriorProbToB, PriorProbFromAtoB, PriorProbNotFromAtoB)
}

// PriorPfromAtoB returns the prior probability of the node to be B if the
// previous node is A.
func (node NodeLog) PriorPfromAtoB(fromA, toB uint64) float64 {
	if node.TotalAccesses == 0 {
		return 0
	}

	return float64(node.FromAToB[fromA][toB]) / float64(node.TotalAccesses)
}

// PriorPNotFromAtoB returns the prior probability of the node NOT to be B
// if the previous node is A.
func (node NodeLog) PriorPNotFromAtoB(fromA, toB uint64) float64 {
	// If there is no access, never reaches to node B. Or, if there is no
	// access from node A, never reaches to node B as well.
	if node.TotalAccesses == 0 || node.FromA[fromA] == 0 {
		return 1
	}

	notA := node.FromA[fromA] - node.FromAToB[fromA][toB]

	return float64(notA) / float64(node.TotalAccesses)
}

// PriorPtoB returns the prior probability of the outgoing node to be nodeB.
//
// Which is the number of outgoing accesses to the node B out of the total number
// of accesses of current node.
func (node NodeLog) PriorPtoB(nodeB uint64) float64 {
	if node.TotalAccesses == 0 {
		return 0
	}

	return float64(node.ToB[nodeB]) / float64(node.TotalAccesses)
}

// Restore restores the records of the logger in gob binary format from pathFile.
func (node *NodeLog) Restore(dt dumptype.DumpType, r io.Reader) error {
	registGob()

	if r == nil {
		return errors.New("restore failed: the io.Reader r is nil")
	}

	node.ClassMap = &Map{}

	switch dt {
	case dumptype.JSON:
		return node.restoreJSON(r)
	case dumptype.GOB:
		return node.restoreGob(r)
	case dumptype.CSV:
		// wip
		fallthrough
	case dumptype.SQL:
		// wip
		fallthrough
	case dumptype.Unknown:
		fallthrough
	default:
		return errors.New("restore failed: unsupported dump type")
	}
}

// restoreJSON restores the records of the logger from JSON formatted data.
func (node *NodeLog) restoreJSON(r io.Reader) error {
	dec := json.NewDecoder(r)

	return errors.Wrap(dec.Decode(&node), "restore failed: json.Decode")
}

// restoreGob restores the records of the logger from gob binary formatted data.
func (node *NodeLog) restoreGob(r io.Reader) error {
	dec := gob.NewDecoder(r)

	return errors.Wrap(dec.Decode(&node), "restore failed: gob.Decode")
}

// Store stores the records of the logger in gob binary format to pathFile.
func (node NodeLog) Store(dt dumptype.DumpType, classMap class.Map, w io.Writer) error {
	registGob()

	if w == nil {
		return errors.New("dump failed: the io.Writer w is nil")
	}

	node.ClassMap = classMap

	switch dt {
	case dumptype.JSON:
		return node.storeJSON(w)
	case dumptype.GOB:
		return node.storeGob(w)
	case dumptype.CSV:
		// wip
		fallthrough
	case dumptype.SQL:
		// wip
		fallthrough
	case dumptype.Unknown:
		fallthrough
	default:
		return errors.New("dump failed: unsupported dump type")
	}
}

// storeJSON stores the records of the logger to w in gob binary format.
func (node NodeLog) storeJSON(w io.Writer) error {
	byteJSON, err := jsonMarshalIndent(node, marshalIndentPrefix, marshalIndentIndent)
	if err != nil {
		return errors.Wrap(err, "dump failed: can not marshal NodeLog to JSON")
	}

	if _, err = w.Write(byteJSON); err != nil {
		return errors.Wrap(err, "dump failed: can not write NodeLog to JSON")
	}

	return nil
}

// storeGob stores the records of the logger to w in gob binary format.
func (node NodeLog) storeGob(w io.Writer) error {
	if err := gob.NewEncoder(w).Encode(node); err != nil {
		return errors.Wrap(err, "dump failed: can not encode to gob")
	}

	return nil
}

// String returns a string representation of the NodeLog which is the node ID.
func (node NodeLog) String() string {
	return fmt.Sprintf("%d", node.NodeID)
}

// Update updates the records of a node.
// It must be called by the next node accessed.
func (node *NodeLog) Update(fromA, toB uint64) {
	if _, ok := node.FromAToB[fromA]; !ok {
		node.FromAToB[fromA] = make(map[uint64]int)
	}

	node.TotalAccesses++
	node.FromA[fromA]++
	node.ToB[toB]++
	node.FromAToB[fromA][toB]++
}
