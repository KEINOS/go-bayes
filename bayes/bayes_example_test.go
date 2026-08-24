package bayes_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/KEINOS/go-bayes/bayes"
)

func Example() {
	score := []string{
		"So", "So", "La", "So", "Do", "Si",
		"So", "So", "La", "So", "Re", "Do",
		"So", "So", "So", "Mi", "Do", "Si", "La",
		"Fa", "Fa", "Mi", "Do", "Re", "Do",
	}

	scopeID := uint64(100)

	foo, err := bayes.New(context.Background(), bayes.MemoryStorage, scopeID)
	if err != nil {
		log.Fatal(err)
	}

	err = foo.Train(context.Background(), score)
	if err != nil {
		log.Fatal(err)
	}

	nextNoteID, err := foo.Predict(context.Background(), []string{"So", "So", "La", "So", "Do", "Si"})
	if err != nil {
		log.Fatal(err)
	}

	nextNoteString := foo.GetClass(nextNoteID)

	fmt.Printf("Next is: %v (Class ID: %v)\n", nextNoteString, nextNoteID)
	//
	// Output:
	// Next is: So (Class ID: 2326549176558064863)
}

//nolint:funlen // allow long function due to example
func Example_iris() {
	// Prepare the iris dataset for training and prediction.
	type irisData struct {
		Data   [][4]float64 `json:"data"`
		Target []float64    `json:"target"`
	}

	loadIrisData := func() (irisData, error) {
		pathJSON := filepath.Join("..", "testdata", "iris.json")
		pathJSON = filepath.Clean(pathJSON)

		jsonFile, err := os.Open(pathJSON)
		if err != nil {
			return irisData{}, fmt.Errorf("failed to open iris testdata: %w", err)
		}

		defer func() {
			_ = jsonFile.Close()
		}()

		var dataset irisData

		err = json.NewDecoder(jsonFile).Decode(&dataset)
		if err != nil {
			return irisData{}, fmt.Errorf("failed to decode iris testdata: %w", err)
		}

		return dataset, nil
	}

	// Load the iris dataset from the JSON file.
	iris, err := loadIrisData()
	if err != nil {
		log.Panic(err)
	}

	// Create an independent predictor with an application-defined scope ID.
	scopeID := uint64(4)

	predictor, err := bayes.New(context.Background(), bayes.MemoryStorage, scopeID)
	if err != nil {
		log.Panic(err)
	}

	// Train the predictor with the iris dataset.
	// Each row consists of 4 features followed by the target class.
	for i, row := range iris.Data {
		// Sepal length, sepal width, petal length, petal width, and target class.
		drill := []float64{row[0], row[1], row[2], row[3], iris.Target[i]}

		err := predictor.Train(context.Background(), drill)
		if err != nil {
			log.Panic(err)
		}
	}

	// Predict the class for a new data point.
	// [0] = sepal length, [1] = sepal width, [2] = petal length, [3] = petal width.
	predictedID, err := predictor.Predict(context.Background(), []float64{5.1, 3.5, 1.4, 0.2})
	if err != nil {
		log.Panic(err)
	}

	// Retrieve the predicted class name.
	class, ok := predictor.GetClass(predictedID).(float64)
	if !ok {
		log.Panic("unexpected class type")
	}

	species := map[float64]string{
		0: "setosa",
		1: "versicolor",
		2: "virginica",
	}

	fmt.Println(species[class])
	//
	// Output: setosa
}

// ----------------------------------------------------------------------------
//  New()
// ----------------------------------------------------------------------------

func ExampleNew() {
	// The predictor and its storage logger report this application-defined ID.
	scopeID := uint64(100)

	// Create a predictor with in-memory storage.
	predictor, err := bayes.New(context.Background(), bayes.MemoryStorage, scopeID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(predictor.ID())
	//
	// Output: 100
}

func ExamplePredictor_foldedContexts() {
	ctx := context.Background()
	predictor, err := bayes.New(ctx, bayes.MemoryStorage, 100)
	if err != nil {
		log.Panic(err)
	}

	defer func() {
		err := predictor.Close()
		if err != nil {
			log.Panic(err)
		}
	}()

	err = predictor.Train(ctx, []string{"A", "B", "C", "D"})
	if err != nil {
		log.Panic(err)
	}

	// Training records each suffix of the ordered context.
	for _, input := range [][]string{{"A", "B", "C"}, {"B", "C"}, {"C"}} {
		classID, err := predictor.Predict(ctx, input)
		if err != nil {
			log.Panic(err)
		}

		fmt.Println(predictor.GetClass(classID))
	}

	// Output:
	// D
	// D
	// D
}

func ExamplePredictor_Reset() {
	ctx := context.Background()
	predictor, err := bayes.New(ctx, bayes.MemoryStorage, 100)
	if err != nil {
		log.Fatal(err)
	}

	err = predictor.Train(ctx, []string{"A", "B"})
	if err != nil {
		log.Fatal(err)
	}

	err = predictor.Reset(ctx)
	if err != nil {
		log.Fatal(err)
	}

	classID, err := predictor.Predict(ctx, []string{"A"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(classID)
	fmt.Println(predictor.GetClass(classID))
	// Output:
	// 0
	// <nil>
}

func ExampleWithHasher() {
	predictor, err := bayes.New(
		context.Background(),
		bayes.MemoryStorage,
		100,
		bayes.WithHasher("blake3"),
	)
	if err != nil {
		log.Fatal(err)
	}

	flowID, err := predictor.HashTrans(10, 11, 12, 13, 14, 15)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(flowID)
	// Output: 6611782858352040389
}

// ----------------------------------------------------------------------------

func ExampleStorage_Type() {
	fmt.Println(bayes.MemoryStorage.Type())
	fmt.Println(bayes.UnknownStorage.Type())
	//
	// Output:
	// in-memory
	// unknown
}
