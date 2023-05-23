package theorem

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBayes_zero_division(t *testing.T) {
	v := Bayes(0., 0., 0.)

	require.Zero(t, v, "divide by zero should return zero")
}

// Fuzzing test.
// Fom the root of the project, run:
//   go test -fuzz=FuzzBayes ./pkg/theorem
func FuzzBayes(f *testing.F) {
	f.Add(float64(0.3), float64(0.4), float64(0.5))
	f.Add(float64(0.), float64(0.), float64(0.))
	f.Add(float64(0.), float64(0.4), float64(0.5))
	f.Add(float64(0.3), float64(0.), float64(0.5))
	f.Add(float64(0.3), float64(0.4), float64(0.))

	f.Fuzz(func(t *testing.T, a, b, c float64) {
		require.NotPanics(t, func() {
			_ = Bayes(a, b, c)
		})
	})
}
