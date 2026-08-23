package main

import "os"

func Example() {
	err := run(os.Stdout)
	if err != nil {
		panic(err)
	}

	// Output:
	// trained: 8124 samples
	// cap=x, odor=p, stalk-root=e, habitat=u -> poisonous
	// cap=x, odor=a, stalk-root=c, habitat=g -> edible
	// cap=x, odor=n, stalk-root=?, habitat=w -> edible
}
