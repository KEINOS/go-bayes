package data

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
//  Functions
// ============================================================================

// ----------------------------------------------------------------------------
//  DecodeItem()
// ----------------------------------------------------------------------------

func TestDecodeItem_fail_decode(t *testing.T) {
	item, err := DecodeItem(nil)

	require.Error(t, err, "nil value should return error")
	require.Nil(t, item, "returned value should be nil on error")
}

// ============================================================================
//  Methods
// ============================================================================

// ----------------------------------------------------------------------------
//  Item.Encoded()
// ----------------------------------------------------------------------------

// func TestItem_Encoded_forced_fail(t *testing.T) {
// 	item := Item{Value: nil, UID: 1}

// 	encV, err := item.Encoded()

// 	require.Error(t, err, "nil value should return error")
// 	require.Nil(t, encV, "returned value should be nil on error")
// }

// ----------------------------------------------------------------------------
//  Item.IsValidUID()
// ----------------------------------------------------------------------------

func TestItem_IsValidUID_invalid(t *testing.T) {
	{
		item := Item{Value: "1", UID: 0}

		require.False(t, item.IsValidUID(), "non-matching UID should return false")
	}
	{
		item := Item{Value: nil, UID: 1}

		require.False(t, item.IsValidUID(), "nil value should return false")
	}
}
