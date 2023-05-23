package bayes_test

import (
	"fmt"
	"log"

	"github.com/Code-Hex/dd"
	"github.com/KEINOS/go-bayes"
)

func Example() {
	// ------------------------------------------------------------------------
	//  Prepare the data for training
	// ------------------------------------------------------------------------
	// Sample data to train. In this case, each row is one sample.
	sampleData := [][]string{
		{"do", "re", "mi", "fa", "so", "la", "si", "do"},
		{"do", "do", "re", "do", "so", "mi"},
	}

	// Convert raw data to Sample object.
	sample, err := bayes.LoadSample(bayes.SliceString, sampleData)
	if err != nil {
		log.Fatal(err)
	}

	// ------------------------------------------------------------------------
	// Train the model
	// ------------------------------------------------------------------------
	trainer := bayes.NewTrainer()

	trainer.Train(sample)

	// ------------------------------------------------------------------------
	// Prediction
	// ------------------------------------------------------------------------
	testData := []string{
		"do", "re", "mi", "fa", "so", "la", "si", // expected: "do"
	}

	// Convert slice of string to slice of uint64.
	testVector, err := bayes.Vectorize(testData)
	if err != nil {
		log.Fatal(err)
	}

	// Predict the most matching class in the sample.Class from the given testData.
	predictedClass := trainer.Predict(sample.Class, testVector)

	// Get the expected ID of "do" from the class map.
	expectID, err := sample.GetID("do")
	if err != nil {
		log.Fatal(err)
	}

	// Get value from class map.
	fmt.Println(predictedClass[expectID])

	// Output: do
}

func ExampleToClass() {
	// Sample data.
	sample := []any{
		"do", "re", "mi", "fa", "so", "la", "si", "do",
	}

	// ToClass function returns a bayas.Class object from the given sample data.
	// The object is the map of Scalar (map[bayes.Scalar]bayes.Label).
	class, err := bayes.ToClass(sample)
	if err != nil {
		log.Fatal(err)
	}

	// Get list of scalar values in the class map.
	scalars := class.Keys()

	for _, scalar := range scalars {
		// Get the label corresponding to the scalar.
		label := class[scalar]

		fmt.Printf("Scalar: %v, Label: %v\n", scalar, label)
	}

	// 9926980098944308213 is the scalar of "do".
	scalar := bayes.Scalar(9926980098944308213)

	{
		// Access via key in the map.
		label := class[scalar]
		fmt.Printf("Label of %v is %v\n", scalar, label)
	}
	{
		// Access via method.
		label, err := class.Label(scalar)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Label of %v is %v\n", scalar, label)
	}

	// Output:
	// Scalar: 1936869697095642849, Label: fa
	// Scalar: 2360553998378955785, Label: so
	// Scalar: 3107976362489010529, Label: re
	// Scalar: 9638313718333727784, Label: mi
	// Scalar: 9926980098944308213, Label: do
	// Scalar: 14551593800680666286, Label: la
	// Scalar: 17238071966115729737, Label: si
	// Label of 9926980098944308213 is do
	// Label of 9926980098944308213 is do
}

func ExampleToClass_two_dimention() {
	// Sample data in 2D matrix. Currently, it does not support the 3D or more
	// dimentional matrix.
	sample := [][]any{
		{"do", "re", "mi", "fa"},
		{"so", "la", "si", "do"},
	}

	// ToClass function returns a bayas.Class object from the given sample data.
	// The object is the map of Scalar (map[bayes.Scalar]bayes.Label).
	class, err := bayes.ToClass(sample)
	if err != nil {
		log.Fatal(err, "foo")
	}

	// Get list of scalar values in the class map.
	scalars := class.Keys()

	for _, scalar := range scalars {
		label := class[scalar]

		fmt.Printf("Scalar: %v, Label: %v\n", scalar, label)
	}

	// Output:
	// Scalar: 1936869697095642849, Label: fa
	// Scalar: 2360553998378955785, Label: so
	// Scalar: 3107976362489010529, Label: re
	// Scalar: 9638313718333727784, Label: mi
	// Scalar: 9926980098944308213, Label: do
	// Scalar: 14551593800680666286, Label: la
	// Scalar: 17238071966115729737, Label: si
}

func ExampleToVector() {
	sample := []any{
		"do", "re", "mi", "fa", "so", "la", "si", "do",
	}

	vector, err := bayes.ToVector(sample)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(dd.Dump(vector))
	// Output:
	// bayes.Vector{
	//   9926980098944308213,
	//   3107976362489010529,
	//   9638313718333727784,
	//   1936869697095642849,
	//   2360553998378955785,
	//   14551593800680666286,
	//   17238071966115729737,
	//   9926980098944308213,
	// }
}

func ExampleToMatrix2D() {
	sample := [][]any{
		{"do", "re", "mi", "fa"},
		{"so", "la", "si", "do"},
	}

	matrix, err := bayes.ToMatrix2D(sample)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(dd.Dump(matrix))
	// Output:
	// bayes.Matrix2D{
	//   bayes.Vector{
	//     9926980098944308213,
	//     3107976362489010529,
	//     9638313718333727784,
	//     1936869697095642849,
	//   },
	//   bayes.Vector{
	//     2360553998378955785,
	//     14551593800680666286,
	//     17238071966115729737,
	//     9926980098944308213,
	//   },
	// }
}
