package field_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"trailofbits.com/ml-dsa/mldsa/internal/field"
)

const (
	q = 8380417
)

func TestAddRandom(t *testing.T) {
	// Test addition of random pairs of field elements
	for i := 0; i < 1000; i++ {
		a := uint32(rand.Intn(int(q)))
		b := uint32(rand.Intn(int(q)))
		sum := (a + b) % q
		aF := field.NewFromReduced(a)
		bF := field.NewFromReduced(b)
		assert.Equal(t, sum, aF.Add(bF).Reduced())
	}
}

func TestAddSpecial(t *testing.T) {
	ns := []uint32{0, 1, q - 1}
	for i := 0; i < 1000; i++ {
		for _, a := range ns {
			for _, b := range ns {
				sum := (a + b) % q
				aF := field.NewFromReduced(a)
				bF := field.NewFromReduced(b)
				assert.Equal(t, sum, aF.Add(bF).Reduced())
			}
		}
	}
}

func TestSubRandom(t *testing.T) {
	// Test subtraction of random pairs of field elements
	for i := 0; i < 1000; i++ {
		a := uint32(rand.Intn(int(q)))
		b := uint32(rand.Intn(int(q)))
		diff := (a + q - b) % q
		aF := field.NewFromReduced(a)
		bF := field.NewFromReduced(b)
		assert.Equal(t, diff, aF.Sub(bF).Reduced())
	}
}

func TestNegRandom(t *testing.T) {
	// Test negation of random field elements
	for i := 0; i < 1000; i++ {
		a := uint32(rand.Intn(int(q)))
		aF := field.NewFromReduced(a)
		negated := (q - a) % q
		assert.Equal(t, negated, aF.Neg().Reduced())
	}
}

func TestSubSpecial(t *testing.T) {
	ns := []uint32{0, 1, q - 1}
	for i := 0; i < 1000; i++ {
		for _, a := range ns {
			for _, b := range ns {
				diff := (a + q - b) % q
				aF := field.NewFromReduced(a)
				bF := field.NewFromReduced(b)
				assert.Equal(t, diff, aF.Sub(bF).Reduced())
			}
		}
	}
}

func TestMulRandom(t *testing.T) {
	// Test multiplication of random pairs of field elements
	for i := 0; i < 1000; i++ {
		a := uint32(rand.Intn(int(q)))
		b := uint32(rand.Intn(int(q)))
		product := (uint64(a) * uint64(b)) % uint64(q)
		aF := field.NewFromReduced(a)
		bF := field.NewFromReduced(b)
		assert.Equal(t, uint32(product), aF.Mul(bF).Reduced())
	}
}

func TestMulSpecial(t *testing.T) {
	ns := []uint32{0, 1, 2, q - 1}
	for i := 0; i < 1000; i++ {
		for _, a := range ns {
			for _, b := range ns {
				product := (uint64(a) * uint64(b)) % uint64(q)
				aF := field.NewFromReduced(a)
				bF := field.NewFromReduced(b)
				assert.Equal(t, uint32(product), aF.Mul(bF).Reduced())
			}
		}
	}
}

func TestPower2RoundRandom(t *testing.T) {
	// Test Power2Round on random field elements
	for i := 0; i < 1000; i++ {
		a := uint32(rand.Intn(int(q)))
		aF := field.NewFromReduced(a)
		r1, r0 := aF.Power2Round()

		// Property checks
		assert.LessOrEqual(t, r0, int32(1<<12))
		assert.Greater(t, r0, -int32(1<<12))
		assert.Equal(t, a, uint32(int32(r1<<13)+r0)%q)

		// Compare to non-constant-time implementation
		expectedR0 := int32(a) % (1 << 13)
		if expectedR0 > 1<<12 {
			expectedR0 -= (1 << 13)
		}
		expectedR1 := (int32(a) - expectedR0) >> 13
		assert.Equal(t, expectedR1, r1)
		assert.Equal(t, expectedR0, r0)
	}
}

