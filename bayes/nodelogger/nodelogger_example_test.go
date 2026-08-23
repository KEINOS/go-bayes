package nodelogger_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/internal/nodeloggers/logmem"
	"github.com/KEINOS/go-bayes/bayes/nodelogger"
)

// ExampleNodeLogger demonstrates how to use the [NodeLogger] interface
// with the in-memory logmem implementation.
func ExampleNodeLogger() {
	// logmem.New() returns a *NodeLog that satisfies the NodeLogger interface.
	var logger nodelogger.NodeLogger = logmem.New(42)

	// Record transitions: 1->2 twice and 3->2 once.
	logger.Update(1, 2)
	logger.Update(1, 2)
	logger.Update(3, 2)

	fmt.Println(logger.ID())
	fmt.Printf("%.2f\n", logger.PriorProbTo(2))
	fmt.Printf("%.2f\n", logger.PriorProbFromTo(1, 2))
	fmt.Printf("%.2f\n", logger.Predict(1, 2))
	//
	// Output:
	// 42
	// 1.00
	// 0.67
	// 1.00
}
