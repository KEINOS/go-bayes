//go:build !cgo

package bayes

import "context"

func loadSQLiteModel(context.Context, string, Hasher) (*Predictor, error) {
	return nil, ErrSQLiteUnavailable
}

func newSQLiteStore(context.Context, PredictorConfig) (ModelStore, error) {
	return nil, ErrSQLiteUnavailable
}

func openSQLiteModel(context.Context, string, PredictorConfig) (*Predictor, error) {
	return nil, ErrSQLiteUnavailable
}

func saveModel(context.Context, *Predictor, string) error {
	return ErrSQLiteUnavailable
}
