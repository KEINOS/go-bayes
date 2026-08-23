package main

import "os"

func Example() {
	err := run(os.Stdout)
	if err != nil {
		panic(err)
	}

	// Output:
	// trained: 178 samples
	// alcohol=14.23, color=5.64, proline=1065 -> cultivar 1
	// alcohol=12.37, color=1.95, proline=520 -> cultivar 2
	// alcohol=12.86, color=4.1, proline=630 -> cultivar 3
}
