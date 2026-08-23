package main

import "os"

func Example() {
	err := run(os.Stdout)
	if err != nil {
		panic(err)
	}

	// Output:
	// So
}
