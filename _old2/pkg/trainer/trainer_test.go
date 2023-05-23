package trainer

import (
	"testing"

	"github.com/Code-Hex/dd"
	"github.com/KEINOS/go-bayes/pkg/kvs"
	"github.com/KEINOS/go-bayes/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestTrained_Increment(t *testing.T) {
	trainer, err := New(kvs.InMemoryDB)
	require.NoError(t, err)

	db, ok := trainer.(Trained)
	require.True(t, ok)

	X := utils.IntsToStateID([]uint64{1})
	Y := utils.IntsToStateID([]uint64{2})
	Z := utils.IntsToStateID([]uint64{3})

	require.Zero(t, 0, db.GetFromAtoB(X, Y), "transition never happened")
	require.Zero(t, 0, db.GetFromA(X), "transition never happened")
	require.Zero(t, 0, db.GetToB(Y), "transition never happened")
	require.Zero(t, 0, db.GetToB(Z), "transition never happened")

	t.Log(dd.Dump(db))

	db.Update(X, Y) // from X to Y
	db.Update(X, Y) // from X to Y
	db.Update(Y, Z) // from Y to Z
	db.Update(X, Z) // from X to Z

	t.Log(dd.Dump(db))

	require.Equal(t, 2, db.GetFromAtoB(X, Y), "transition from X to Y happened twice")
	require.Equal(t, 1, db.GetFromAtoB(Y, Z), "transition from Y to Z happened once")
	require.Equal(t, 1, db.GetFromAtoB(X, Z), "transition from X to Z happened once")
	require.Equal(t, 0, db.GetFromA(Z), "transition from Z never happened")
	require.Equal(t, 1, db.GetFromA(Y), "transition from Y happened once")
	require.Equal(t, 0, db.GetToB(X), "transition to X never happened")
	require.Equal(t, 2, db.GetToB(Y), "transition to Y happened twice")
	require.Equal(t, 2, db.GetToB(Z), "transition to Z happened once")
	require.Equal(t, 4, db.GetTotalAccess(), "total access happened four times")

	t.Fail()
}
