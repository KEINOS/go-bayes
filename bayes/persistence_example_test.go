//go:build cgo

package bayes_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/KEINOS/go-bayes/bayes"
)

func ExamplePredictor_Save() {
	ctx := context.Background()
	directory, err := os.MkdirTemp("", "go-bayes-save-example-")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := os.RemoveAll(directory)
		if err != nil {
			log.Panic(err)
		}
	}()

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

	err = predictor.Train(ctx, []string{"A", "B", "C"})
	if err != nil {
		log.Panic(err)
	}

	path := filepath.Join(directory, "model.db")
	err = predictor.Save(ctx, path)
	if err != nil {
		log.Panic(err)
	}

	restored, err := bayes.Load(ctx, path)
	if err != nil {
		log.Panic(err)
	}

	defer func() {
		err := restored.Close()
		if err != nil {
			log.Panic(err)
		}
	}()

	classID, err := restored.Predict(ctx, []string{"A", "B"})
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(restored.GetClass(classID))
	// Output: C
}

func ExampleOpen() {
	ctx := context.Background()
	directory, err := os.MkdirTemp("", "go-bayes-open-example-")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := os.RemoveAll(directory)
		if err != nil {
			log.Panic(err)
		}
	}()

	path := filepath.Join(directory, "model.db")
	predictor, err := bayes.New(
		ctx,
		bayes.SQLiteStorage,
		100,
		bayes.WithSQLitePath(path),
	)
	if err != nil {
		log.Panic(err)
	}

	err = predictor.Train(ctx, []string{"A", "B", "C"})
	if err != nil {
		log.Panic(err)
	}

	err = predictor.Close()
	if err != nil {
		log.Panic(err)
	}

	predictor, err = bayes.Open(ctx, path)
	if err != nil {
		log.Panic(err)
	}

	defer func() {
		err := predictor.Close()
		if err != nil {
			log.Panic(err)
		}
	}()

	classID, err := predictor.Predict(ctx, []string{"A", "B"})
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(predictor.GetClass(classID))
	// Output: C
}
