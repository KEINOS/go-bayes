package kvs_test

import (
	"fmt"
	"log"

	"github.com/KEINOS/go-bayes/pkg/kvs"
)

func Example() {
	exitOnError := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	// DB type
	dbType := kvs.InMemoryDB

	// Instantiate a KVS
	db := kvs.New(dbType)
	if db == nil {
		log.Fatal("failed to instantiate a KVS db")
	}

	// Put a key-value pair
	exitOnError(db.Set(1, "foo"))
	exitOnError(db.Set(2, "bar"))
	exitOnError(db.Set(3, "baz"))

	// Get the values from the key and print them
	v, err := db.GetValue(1)
	fmt.Println(v, err)
	fmt.Println(db.GetValue(2))
	fmt.Println(db.GetValue(3))

	// Unregistered key will return an error
	fmt.Println(db.GetValue(4))

	// Output:
	// foo <nil>
	// bar <nil>
	// baz <nil>
	// <nil> key 4 not found
}

func Example_regist_original_struct() {
	exitOnError := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	// DB type
	dbType := kvs.InMemoryDB

	// Instantiate a KVS
	db := kvs.New(dbType)
	if db == nil {
		log.Fatal("failed to instantiate a KVS db")
	}

	// Your original struct
	type MyStruct struct {
		Name string
		Age  int
	}

	// Regist your original struct to gob registry.
	// This is important to serialize your struct during saving the data to a file.
	db.RegistGob(*new(MyStruct))

	// Put a key-value pair
	exitOnError(db.Set(1, MyStruct{Name: "foo", Age: 10}))
	exitOnError(db.Set(2, MyStruct{Name: "bar", Age: 20}))
	exitOnError(db.Set(3, MyStruct{Name: "baz", Age: 30}))

	// Get the values from the key and print them
	for _, key := range db.GetKeys() {
		anyValue, err := db.GetValue(key)
		if err != nil {
			log.Fatal(err)
		}

		v, ok := anyValue.(MyStruct) // type assertion
		if !ok {
			log.Fatal("failed to type assertion")
		}

		fmt.Println(key, v.Name, v.Age)
	}

	// Output:
	// 1 foo 10
	// 2 bar 20
	// 3 baz 30
}
