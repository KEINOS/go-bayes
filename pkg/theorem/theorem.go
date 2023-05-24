/*
Package theorem is an Bayes' theorem implementation.
*/
package theorem

// Bayes returns the probability from the given probability based on Bayes' theorem.
// In other words, it returns the probability of event B occurring when event A occurred.
//
//   - priorPtoB is the prior probability occurrence of event B.
//   - priorPfromAtoB is the prior probability occurrence of event B when previous event A occurred.
//   - priorPNotFromAtoB is the prior probability that did not occur event B when previous event A occurred.
//
//nolint:varnamelen // short names are more readable in this case
func Bayes(priorPtoB, priorPfromAtoB, priorPNotFromAtoB float64) float64 {
	zero := float64(0)
	one := float64(1)

	// Bayes' theorem
	x := priorPtoB * priorPfromAtoB
	y := x + (one-priorPtoB)*priorPNotFromAtoB

	// Avoid zero division
	if y == zero {
		return zero
	}

	return x / y
}
