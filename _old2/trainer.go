package bayes

import (
	"github.com/KEINOS/go-bayes/pkg/theorem"
	"github.com/KEINOS/go-bayes/pkg/utils"
)

// ----------------------------------------------------------------------------
//  Type: Trainer
// ----------------------------------------------------------------------------

// Trainer holds the records of transitions.
type Trainer struct {
	// FromAtoB is the number of accesses from node A to node B as map[A]map[B].
	// A is the incoming access and B is the outgoing access.
	FromAToB map[uint64]map[uint64]int `json:"from_a_to_b"`
	// FromA is the number of incoming accesses from node A as map[A].
	FromA map[uint64]int `json:"from_a"`
	// ToB is the number of outgoing accesses to node B as map[B].
	ToB map[uint64]int `json:"to_b"`
	// TotalAccesses is the total number of accesses to the node.
	TotalAccesses int `json:"total_accesses"`
}

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

func NewTrainer() *Trainer {
	return &Trainer{
		FromAToB: make(map[uint64]map[uint64]int),
		FromA:    make(map[uint64]int),
		ToB:      make(map[uint64]int),
	}
}

// ----------------------------------------------------------------------------
//  Methods
// ----------------------------------------------------------------------------

// Predict returns the class predicted by the given input.
func (t Trainer) Predict(label map[uint64]any, input []uint64) map[uint64]any {
	weight := float64(0)
	currClass := map[uint64]any{}

	fromA := utils.IntsToStateID(input)

	for toB, v := range label {
		priorPtoB := t.PriorPtoB(toB)
		priorPfromAtoB := t.PriorPfromAtoB(fromA, toB)
		priorPNotFromAtoB := t.PriorPNotFromAtoB(fromA, toB)

		prov := theorem.Bayes(priorPtoB, priorPfromAtoB, priorPNotFromAtoB)
		if prov > weight {
			weight = prov
			currClass = map[uint64]any{toB: v}
		}
	}

	return currClass
}

func (t Trainer) PriorPtoB(toB uint64) float64 {
	return float64(t.ToB[toB]) / float64(t.TotalAccesses)
}

func (t Trainer) PriorPfromAtoB(fromA uint64, toB uint64) float64 {
	return float64(t.FromAToB[fromA][toB]) / float64(t.FromA[fromA])
}

func (t Trainer) PriorPNotFromAtoB(fromA uint64, toB uint64) float64 {
	// If there is no access, never reaches to node B. Or, if there is no
	// access from node A, never reaches to node B as well.
	// Therefore, the probability of not reaching node B is 100%.
	if t.TotalAccesses == 0 || t.FromA[fromA] == 0 {
		return 1
	}

	notFromAtoB := t.FromA[fromA] - t.FromAToB[fromA][toB]

	return float64(notFromAtoB) / float64(t.TotalAccesses)
}

func (t *Trainer) Update(fromA, toB uint64) {
	if _, ok := t.FromAToB[fromA]; !ok {
		t.FromAToB[fromA] = make(map[uint64]int)
	}

	t.FromAToB[fromA][toB]++
	t.FromA[fromA]++
	t.ToB[toB]++
	t.TotalAccesses++
}

func (t *Trainer) trainRow(columns []uint64) {
	postID := uint64(0)

	for i, column := range columns {
		if i == 0 {
			postID = column

			continue
		}

		t.Update(postID, column)

		// Regression
		for ii := 1; ii < i; ii++ {
			nToLast := i - ii - 1 // Nth to last
			drill := columns[nToLast:i]

			fromA := utils.IntsToStateID(drill)
			toB := column

			t.Update(fromA, toB)
		}

		postID = column
	}
}

func (t *Trainer) Train(s *Sample) error {
	for _, row := range s.Data {
		t.trainRow(row)
	}

	return nil
}
