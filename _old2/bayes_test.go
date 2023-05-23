package bayes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToClass(t *testing.T) {
	data := []any{
		1, "1", int32(1), int64(1), float32(1.0), float64(1.0),
	}

	class, err := ToClass(data)
	require.NoError(t, err)

	scalars := class.Keys() // sorted

	actualLen := len(scalars)
	expectLen := len(data)
	require.Equal(t, expectLen, actualLen)

	expectScalars := []Scalar{
		0x29f55af68ab3fbfa, // float64(1.0)
		0x3eb603e1c1bc5636, // int(1)
		0x60be489703513908, // int64(1)
		0xa7cdd6f4cd7a3040, // string("1")
		0xac48e405eea42a2e, // float32(1.0)
		0xeabd6cab4e5f4d2f, // int32(1)
	}

	require.Equal(t, expectScalars, scalars)
}

func Test_train_golden(t *testing.T) {
	// One row of data to train
	trainData := []uint64{1, 2, 3, 4, 5, 6}

	trainer := NewTrainer()
	trainer.trainRow(trainData)

	expectTotalAccesses := 15
	expectFromAToB := map[uint64]map[uint64]int{
		1:                    {2: 1},
		2:                    {3: 1},
		3:                    {4: 1},
		4:                    {5: 1},
		5:                    {6: 1},
		1234123722237457922:  {6: 1},
		1400453063984103771:  {5: 1},
		1503651559852751965:  {6: 1},
		1950187959935471838:  {4: 1},
		2063133131322548733:  {4: 1},
		9136711267509372717:  {5: 1},
		10971205263938642109: {6: 1},
		13014536700492816144: {6: 1},
		17760205484716448731: {5: 1},
		18295801019946604924: {3: 1},
	}
	expectFromA := map[uint64]int{
		1:                    1,
		2:                    1,
		3:                    1,
		4:                    1,
		5:                    1,
		1234123722237457922:  1,
		1400453063984103771:  1,
		1503651559852751965:  1,
		1950187959935471838:  1,
		2063133131322548733:  1,
		9136711267509372717:  1,
		10971205263938642109: 1,
		13014536700492816144: 1,
		17760205484716448731: 1,
		18295801019946604924: 1,
	}
	expectToB := map[uint64]int{
		2: 1,
		3: 2,
		4: 3,
		5: 4,
		6: 5,
	}

	require.Equal(t, expectTotalAccesses, trainer.TotalAccesses)
	require.Equal(t, expectFromAToB, trainer.FromAToB)
	require.Equal(t, expectFromA, trainer.FromA)
	require.Equal(t, expectToB, trainer.ToB)
}

func TestTrainer_Train_golden(t *testing.T) {
	trainData := [][]uint64{
		{1, 2, 3}, // 1->2, 2->3, 1->2->3
		{2, 3, 4}, // 2->3, 3->4, 2->3->4
	}

	sample := &Sample{
		Data: trainData,
	}

	trainer := NewTrainer()
	trainer.Train(sample)

	expectTotalAccesses := 6
	expectFromAToB := map[uint64]map[uint64]int{
		1:                    {2: 1},
		2:                    {3: 2},
		3:                    {4: 1},
		18295801019946604924: {3: 1}, // 1->2->3
		1950187959935471838:  {4: 1}, // 2->3->4
	}
	expectFromA := map[uint64]int{
		1:                    1,
		2:                    2,
		3:                    1,
		1950187959935471838:  1,
		18295801019946604924: 1,
	}
	expectToB := map[uint64]int{
		2: 1,
		3: 3,
		4: 2,
	}

	require.Equal(t, expectTotalAccesses, trainer.TotalAccesses)
	require.Equal(t, expectFromAToB, trainer.FromAToB)
	require.Equal(t, expectFromA, trainer.FromA)
	require.Equal(t, expectToB, trainer.ToB)

	// fmt.Println(dd.Dump(trainer))
	// t.Fail()
}
