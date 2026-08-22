package logmem

import "maps"

// ============================================================================
//  Type: Snapshot
// ============================================================================

// Snapshot is a serializable copy of NodeLog state.
type Snapshot struct {
	NodeID        uint64                    `json:"nodeId"`
	TotalAccesses int                       `json:"totalAccesses"`
	FromA         map[uint64]int            `json:"fromA"`
	ToB           map[uint64]int            `json:"toB"`
	FromAToB      map[uint64]map[uint64]int `json:"fromAToB"`
}

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

// NewFromSnapshot restores a NodeLog from a snapshot.
func NewFromSnapshot(snapshot Snapshot) *NodeLog {
	return &NodeLog{
		nodeID:        snapshot.NodeID,
		totalAccesses: snapshot.TotalAccesses,
		fromA:         cloneIntMap(snapshot.FromA),
		toB:           cloneIntMap(snapshot.ToB),
		fromAToB:      cloneNestedIntMap(snapshot.FromAToB),
	}
}

// ----------------------------------------------------------------------------
//  Methods
// ----------------------------------------------------------------------------

// Snapshot returns a deep-copied snapshot of current state.
func (n NodeLog) Snapshot() Snapshot {
	return Snapshot{
		NodeID:        n.nodeID,
		TotalAccesses: n.totalAccesses,
		FromA:         cloneIntMap(n.fromA),
		ToB:           cloneIntMap(n.toB),
		FromAToB:      cloneNestedIntMap(n.fromAToB),
	}
}

// ----------------------------------------------------------------------------
//  Helper Functions
// ----------------------------------------------------------------------------

func cloneIntMap(src map[uint64]int) map[uint64]int {
	out := make(map[uint64]int, len(src))

	maps.Copy(out, src)

	return out
}

func cloneNestedIntMap(src map[uint64]map[uint64]int) map[uint64]map[uint64]int {
	out := make(map[uint64]map[uint64]int, len(src))

	for outerKey, inner := range src {
		out[outerKey] = cloneIntMap(inner)
	}

	return out
}
