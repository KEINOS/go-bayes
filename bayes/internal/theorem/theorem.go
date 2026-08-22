/*
Package theorem is an Bayes' theorem implementation.
*/
package theorem

// Bayes returns the posterior probability by Bayes' theorem.
// It estimates how likely event B is when event A has already happened.
//
//   - priorPtoB is the base chance of event B before looking at event A.
//   - priorPfromAtoB is the chance that A is observed when B is true.
//   - priorPNotFromAtoB is the chance that A is observed when B is not true.
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
