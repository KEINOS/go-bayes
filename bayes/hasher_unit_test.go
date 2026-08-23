package bayes

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPredictor_itemID_canonicalBytes(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value any
		want  []byte
	}{
		"bool":     {true, []byte{0x01, 0x01, 0x01}},
		"string":   {"", []byte{0x01, 0x02}},
		"int":      {int(-1), append([]byte{0x01, 0x03}, eightBytes(math.MaxUint64)...)},
		"int16":    {int16(-2), append([]byte{0x01, 0x04}, eightBytes(math.MaxUint64-1)...)},
		"int32":    {int32(-3), append([]byte{0x01, 0x05}, eightBytes(math.MaxUint64-2)...)},
		"int64":    {int64(-4), append([]byte{0x01, 0x06}, eightBytes(math.MaxUint64-3)...)},
		"uint":     {uint(1), append([]byte{0x01, 0x07}, eightBytes(1)...)},
		"uint16":   {uint16(2), append([]byte{0x01, 0x08}, eightBytes(2)...)},
		"uint32":   {uint32(3), append([]byte{0x01, 0x09}, eightBytes(3)...)},
		"uint64":   {uint64(0), append([]byte{0x01, 0x0a}, eightBytes(0)...)},
		"float32":  {float32(1.5), append([]byte{0x01, 0x0b}, eightBytes(uint64(math.Float32bits(1.5)))...)},
		"float64":  {float64(-1.5), append([]byte{0x01, 0x0c}, eightBytes(math.Float64bits(-1.5))...)},
		"NaN bits": {math.Float64frombits(0x7ff8000000000001), append([]byte{0x01, 0x0c}, eightBytes(0x7ff8000000000001)...)},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hasher := &stubHasher{got: nil, name: "recording", out: 7}
			predictor := &Predictor{
				predictor: nil,
				classes:   nil,
				storage:   UnknownStorage,
				scopeID:   0,
				hasher:    hasher,
				scratch:   codecScratch{bytes: nil, ids: nil},
			}

			actual, err := predictor.itemID(test.value)

			require.NoError(t, err)
			require.Equal(t, uint64(7), actual)
			require.Equal(t, [][]byte{test.want}, hasher.got)
		})
	}
}

func TestPredictor_itemID_typeSeparation(t *testing.T) {
	t.Parallel()

	predictor := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    NewDefaultHasher(),
		scratch:   codecScratch{bytes: nil, ids: nil},
	}
	values := []any{true, int(1), uint64(1), float64(1)}
	ids := make(map[uint64]struct{}, len(values))

	for _, value := range values {
		id, err := predictor.itemID(value)
		require.NoError(t, err)

		ids[id] = struct{}{}
	}

	require.Len(t, ids, len(values))
}

func TestPredictor_itemID_goldenXXHash3(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value any
		want  uint64
	}{
		"bool":    {true, 0x0236d86862625904},
		"string":  {"go-bayes", 0xbeec91d1de3c69c1},
		"int":     {int(-1), 0x17e127df5c095aad},
		"int16":   {int16(-1), 0x9250506888362830},
		"int32":   {int32(-1), 0x25ce7212570a7d02},
		"int64":   {int64(-1), 0x6e681cc045aed1b8},
		"uint":    {uint(1), 0x47b1c3c796eda90c},
		"uint16":  {uint16(1), 0xe459ab3798769edb},
		"uint32":  {uint32(1), 0xaa299ab43f7339b2},
		"uint64":  {uint64(1), 0x640d809f0ccb9099},
		"float32": {float32(1.5), 0x18d04825981e934a},
		"float64": {float64(1.5), 0xf08af89efbc7f255},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			predictor := &Predictor{
				predictor: nil,
				classes:   nil,
				storage:   UnknownStorage,
				scopeID:   0,
				hasher:    NewXXHash3Hasher(),
				scratch:   codecScratch{bytes: nil, ids: nil},
			}

			actual, err := predictor.itemID(test.value)
			require.NoError(t, err)
			require.Equal(t, test.want, actual)
		})
	}
}

func TestPredictor_contextID_canonicalBytes(t *testing.T) {
	t.Parallel()

	hasher := &stubHasher{got: nil, name: "recording", out: 9}
	predictor := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    hasher,
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	require.Equal(t, uint64(9), predictor.contextID(nil))
	require.Equal(t, []byte{0x02, 0x00}, hasher.got[0])

	require.Equal(t, uint64(9), predictor.contextID([]uint64{1, math.MaxUint64}))

	want := append([]byte{0x02, 0x02}, eightBytes(1)...)
	want = append(want, eightBytes(math.MaxUint64)...)
	require.Equal(t, want, hasher.got[1])

	manyIDs := make([]uint64, 128)
	predictor.contextID(manyIDs)
	require.Equal(t, []byte{contextDomain, 0x80, 0x01}, hasher.got[2][:3])
}

func TestPredictor_HashTrans_isDeterministicAndOrderSensitive(t *testing.T) {
	t.Parallel()

	predictor := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    NewDefaultHasher(),
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	forward, err := predictor.HashTrans("a", int(2), true)
	require.NoError(t, err)
	repeated, err := predictor.HashTrans("a", int(2), true)
	require.NoError(t, err)
	reversed, err := predictor.HashTrans(true, int(2), "a")
	require.NoError(t, err)
	empty, err := predictor.HashTrans()
	require.NoError(t, err)

	require.Equal(t, forward, repeated)
	require.NotEqual(t, forward, reversed)
	require.NotZero(t, empty)
}

func eightBytes(value uint64) []byte {
	encoded := make([]byte, 8)
	binary.BigEndian.PutUint64(encoded, value)

	return encoded
}