func TestPower2RoundBoundary(t *testing.T) {
	// r0 is exactly d/2
	a := uint32(42<<13) + (1 << 12) // 42 * 2^13 + 2^12
	aF := field.NewFromReduced(a)
	r1, r0 := aF.Power2Round()
	assert.Equal(t, int32(42), r1)
	assert.Equal(t, int32(1<<12), r0)

	// r0 is just above d/2
	a = uint32(42<<13) + (1 << 12) + 1 // 42 * 2^13 + 2^12 + 1
	aF = field.NewFromReduced(a)
	r1, r0 = aF.Power2Round()
	assert.Equal(t, int32(43), r1)
	assert.Equal(t, -int32(1<<12)+1, r0)
}

func TestInfinityNormRandom(t *testing.T) {
	for i := 0; i < 1000; i++ {
		a := uint32(rand.Intn(int(q)))
		aF := field.NewFromReduced(a)
		aNorm := aF.InfinityNorm()
		expected := min(a, q-a)
		assert.Equal(t, expected, aNorm)
	}
}

func TestInfinityNormBoundary(t *testing.T) {
	ns := []uint32{0, 1, q - 1, q / 2, q/2 + 1}
	for _, a := range ns {
		aF := field.NewFromReduced(a)
		aNorm := aF.InfinityNorm()
		expected := min(a, q-a)
		assert.Equal(t, expected, aNorm)
	}
}

func TestReduce(t *testing.T) {
	// Reduction should be a no-op for values in [0, q)
	for i := 0; i < 1000; i++ {
		a := uint32(rand.Intn(int(q)))
		aF := field.NewFromReduced(a)
		reduced := aF.Reduced()
		assert.Equal(t, a, reduced)
	}

	// Test reduction of special values
	ns := []uint32{0, 1, q - 1}
	for _, a := range ns {
		aF := field.NewFromReduced(a)
		reduced := aF.Reduced()
		assert.Equal(t, a, reduced)
	}
}

func TestSymmetric(t *testing.T) {
	for i := 0; i < 1000; i++ {
		a := uint32(rand.Intn(int(q)))
		aF := field.NewFromReduced(a)
		sym := aF.Symmetric()
		assert.LessOrEqual(t, sym, int32(q/2))
		assert.GreaterOrEqual(t, sym, -int32(q/2))
		assert.Equal(t, a, uint32(int32(sym)+int32(q))%q)
	}

	// Test symmetric representation of special values
	ns := []uint32{0, 1, q - 1, q/2 - 1, q / 2, q/2 + 1}
	for _, a := range ns {
		aF := field.NewFromReduced(a)
		sym := aF.Symmetric()
		assert.LessOrEqual(t, sym, int32(q/2))
		assert.GreaterOrEqual(t, sym, -int32(q/2))
		assert.Equal(t, a, uint32(int32(sym)+int32(q))%q)
	}
}

func TestDivConstTime32(t *testing.T) {
	moduli := []uint32{95232, 261888}
	x := uint32(1000) // number of sequential values to test; originally set to m
	for _, m := range moduli {
		for i := range x {
			w := i / m
			y := i % m
			wp, z := field.DivConstTime32(i, m)
			assert.Equal(t, y, z)
			assert.Equal(t, w, wp)
			wp, z = field.DivBarrett(i, m)
			assert.Equal(t, y, z)
			assert.Equal(t, w, wp)

			// Add m to i, expect i
			w = (i + m) / m
			wp, z = field.DivConstTime32(i+m, m)
			assert.Equal(t, y, z)
			assert.Equal(t, w, wp)
			wp, z = field.DivBarrett(i+m, m)
			assert.Equal(t, y, z)
			assert.Equal(t, w, wp)

			// Test division directly
			w = (i * m) / m
			wp, _ = field.DivConstTime32(i*m, m)
			assert.Equal(t, w, wp)
			wp, z = field.DivBarrett(i*m, m)
			assert.Equal(t, w, wp)
		}
	}
}
