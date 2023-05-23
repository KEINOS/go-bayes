package trainer_test

import (
	"fmt"
	"log"

	"github.com/KEINOS/go-bayes/pkg/dataset"
	"github.com/KEINOS/go-bayes/pkg/kvs"
	"github.com/KEINOS/go-bayes/pkg/predictor"
	"github.com/KEINOS/go-bayes/pkg/trainer"
	"github.com/KEINOS/go-bayes/pkg/utils"
)

func Example() {
	db, err := trainer.New(kvs.InMemoryDB)
	if err != nil {
		log.Fatal(err)
	}

	convToUID := func(str string) uint64 {
		id, err := utils.AnyToStateID(str)
		if err != nil {
			log.Fatal(err)
		}

		return id
	}

	Tokyo := convToUID("Tokyo")
	Osaka := convToUID("Osaka")
	Kyoto := convToUID("Kyoto")

	db.Update(Tokyo, Osaka)
	db.Update(Osaka, Kyoto)
	db.Update(Tokyo, Kyoto)

	// Prior probabilities to be B
	fmt.Println("To Tokyo:", db.PriorPtoB(Tokyo))
	fmt.Println("To Osaka:", db.PriorPtoB(Osaka))
	fmt.Println("To Kyoto:", db.PriorPtoB(Kyoto))

	// Prior probabilities from A to B
	fmt.Println("From Tokyo to Tokyo:", db.PriorPfromAtoB(Tokyo, Tokyo))
	fmt.Println("From Tokyo to Osaka:", db.PriorPfromAtoB(Tokyo, Osaka))
	fmt.Println("From Tokyo to Kyoto:", db.PriorPfromAtoB(Tokyo, Kyoto))

	// Prior probabilities from A but not to B
	fmt.Println("Not from Tokyo to Tokyo", db.PriorPNotFromAtoB(Tokyo, Tokyo))
	fmt.Println("Not from Tokyo to Osaka", db.PriorPNotFromAtoB(Tokyo, Osaka))
	fmt.Println("Not from Tokyo to Kyoto", db.PriorPNotFromAtoB(Tokyo, Kyoto))

	// Output:
	// To Tokyo: 0
	// To Osaka: 0.3333333333333333
	// To Kyoto: 0.6666666666666666
	// From Tokyo to Tokyo: 0
	// From Tokyo to Osaka: 0.5
	// From Tokyo to Kyoto: 0.5
	// Not from Tokyo to Tokyo 0.6666666666666666
	// Not from Tokyo to Osaka 0.3333333333333333
	// Not from Tokyo to Kyoto 0.3333333333333333
}

func ExampleTrained_Train() {
	exitOnError := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	// Sice of any type as a data to train.
	corpus := []any{
		"Do", "Re", "Mi", "Fa", "So", "La", "Ti",
	}

	// Instanciate a dataset.
	dataSet, err := dataset.New(kvs.InMemoryDB)
	exitOnError(err)

	// Convert corpus to dataset and get the list of classes to be predicted.
	class, err := dataSet.AddFromSlice(corpus)
	exitOnError(err)

	// Instanciate the trainer.
	train, err := trainer.New(kvs.InMemoryDB)
	exitOnError(err)

	// Train the model.
	train.Train(dataSet)

	// Instanciate the predictor.
	predict := predictor.New(train, *class)

	listPredicted, err := predict.Predict(0)
	exitOnError(err)

	for _, predicted := range listPredicted {
		fmt.Printf(
			"Class: %v, Prov: %f, Value: %v\n",
			predicted.ClassID,
			predicted.Prob,
			predicted.RawValue,
		)
	}

	// Output:
}
