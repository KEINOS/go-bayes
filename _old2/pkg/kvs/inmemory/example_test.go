package inmemory_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Code-Hex/dd"
	"github.com/KEINOS/go-bayes/pkg/kvs/inmemory"
	"github.com/KEINOS/go-bayes/pkg/utils"
)

// ----------------------------------------------------------------------------
//  Helper functions for examples
// ----------------------------------------------------------------------------

func getPathDirTemp() (string, func()) {
	tempDir := os.TempDir()
	hashedPath := utils.BytesToHash([]byte(tempDir))
	nameDir := fmt.Sprintf("%x", hashedPath)
	pathDirTemp := filepath.Join(tempDir, nameDir)

	if err := os.MkdirAll(pathDirTemp, 0o755); err != nil {
		log.Fatal(err)
	}

	return pathDirTemp, func() {
		if err := os.RemoveAll(pathDirTemp); err != nil {
			log.Fatal(err)
		}
	}
}

// ----------------------------------------------------------------------------
//  Examples
// ----------------------------------------------------------------------------

//nolint: funlen,cyclop // ignore the length and complexity for this example
func Example() {
	// Test data to store.
	data := []any{
		"Hello, world!",    // string
		1234,               // int
		true,               // bool
		[]byte("foo bar!"), // []byte
	}

	// Instantiate a new in-memory key-value store.
	db := inmemory.New()

	// Store the data.
	for _, value := range data {
		// Generate a unique uint64 key for the value.
		key, err := utils.AnyToStateID(value)
		if err != nil {
			log.Fatal(err)
		}

		// Set method sets the value for the given key.
		if err := db.Set(key, value); err != nil {
			log.Fatal(err)
		}
	}

	// Keys method returns the list of keys in the KVS.
	keys := db.GetKeys()

	// GetValue method returns the value for the given key.
	for _, key := range keys {
		v, err := db.GetValue(key)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Get value:", v)
	}

	// Delete method deletes the value for the given key.
	// 6600071756362922836 is the key of second element int(1234).
	if err := db.Delete(6600071756362922836); err != nil {
		log.Fatal(err)
	}

	// --------------------------------------------------
	//  Iteration
	// --------------------------------------------------
	//  inmemory.DB implements iterator.Iterator interface.

	// Reset method resets the iterator to the first element.
	db.Reset()

	// HasNext method returns true if there is a next element.
	for db.HasNext() {
		// Next method returns the next key-value pair.
		key, v, err := db.Next()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Key: %v, Value: %v (Type: %T)\n", key, v, v)
	}

	// --------------------------------------------------
	//  Storing and restoring data to a file
	// --------------------------------------------------

	// Get directory path to store the data.
	// `getPathDirTemp` function is a dummy function for the example which
	// returns a temporary directory path.
	pathDirTmp, cleanup := getPathDirTemp()
	defer cleanup()

	// File path to save the data.
	filePath := filepath.Join(pathDirTmp, "sample.db")

	// Save stores the data to a file.
	if err := db.Save(filePath); err != nil {
		log.Fatal(err)
	}

	// Create another DB
	db2 := inmemory.New()

	// Load restores the data from a file to the object.
	if err := db2.Load(filePath); err != nil {
		log.Fatal(err)
	}

	// Dump the structure after load. `dd.Dump()` is a pretty-printer for the
	// structure.
	fmt.Println("DB After load:", dd.Dump(db2))

	// Output:
	// Get value: Hello, world!
	// Get value: 1234
	// Get value: true
	// Get value: [102 111 111 32 98 97 114 33]
	// Key: 15100329949952889093, Value: Hello, world! (Type: string)
	// Key: 15430900891160433076, Value: true (Type: bool)
	// Key: 10440356249599916122, Value: [102 111 111 32 98 97 114 33] (Type: []uint8)
	// DB After load: &inmemory.DB{
	//   Data: map[uint64]interface {}{
	//     10440356249599916122: []uint8{
	//       102,
	//       111,
	//       111,
	//       32,
	//       98,
	//       97,
	//       114,
	//       33,
	//     },
	//     15100329949952889093: "Hello, world!",
	//     15430900891160433076: true,
	//   },
	//   Keys: []uint64{
	//     15100329949952889093,
	//     15430900891160433076,
	//     10440356249599916122,
	//   },
	//   Index: 0,
	// }
}
