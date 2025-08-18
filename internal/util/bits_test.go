package util_test

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"trailofbits.com/ml-dsa/internal/params"
	"trailofbits.com/ml-dsa/internal/ring"
	"trailofbits.com/ml-dsa/internal/util"
)

func TestSimpleBitPack(t *testing.T) {
	var w ring.Rz
	w[0] = 0x2ab
	w[1] = 0x3de
	packed := util.SimpleBitPack(w, 10)
	assert.Equal(t, byte(0xab), packed[0])
	assert.Equal(t, byte(0x7a), packed[1])
	assert.Equal(t, byte(0x0f), packed[2])
}

func TestSimpleBitPackUnpack(t *testing.T) {
	ks := []uint8{4, 6, 10}
	for _, k := range ks {
		var w ring.Rz
		for i := range len(w) {
			w[i] = int32(rand.IntN(1 << k))
		}
		packed := util.SimpleBitPack(w, k)
		unpacked := util.SimpleBitUnpack(packed, k)

		assert.Equal(t, w, unpacked)
	}
}

func TestBitPackUnpackClosed(t *testing.T) {
	for _, k := range []uint8{1, 2} {
		var w ring.Rz
		for i := 0; i < params.N; i++ {
			w[i] = int32((1 << k) - rand.IntN((2<<k)+1))
		}
		packed := util.BitPackClosed(w, k)
		unpacked, err := util.BitUnpackClosed(packed, k)
		assert.NoError(t, err)
		assert.Equal(t, w, unpacked)
	}
}

func TestBitPackUnpackClosedErr(t *testing.T) {
	for _, k := range []uint8{1, 2} {
		var w ring.Rz
		for i := 0; i < params.N; i++ {
			w[i] = int32((1 << k) - rand.IntN((2<<k)+1))
		}

		w[0] = int32((1 << k) + 1)

		packed := util.BitPackClosed(w, k)
		_, err := util.BitUnpackClosed(packed, k)
		assert.Error(t, err)
	}
}

// Test bit packing with the bounds used in T0 (2^d)
func TestBitPackT0(t *testing.T) {
	var w ring.Rz
	for i := 0; i < params.N; i++ {
		w[i] = int32((1 << (params.D - 1)))
	}

	w[0] = 42
	w[1] = -42
	packed := util.BitPack(w, params.D-1)

	assert.Equal(t, byte(0xd6), packed[0])
	assert.Equal(t, byte(0x4f), packed[1])
	assert.Equal(t, byte(0x05), packed[2])
	assert.Equal(t, byte(0x05), packed[2])
	assert.Equal(t, byte(0x02), packed[3])

	unpacked := util.BitUnpack(packed, params.D-1)
	assert.Equal(t, w, unpacked)
}

func TestHintPacking(t *testing.T) {
	t.Run("TestHintPackingForK4", func(t *testing.T) {
		hintPackingForK(t, uint8(4), uint8(80))
	})
	t.Run("TestHintPackingForK6", func(t *testing.T) {
		hintPackingForK(t, uint8(6), uint8(55))
	})
	t.Run("TestHintPackingForK8", func(t *testing.T) {
		hintPackingForK(t, uint8(8), uint8(75))
	})
}

func hintPackingForK(t *testing.T, k, omega uint8) {
	vec := make([]ring.R2, k)
	for i := range k {
		x := uint8(0)
		for j := range 256 {
			// Ensure at most omega are nonzero
			if x%64 == k {
				vec[i][j] = 1
			}
			x++
		}
	}
	packed := util.HintBitPack(k, omega, vec)
	unpacked, err := util.HintBitUnpack(k, omega, packed)
	if err != nil {
		panic(err)
	}

	for i := range k {
		x := uint8(0)
		for j := range 256 {
			expected := uint8(0)
			if x%64 == k {
				expected = 1
			}
			assert.Equal(t, expected, unpacked[i][j])
			x++
		}
	}
}
