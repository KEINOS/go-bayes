/*
Package theorem is an Bayes' theorem implementation.
*/
package theorem

// Bayes returns the probability from the given probability based on Bayes' theorem.
func Bayes(PriorPtoB, PriorPfromAtoB, PriorPNotFromAtoB float64) float64 {
	// Bayes' theorem
	x := PriorPtoB * PriorPfromAtoB
	y := x + (float64(1)-PriorPtoB)*PriorPNotFromAtoB

	if y == float64(0) {
		return float64(0)
	}

	return x / y
}
