package logmem

import (
	"bytes"
	"testing"

	"github.com/KEINOS/go-bayes/pkg/dumptype"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
//  Helper functions
// ----------------------------------------------------------------------------

// DummyWriter is a dummy writer to test write error case.
type DummyWriter struct{}

// Write is a dummy writer to test write error case.
func (d *DummyWriter) Write(p []byte) (int, error) {
	return 0, errors.New("forced error to write")
}

// ----------------------------------------------------------------------------
//  NodeLog
// ----------------------------------------------------------------------------

func TestNodeLog_zero_division(t *testing.T) {
	n := New(12345)

	require.Equal(t, n.Predict(1, 2), float64(0))
	require.Equal(t, n.PriorPfromAtoB(1, 2), float64(0))
	require.Equal(t, n.PriorPNotFromAtoB(1, 2), float64(1))
	require.Equal(t, n.PriorPtoB(1), float64(0))
	require.NotPanics(t, func() { n.Update(1, 2) })
}

// ----------------------------------------------------------------------------
//  NodeLog.Restore()
// ----------------------------------------------------------------------------

func TestNodeLog_Restore_arg_is_nil(t *testing.T) {
	for _, dumpType := range []dumptype.DumpType{
		dumptype.JSON,
		dumptype.GOB,
	} {
		err := New(12345).Restore(dumpType, nil)

		require.Error(t, err)

		assert.Contains(t, err.Error(), "restore failed: the io.Reader r is nil")
	}
}

func TestNodeLog_Restore_dumptype_unknown(t *testing.T) {
	for _, dumpType := range []dumptype.DumpType{
		dumptype.CSV,
		dumptype.SQL,
		dumptype.Unknown,
		dumptype.DumpType(999999),
	} {
		err := New(12345).Restore(dumpType, new(bytes.Buffer))

		require.Error(t, err)

		assert.Contains(t, err.Error(), "restore failed: unsupported dump type")
	}
}

// ----------------------------------------------------------------------------
//  NodeLog.Store()
// ----------------------------------------------------------------------------

func TestNodeLog_Store_arg_is_nil(t *testing.T) {
	for _, dumpType := range []dumptype.DumpType{
		dumptype.JSON,
		dumptype.GOB,
	} {
		err := New(12345).Store(dumpType, nil, nil)

		require.Error(t, err)

		assert.Contains(t, err.Error(), "dump failed: the io.Writer w is nil")
	}
}

func TestNodeLog_storeGob_fail_encode(t *testing.T) {
	var d DummyWriter

	err := new(NodeLog).storeGob(&d)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dump failed: can not encode to gob")
	assert.Contains(t, err.Error(), "forced error to write")
}

func TestNodeLog_storeJSON_fail_encode(t *testing.T) {
	var d DummyWriter

	err := new(NodeLog).storeJSON(&d)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dump failed: can not write NodeLog to JSON")
	assert.Contains(t, err.Error(), "forced error to write")
}

func TestNodeLog_Store_dumptype_unknown(t *testing.T) {
	for _, dumpType := range []dumptype.DumpType{
		dumptype.CSV,
		dumptype.SQL,
		dumptype.Unknown,
		dumptype.DumpType(999999),
	} {
		err := New(12345).Store(dumpType, nil, new(bytes.Buffer))

		require.Error(t, err)

		assert.Contains(t, err.Error(), "dump failed: unsupported dump type")
	}
}

func TestNodeLog_Store_fail_marshal(t *testing.T) {
	oldJSONMarshalIndent := jsonMarshalIndent
	defer func() {
		jsonMarshalIndent = oldJSONMarshalIndent
	}()

	jsonMarshalIndent = func(v any, prefix string, indent string) ([]byte, error) {
		return nil, errors.New("forced error to marshal")
	}

	l := New(12345)

	var b bytes.Buffer

	err := l.Store(dumptype.JSON, nil, &b)

	require.Error(t, err, "it should return error. Output: %v", b.String())

	assert.Contains(t, err.Error(), "dump failed: can not marshal NodeLog to JSON")
	assert.Contains(t, err.Error(), "forced error to marshal")
}

// ----------------------------------------------------------------------------
//  NodeLog.PriorPNotFromAtoB()
// ----------------------------------------------------------------------------

func TestNodeLog_PriorPNotFromAtoB(t *testing.T) {
	const (
		X = 1 // node #1
		Y = 2 // node #2
		Z = 3 // node #3
	)

	n := New(12345)

	require.Equal(t, float64(1), n.PriorPNotFromAtoB(X, Y),
		"if no access, never reaches to Y. Thus, it should be 1 = 100%")

	// Train: from X to Y (node #1 --> #2)
	n.Update(X, Y)
	require.Equal(t, float64(0), n.PriorPNotFromAtoB(X, Y),
		"1 access to Y out of 1 access total, so it should be 0 = 0%")

	// Train: from X to Z (node #1 --> #3)
	n.Update(X, Z)
	require.Equal(t, float64(0.5), n.PriorPNotFromAtoB(X, Z),
		"1 access to X out of 2 access total, so it should be 0.5 = 50%")

	// Train: from X to Z (node #1 --> #3) * 2
	n.Update(X, Z)
	n.Update(X, Z)
	require.Equal(t, float64(0.75), n.PriorPNotFromAtoB(X, Y),
		"1 access to Y out of 4 access total, so it should be 0.75 = 75%")

	require.Equal(t, float64(1), n.PriorPNotFromAtoB(Y, X),
		"0 access from Y out of 4 access total, so it should be 1.00 = 100%")
	require.Equal(t, float64(1), n.PriorPNotFromAtoB(Z, X),
		"0 access from Z out of 4 access total, so it should be 1.00 = 100%")
}
