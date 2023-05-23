package logmem

import (
	"testing"

	"github.com/KEINOS/go-bayes/pkg/class"
	"github.com/stretchr/testify/require"
)

// ============================================================================
//  Functions
// ============================================================================

// ----------------------------------------------------------------------------
//  NewMap()
// ----------------------------------------------------------------------------

func TestNewMap_err(t *testing.T) {
	mapClass, err := NewMap(make(chan int))

	require.Error(t, err)
	require.Nil(t, mapClass)
	require.Contains(t, err.Error(), "failed to add Class objects")
	require.Contains(t, err.Error(), "failed to instantiate Map object")
}

// ============================================================================
//  Methods
// ============================================================================

// ----------------------------------------------------------------------------
//  Map.AddAny()
// ----------------------------------------------------------------------------

func TestMap_AddAny_id_collision(t *testing.T) {
	classTmp1, err := class.New("foo")
	require.NoError(t, err)

	classTmp2, err := class.New("bar")
	require.NoError(t, err)

	mapClass := &Map{
		classTmp1.ID: *classTmp2,
		classTmp2.ID: *classTmp1,
	}

	err = mapClass.AddAny("bar")

	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate class ID with different type of value")
}

// ----------------------------------------------------------------------------
//  Map.GetClassID()
// ----------------------------------------------------------------------------

func TestMap_GetClassID_unsupported_type(t *testing.T) {
	mapClass, err := NewMap("foo")

	require.NoError(t, err)

	classID, err := mapClass.GetClassID(nil)

	require.Error(t, err)
	require.Zero(t, classID, "class ID should be zero on error")
	require.Contains(t, err.Error(), "failed to instantiate Class object")
	require.Contains(t, err.Error(), "Unsupported type: <nil>")
}

func TestMap_GetClassID_unregistered_item(t *testing.T) {
	mapClass, err := NewMap("foo")

	require.NoError(t, err)

	classID, err := mapClass.GetClassID("bar")

	require.Error(t, err)
	require.Zero(t, classID, "class ID should be zero on error")
	require.Contains(t, err.Error(), "item bar is not in the map")
}
