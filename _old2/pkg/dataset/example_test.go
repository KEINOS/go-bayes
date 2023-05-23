package dataset_test

import (
	"fmt"
	"log"

	"github.com/KEINOS/go-bayes/pkg/dataset"
	"github.com/KEINOS/go-bayes/pkg/kvs"
)

func Example() {
	exitOnError := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	corpus := []any{
		"Do", "Re", "Mi", "Fa", "Sol", "La", "Ti", "Do",
	}

	db, err := dataset.New(kvs.InMemoryDB)
	exitOnError(err)

	// Bulk import to a DB.
	// The returned DB is a set of class IDs and their original values.
	class, err := db.AddFromSlice(corpus)
	exitOnError(err)

	// Iterate while db has a next item.
	for db.HasNext() {
		// Get the next item. The returned value is the Class ID of the item.
		key, value, err := db.Next()
		exitOnError(err)

		// Get the original value from the class DB.
		originalValue, err := class.Get(value)
		exitOnError(err)

		fmt.Printf("Key: %v, Value (ClassID): %v (-> Orivinal value: %v)\n", key, value, originalValue)
	}

	fmt.Println("Keys in DB:", db.GetKeys())
	fmt.Println("Keys in Class DB:", class.GetKeys())

	// Output:
	// Key: 0, Value (ClassID): 15698525381150397392 (-> Orivinal value: Do)
	// Key: 1, Value (ClassID): 12238976554927056022 (-> Orivinal value: Re)
	// Key: 2, Value (ClassID): 2918913830112925382 (-> Orivinal value: Mi)
	// Key: 3, Value (ClassID): 1955428390596480609 (-> Orivinal value: Fa)
	// Key: 4, Value (ClassID): 9861478780717170130 (-> Orivinal value: Sol)
	// Key: 5, Value (ClassID): 5302957724057310290 (-> Orivinal value: La)
	// Key: 6, Value (ClassID): 1572347194862695639 (-> Orivinal value: Ti)
	// Key: 7, Value (ClassID): 15698525381150397392 (-> Orivinal value: Do)
	// Keys in DB: [0 1 2 3 4 5 6 7]
	// Keys in Class DB: [15698525381150397392 12238976554927056022 2918913830112925382 1955428390596480609 9861478780717170130 5302957724057310290 1572347194862695639]
}
