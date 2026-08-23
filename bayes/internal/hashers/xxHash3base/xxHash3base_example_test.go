package xxhash3base_test

import (
	"fmt"
	"log"

	"github.com/KEINOS/go-bayes/bayes/internal/hashers/xxHash3base"
)

func ExampleHasher_HashTrans() {
	hasher := xxhash3base.New()

	flowID, err := hasher.HashTrans(10, 11, 12, 13, 14, 15)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(flowID)
	// Output: 7177008186327462541
}
