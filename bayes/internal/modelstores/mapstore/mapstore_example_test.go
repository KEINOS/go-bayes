package mapstore_test

import (
	"context"
	"fmt"
	"iter"
	"log"

	"github.com/KEINOS/go-bayes/bayes/internal/modelstores/mapstore"
	"github.com/KEINOS/go-bayes/bayes/modelstore"
)

func ExampleStore() {
	ctx := context.Background()
	store := mapstore.New(42)

	defer func() {
		err := store.Close()
		if err != nil {
			log.Panic(err)
		}
	}()

	batch := modelstore.TrainingBatch{
		Classes: []modelstore.Class{{ID: 2, TypeTag: 2, Payload: []byte("answer")}},
		Transitions: func() iter.Seq[modelstore.TransitionDelta] {
			return func(yield func(modelstore.TransitionDelta) bool) {
				yield(modelstore.TransitionDelta{FromID: 10, ToID: 2, Count: 5})
			}
		},
	}

	err := store.Apply(ctx, batch)
	if err != nil {
		log.Panic(err)
	}

	stats, err := store.Stats(ctx, 10)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(store.ScopeID())
	fmt.Println(stats.Total, stats.FromCount)
	fmt.Println(stats.Candidates[0].ClassID, stats.Candidates[0].PairCount)

	classes, err := store.Classes(ctx)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(string(classes[0].Payload))

	err = store.ExportTransitions(ctx, func(record modelstore.TransitionCount) error {
		fmt.Println(record.FromID, record.ToID, record.Count)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	err = store.Reset(ctx)
	if err != nil {
		log.Panic(err)
	}

	stats, err = store.Stats(ctx, 10)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(stats.Total)
	// Output:
	// 42
	// 5 5
	// 2 5
	// answer
	// 10 2 5
	// 0
}
