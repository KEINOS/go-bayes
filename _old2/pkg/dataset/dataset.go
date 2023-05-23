package dataset

import (
	"github.com/KEINOS/go-bayes/pkg/kvs"
	"github.com/KEINOS/go-bayes/pkg/utils"
	"github.com/pkg/errors"
)

type DB struct {
	kvs.DB
	dbType kvs.DBType
}

func New(dbType kvs.DBType) (*DB, error) {
	db := kvs.New(dbType)

	if db == nil {
		return nil, errors.Errorf("failed to create DB. DB type: %v", dbType)
	}

	return &DB{
		DB:     db,
		dbType: dbType,
	}, nil
}

func (db *DB) AddFromSlice(data []any) (classDB *DB, err error) {
	dbType := db.dbType

	classes := kvs.New(dbType)
	if classes == nil {
		return nil, errors.Errorf("failed to create Class DB. DB type: %v", db.dbType)
	}

	for i, anyValue := range data {
		id, err := utils.AnyToStateID(anyValue)
		if err != nil {
			return nil, errors.Wrap(err, "failed to add items from slice")
		}

		db.DB.Set(uint64(i), id)
		classes.Set(id, anyValue)
	}

	return &DB{
		DB:     classes,
		dbType: dbType,
	}, nil
}

func (db *DB) Get(key any) (any, error) {
	intKey, ok := key.(uint64)
	if !ok {
		return nil, errors.Errorf("key type is not uint64")
	}

	return db.DB.GetValue(intKey)
}
