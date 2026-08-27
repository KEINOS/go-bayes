//go:build cgo

package sqlitestore_test

import (
	"context"
	"fmt"
	"iter"
	"log"
	"os"
	"path/filepath"

	"github.com/KEINOS/go-bayes/bayes/internal/modelstores/sqlitestore"
	"github.com/KEINOS/go-bayes/bayes/modelstore"
)

//nolint:funlen // Keep create, close, reopen, and query in one complete example.
func ExampleStore() {
	ctx := context.Background()
	directory, err := os.MkdirTemp("", "go-bayes-sqlite-example-")
	if err != nil {
		log.Panic(err)
	}

	defer func() {
		err := os.RemoveAll(directory)
		if err != nil {
			log.Panic(err)
		}
	}()

	path := filepath.Join(directory, "model.db")
	metadata := sqlitestore.Metadata{
		CodecVersion: 1,
		HasherName:   "example",
		ItemProbe:    1,
		ContextProbe: 2,
		ScopeID:      42,
	}
	store, err := sqlitestore.Create(ctx, path, metadata, sqlitestore.OpenConfig{})
	if err != nil {
		log.Panic(err)
	}

	batch := modelstore.TrainingBatch{
		Classes: []modelstore.Class{{ID: 2, TypeTag: 2, Payload: []byte("answer")}},
		Transitions: func() iter.Seq[modelstore.TransitionDelta] {
			return func(yield func(modelstore.TransitionDelta) bool) {
				yield(modelstore.TransitionDelta{FromID: 10, ToID: 2, Count: 5})
			}
		},
	}
	err = store.Apply(ctx, batch)
	if err != nil {
		log.Panic(err)
	}

	err = store.Close()
	if err != nil {
		log.Panic(err)
	}

	store, err = sqlitestore.Open(ctx, path, sqlitestore.OpenConfig{})
	if err != nil {
		log.Panic(err)
	}

	defer func() {
		err := store.Close()
		if err != nil {
			log.Panic(err)
		}
	}()

	stats, err := store.Stats(ctx, 10)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(store.ScopeID())
	fmt.Println(stats.Total, stats.FromCount)
	fmt.Println(stats.Candidates[0].ClassID, stats.Candidates[0].PairCount)
	// Output:
	// 42
	// 5 5
	// 2 5
}
