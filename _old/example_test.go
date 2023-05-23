package bayes_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/KEINOS/go-bayes"
)

// ============================================================================
//  Helper Functions for Example
// ============================================================================

func getTempDir() (pathDir string, cleanUp func()) {
	pathDir, err := os.MkdirTemp(os.TempDir(), "go-bayes-test-*")
	if err != nil {
		log.Fatal(err)
	}

	return pathDir, func() {
		os.RemoveAll(pathDir)
	}
}

// ============================================================================
//  Examples
// ============================================================================

func Example() {
	// Train data: "Happy Birthday"
	// This train data is a slice of strings but it can be any other type of slice.
	// Currently supported types are as follows:
	//   bool, int, int16-int64, uint, uint16-uint64, float32, float64, string.
	score := []string{
		"So", "So", "La", "So", "Do", "Si",
		"So", "So", "La", "So", "Re", "Do",
		"So", "So", "So", "Mi", "Do", "Si", "La",
		"Fa", "Fa", "Mi", "Do", "Re", "Do",
	}

	// Reset the trained model
	bayes.Reset()

	// Train
	if err := bayes.Train(score); err != nil {
		log.Fatal(err)
	}

	// Predict the next note from the introduction notes
	for _, intro := range [][]string{
		{"So", "So", "La", "So", "Do", "Si"},             // --> So
		{"So", "So", "La", "So", "Do", "Si", "So", "So"}, // --> La
		{"So", "So", "La"},                               // --> So
		{"So", "So", "So"},                               // --> Mi
	} {
		// Predict the next note from the intro
		nextNoteID, err := bayes.Predict(intro)
		if err != nil {
			log.Fatal(err)
		}

		// Print the predicted next note
		nextNoteString := bayes.GetClass(nextNoteID)
		fmt.Printf("Next is: %v (Class ID: %v)\n", nextNoteString, nextNoteID)
	}

	// Store the trained model/data to a file
	pathDir, cleanUp := getTempDir()
	defer cleanUp() // clean up the temporary directory after the test

	pathFile := filepath.Join(pathDir, "example.model")

	err := bayes.Store(pathFile)
	if err != nil {
		log.Fatal(err)
	}

	// Reset the trained model again before loading the model from the file
	bayes.Reset()

	// Load the trained model from the file
	err = bayes.Restore(pathFile)
	if err != nil {
		log.Fatal(err)
	}

	// Predict the next note from the introduction notes
	nextNoteID, err := bayes.Predict(
		[]string{"So", "So", "La", "So", "Do", "Si", "So", "So"},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Print the predicted next note
	fmt.Printf("Next is: %v (Class ID: %v)\n", bayes.GetClass(nextNoteID), nextNoteID)

	// Output:
	// Next is: So (Class ID: 10062876669317908741)
	// Next is: La (Class ID: 17627200281938459623)
	// Next is: So (Class ID: 10062876669317908741)
	// Next is: Mi (Class ID: 6586414841969023711)
	// Next is: La (Class ID: 17627200281938459623)
}

// ----------------------------------------------------------------------------
//  HashTrans()
// ----------------------------------------------------------------------------

func ExampleHashTrans() {
	// list of transition IDs. If the order or the value of the list is changed,
	// the hash will be changed.
	for _, transitions := range [][]uint64{
		{10, 11, 12, 13, 14, 15},
		{10, 11, 12, 13, 15, 14},
		{1, 11, 12, 13, 14, 15},
		{1},
	} {
		hashed, err := bayes.HashTrans(transitions...)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Dec: %020d\n", hashed)
		fmt.Printf("Hex: %016x\n", hashed)
	}

	// Output:
	// Dec: 07573192273568316974
	// Hex: 6919623f91c5be2e
	// Dec: 07941539160827123980
	// Hex: 6e3603ca6af4590c
	// Dec: 16813156106886104905
	// Hex: e95454663822c749
	// Dec: 01877176418821510543
	// Hex: 1a0d1201d898958f
}

// ----------------------------------------------------------------------------
//  New()
// ----------------------------------------------------------------------------

func ExampleNew() {
	// Scope ID is used to distinguish the stored data.
	scopeID := uint64(100)

	// Create a new bayes instance with in-memory storage.
	trainer, err := bayes.New(bayes.MemoryStorage, scopeID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(trainer.ID())

	// Output: 100
}

// ----------------------------------------------------------------------------
//  Reset(), Train() and Predict()
// ----------------------------------------------------------------------------

func ExampleTrain_bool() {
	defer bayes.Reset()

	// "Save Our Souls" Morse code
	codes := []bool{
		true, true, true, // ... ==> S
		false, false, false, // ___ ==> O
		true, true, true, // ... ==> S
	}

	// Train
	if err := bayes.Train(codes); err != nil {
		log.Fatal(err)
	}

	// Quiz
	quiz := []bool{
		true, true, true,
		false, false, false,
		true, true, // --> expect next to be true
	}

	// Predict the next code
	nextCode, err := bayes.Predict(quiz)
	if err != nil {
		log.Fatal(err)
	}

	if bayes.GetClass(nextCode).(bool) {
		fmt.Println("OK")
	}

	// Output: OK
}

//nolint: funlen // This function is a little long and complex but leave it as is.
func ExampleTrain_int() {
	defer bayes.Reset()

	const (
		Do int = iota
		Re
		Mi
		Fa
		So
		La
		Si
	)

	// Happy Birthday
	score := []int{
		So, So, La, So, Do, Si,
		So, So, La, So, Re, Do,
		So, So, So, Mi, Do, Si, La,
		Fa, Fa, Mi, Do, Re, Do,
	}

	// Train
	if err := bayes.Train(score); err != nil {
		log.Fatal(err)
	}

	// Convert int to string that represents the note
	getNote := func(noteID int) string {
		switch noteID {
		case Do:
			return "Do"
		case Re:
			return "Re"
		case Mi:
			return "Mi"
		case Fa:
			return "Fa"
		case So:
			return "So"
		case La:
			return "La"
		case Si:
			return "Si"
		}

		return "Unknown"
	}

	// Predict the next note
	for _, notes := range [][]int{
		{So, So, La, So, Do, Si},         // --> So
		{So, So, La, So, Do, Si, So, So}, // --> La
		{So, So, La},                     // --> So
		{So, So, So},                     // --> Mi
	} {
		nextNote, err := bayes.Predict(notes)
		if err != nil {
			log.Fatal(err)
		}

		// Print the next note
		noteID, ok := bayes.GetClass(nextNote).(int)
		if !ok {
			log.Fatal("Invalid class type")
		}

		fmt.Printf("Class: %v (ID: %v)\n", getNote(noteID), nextNote)
	}

	// Output:
	// Class: So (ID: 4)
	// Class: La (ID: 5)
	// Class: So (ID: 4)
	// Class: Mi (ID: 2)
}

func ExampleTrain_string() {
	defer bayes.Reset()

	// Happy Birthday
	score := []string{
		"So", "So", "La", "So", "Do", "Si",
		"So", "So", "La", "So", "Re", "Do",
		"So", "So", "So", "Mi", "Do", "Si", "La",
		"Fa", "Fa", "Mi", "Do", "Re", "Do",
	}

	// Train
	if err := bayes.Train(score); err != nil {
		log.Fatal(err)
	}

	// Predict the next note
	for _, notes := range [][]string{
		{"So", "So", "La", "So", "Do", "Si"},             // --> So
		{"So", "So", "La", "So", "Do", "Si", "So", "So"}, // --> La
		{"So", "So", "La"},                               // --> So
		{"So", "So", "So"},                               // --> Mi
	} {
		nextNote, err := bayes.Predict(notes)
		if err != nil {
			log.Fatal(err)
		}

		// Print the next note
		nextNoteString := bayes.GetClass(nextNote)

		fmt.Printf("Class: %v (ID: %v)\n", nextNoteString, nextNote)
	}

	// Output:
	// Class: So (ID: 10062876669317908741)
	// Class: La (ID: 17627200281938459623)
	// Class: So (ID: 10062876669317908741)
	// Class: Mi (ID: 6586414841969023711)
}

// ============================================================================
//  Type: Storage
// ============================================================================
// ----------------------------------------------------------------------------
//  Storage.IsUnknown()
// ----------------------------------------------------------------------------

func ExampleStorage_IsUnknown() {
	if bayes.UnknwonStorage.IsUnknown() {
		fmt.Println("Unknown")
	}

	if bayes.Storage(int(999999999)).IsUnknown() {
		fmt.Println("Unknown")
	}

	// Output:
	// Unknown
	// Unknown
}

// ----------------------------------------------------------------------------
//  Storage.String()
// ----------------------------------------------------------------------------

func ExampleStorage_String() {
	// bayes.Storage type implements the Stringer interface.
	fmt.Println(bayes.MemoryStorage)
	fmt.Println(bayes.SQLite3Storage)
	fmt.Println(bayes.UnknwonStorage)

	// Output:
	// in-memory
	// SQLite3
	// unknown
}

// ----------------------------------------------------------------------------
//  Storage.Type()
// ----------------------------------------------------------------------------

func ExampleStorage_Type() {
	fmt.Println(bayes.MemoryStorage.Type())
	fmt.Println(bayes.SQLite3Storage.Type())
	fmt.Println(bayes.UnknwonStorage.Type())

	// Output:
	// in-memory
	// SQLite3
	// unknown
}
