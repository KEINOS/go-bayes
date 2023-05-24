//nolint:varnamelen // short names are more readable in this case
package logmem_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/pkg/nodelogger/logmem"
)

func ExampleNodeLog_ID() {
	n := logmem.New(12345)

	fmt.Println(n.ID())

	// Output: 12345
}

func ExampleNodeLog_Predict() {
	const (
		x = uint64(1) // Node ID of node x
		y = uint64(2) // Node ID of node y
		z = uint64(3) // Node ID of node z
	)

	nodeY := logmem.New(y) // Create a new node y

	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.
	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.

	fmt.Println("Prediction of outgoing node to be x, on incoming node as x:", nodeY.Predict(x, x))
	fmt.Println("Prediction of outgoing node to be y, on incoming node as x:", nodeY.Predict(x, y))
	fmt.Println("Prediction of outgoing node to be z, on incoming node as x:", nodeY.Predict(x, z))

	// Output:
	// Prediction of outgoing node to be x, on incoming node as x: 0
	// Prediction of outgoing node to be y, on incoming node as x: 0
	// Prediction of outgoing node to be z, on incoming node as x: 1
}

func ExampleNodeLog_PriorPfromAtoB() {
	const (
		x = uint64(1) // Node ID of node x
		y = uint64(2) // Node ID of node y
		z = uint64(3) // Node ID of node z
	)

	nodeY := logmem.New(y) // Create a new node y

	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.
	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.

	fmt.Println("Prior probability of outgoing node x, on incoming node as x (x -> y -> x):",
		nodeY.PriorPfromAtoB(x, x))
	fmt.Println("Prior probability of outgoing node y, on incoming node as x (x -> y -> y):",
		nodeY.PriorPfromAtoB(x, y))
	fmt.Println("Prior probability of outgoing node z, on incoming node as x (x -> y -> z):",
		nodeY.PriorPfromAtoB(x, z))

	// Output:
	// Prior probability of outgoing node x, on incoming node as x (x -> y -> x): 0
	// Prior probability of outgoing node y, on incoming node as x (x -> y -> y): 0
	// Prior probability of outgoing node z, on incoming node as x (x -> y -> z): 0.5
}

func ExampleNodeLog_PriorPNotFromAtoB() {
	const (
		x = uint64(1) // Node ID of node x
		y = uint64(2) // Node ID of node y
		z = uint64(3) // Node ID of node z
	)

	nodeY := logmem.New(y) // Create a new node y

	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.
	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.

	fmt.Println("Prior probability of outgoing node is not x, on incoming node as x (x -> y -> not x):",
		nodeY.PriorPNotFromAtoB(x, x))
	fmt.Println("Prior probability of outgoing node is not y, on incoming node as x (x -> y -> not y):",
		nodeY.PriorPNotFromAtoB(x, y))
	fmt.Println("Prior probability of outgoing node is not z, on incoming node as x (x -> y -> not z):",
		nodeY.PriorPNotFromAtoB(x, z))

	// Output:
	// Prior probability of outgoing node is not x, on incoming node as x (x -> y -> not x): 0.5
	// Prior probability of outgoing node is not y, on incoming node as x (x -> y -> not y): 0.5
	// Prior probability of outgoing node is not z, on incoming node as x (x -> y -> not z): 0
}

func ExampleNodeLog_PriorPtoB() {
	const (
		x = uint64(1)
		y = uint64(2)
		z = uint64(3)
	)

	nodeY := logmem.New(y)

	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.
	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.

	fmt.Println("Prior probability of outgoing node x (y -> x):", nodeY.PriorPtoB(x))
	fmt.Println("Prior probability of outgoing node y (y -> y):", nodeY.PriorPtoB(y))
	fmt.Println("Prior probability of outgoing node z (y -> z):", nodeY.PriorPtoB(z))

	// Output:
	// Prior probability of outgoing node x (y -> x): 0.5
	// Prior probability of outgoing node y (y -> y): 0
	// Prior probability of outgoing node z (y -> z): 0.5
}

func ExampleNodeLog_String() {
	n := logmem.New(12345)

	// Stringer implementation for the NodeLog type.
	fmt.Println(n)

	// Output: 12345
}

func ExampleNodeLog_Update() {
	const (
		x = uint64(1)
		y = uint64(2)
		z = uint64(3)
	)

	nodeY := logmem.New(y)

	// Update the access log of nodeY.
	//
	// It will update the access log as "x -> y -> z".
	// Which means that node x is the predecessor of node y, and node z is the
	// successor of node y. In other words, node x is the incoming node and node
	// z is the outgoing node of node y.
	//
	// Note that it must be called by the next node accessed. In this case,
	// node z should call this function.
	nodeY.Update(x, z)

	fmt.Println("Total access:", nodeY.TotalAccesses)
	fmt.Println("Number of access from node x:", nodeY.FromA[x])
	fmt.Println("Number of access from node z:", nodeY.FromA[z])
	fmt.Println("Number of outgoing node x:", nodeY.ToB[x])
	fmt.Println("Number of outgoing node z:", nodeY.ToB[z])

	// Output:
	// Total access: 1
	// Number of access from node x: 1
	// Number of access from node z: 0
	// Number of outgoing node x: 0
	// Number of outgoing node z: 1
}
