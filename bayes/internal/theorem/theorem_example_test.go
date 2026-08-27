package theorem_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/internal/theorem"
)

func ExampleBayes() {
	prior := 0.3
	observedWithClass := 0.4
	observedWithoutClass := 0.5

	probability := theorem.Bayes(prior, observedWithClass, observedWithoutClass)

	fmt.Printf("%.3f\n", probability)
	// Output: 0.255
}
