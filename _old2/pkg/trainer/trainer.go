package trainer

import (
	"github.com/KEINOS/go-bayes/pkg/dataset"
	"github.com/KEINOS/go-bayes/pkg/kvs"
	"github.com/KEINOS/go-bayes/pkg/utils"
	"github.com/pkg/errors"
)

// ----------------------------------------------------------------------------
//  Type: Trainer
// ----------------------------------------------------------------------------

// Trainer is a wrapper of kvs.DB that stores the records of transitions.
type Trainer interface {
	PriorPfromAtoB(fromA uint64, toB uint64) float64
	PriorPNotFromAtoB(fromA uint64, toB uint64) float64
	PriorPtoB(toB uint64) float64
	Train(dataSet *dataset.DB) error
	Update(fromA, toB uint64)
}

// ----------------------------------------------------------------------------
//  Constructor
// ----------------------------------------------------------------------------

type Trained struct {
	kvs.DB
}

func New(dbType kvs.DBType) (Trainer, error) {
	db := kvs.New(dbType)
	if db == nil {
		return nil, errors.New("failed to create a kvs.DB")
	}

	return &Trained{DB: db}, nil
}

// ----------------------------------------------------------------------------
//  Public Methods
// ----------------------------------------------------------------------------

func (db Trained) GetFromA(fromA uint64) int {
	key := db.getKeyFromA(fromA)

	return db.getValue(key)
}

func (db Trained) GetFromAtoB(fromA, toB uint64) int {
	key := db.getKeyFromAtoB(fromA, toB)

	return db.getValue(key)
}

func (db Trained) GetToB(toB uint64) int {
	key := db.getKeyToB(toB)

	return db.getValue(key)
}

func (db Trained) GetTotalAccess() int {
	key := db.getKeyTotalAccess()

	return db.getValue(key)
}

func (db Trained) Increment(key uint64) {
	value := db.getValue(key)

	value++

	db.setValue(key, value)
}

func (db Trained) IncrementTotalAccess() {
	key := db.getKeyTotalAccess()

	db.Increment(key)
}

func (db Trained) IncrementToB(toB uint64) {
	key := db.getKeyToB(toB)

	db.Increment(key)
}

func (db Trained) IncrementFromA(fromA uint64) {
	key := db.getKeyFromA(fromA)

	db.Increment(key)
}

func (db Trained) IncrementFromAtoB(fromA, toB uint64) {
	key := db.getKeyFromAtoB(fromA, toB)

	db.Increment(key)
}

func (db Trained) PriorPfromAtoB(fromA uint64, toB uint64) float64 {
	x := db.GetFromAtoB(fromA, toB)
	y := db.GetFromA(fromA)

	return float64(x) / float64(y)
}

func (db Trained) PriorPNotFromAtoB(fromA uint64, toB uint64) float64 {
	// If there is no access, never reaches to node B. Or, if there is no
	// access from node A, never reaches to node B as well.
	// Therefore, the probability of not reaching node B is 100%.
	if db.GetTotalAccess() == 0 || db.GetFromA(fromA) == 0 {
		return 1
	}

	notFromAtoB := db.GetFromA(fromA) - db.GetFromAtoB(fromA, toB)

	return float64(notFromAtoB) / float64(db.GetTotalAccess())
}

func (db Trained) PriorPtoB(toB uint64) float64 {
	x := db.GetToB(toB)
	y := db.GetTotalAccess()

	return float64(x) / float64(y)
}

func (db Trained) recursion(transitionData []uint64, toB uint64) {
	lenData := len(transitionData)

	for i := 0; i < lenData; i++ {
		fromAs := transitionData[i:]
		fromA := utils.IntsToStateID(fromAs)

		db.Update(fromA, toB)
	}
}

func (db Trained) Train(dataSet *dataset.DB) error {
	dataSet.Reset()

	postItem := uint64(0)
	recurrent := []uint64{}

	for dataSet.HasNext() {
		i, data, err := dataSet.Next()
		if err != nil {
			return errors.Wrap(err, "failed to get train data from the dataset")
		}

		id, ok := data.(uint64)
		if !ok {
			return errors.Errorf("failed to assert type of train data. Type: %T", data)
		}

		if i == 0 {
			postItem = id
			recurrent = append(recurrent, id)

			continue
		}

		fromA := postItem
		toB := id

		db.Update(fromA, toB)
		db.recursion(recurrent, toB)

		postItem = id
		recurrent = append(recurrent, id)
	}

	return nil
}

func (db Trained) Update(fromA, toB uint64) {
	db.IncrementFromAtoB(fromA, toB) // update FromAtoB
	db.IncrementFromA(fromA)         // update FromA
	db.IncrementToB(toB)             // update ToB
	db.IncrementTotalAccess()        // update TotalAccess
}

// ----------------------------------------------------------------------------
//  Private Methods
// ----------------------------------------------------------------------------

func (db Trained) getKeyFromA(fromA uint64) uint64 {
	fieldType := utils.BytesToStateID([]byte("formA"))
	fieldID := utils.IntsToStateID([]uint64{fromA})

	return utils.IntsToStateID([]uint64{fieldType, fieldID})
}

func (db Trained) getKeyFromAtoB(fromA, toB uint64) uint64 {
	fieldType := utils.BytesToStateID([]byte("formAtoB"))
	fieldID := utils.IntsToStateID([]uint64{fromA, toB})

	return utils.IntsToStateID([]uint64{fieldType, fieldID})
}

func (db Trained) getKeyToB(toB uint64) uint64 {
	fieldType := utils.BytesToStateID([]byte("toB"))
	fieldID := utils.IntsToStateID([]uint64{toB})

	return utils.IntsToStateID([]uint64{fieldType, fieldID})
}

func (db Trained) getKeyTotalAccess() uint64 {
	key := uint64(0xFFFFFFFFFFFFFFFF) // 18446744073709551615

	return key
}

func (db Trained) getValue(key uint64) int {
	anyValue, err := db.DB.GetValue(key)
	intValue, ok := anyValue.(int)

	if err != nil || !ok {
		intValue = 0
	}

	return intValue
}

func (db Trained) setValue(key uint64, value int) {
	if err := db.DB.Set(key, value); err != nil {
		panic(err)
	}
}
