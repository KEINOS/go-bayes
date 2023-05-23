package theorem_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/pkg/theorem"
)

func ExampleBayes() {
	// Validation of figures
	// Bayes' theorem
	x := 0.3 * 0.4
	y := x + (1.-0.3)*0.5
	expect := x / y //nolint: ifshort // not using short syntax for readability

	// Prior probability to be B.
	PriorProbToB := 0.3
	// Prior probability to be B if the previous was A.
	PriorProbFromAtoB := 0.4
	// Prior probability not to be B if the previous was A.
	PriorProbNotFromAtoB := 0.5

	actual := theorem.Bayes(PriorProbToB, PriorProbFromAtoB, PriorProbNotFromAtoB)

	if expect == actual {
		fmt.Println("OK")
	}

	// Output: OK
}
