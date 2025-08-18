package ring_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trailofbits/ml-dsa/internal/field"
	"github.com/trailofbits/ml-dsa/internal/params"
	"github.com/trailofbits/ml-dsa/internal/ring"
)

// helper to build a deterministic symmetric vector covering negatives/positives
func makeSymVec() ring.Rz {
	var z ring.Rz
	for i := range z {
		v := int32((i % 11) - 5) // values in [-5, 5]
		z[i] = v
	}
	return z
}

func TestFromSymmetricRoundTrip(t *testing.T) {
	z := makeSymVec()
	a := ring.FromSymmetric(z)
	back := a.Symmetric()
	assert.Equal(t, z, back)
}

func TestAddSubNegProperties(t *testing.T) {
	// build two random-ish inputs with small magnitudes
	var z1, z2 ring.Rz
	for i := range z1 {
		z1[i] = int32(rand.Intn(11) - 5)
		z2[i] = int32(rand.Intn(11) - 5)
	}
	a := ring.FromSymmetric(z1)
	b := ring.FromSymmetric(z2)

	// a + b - b == a
	sum := a.Add(b)
	diff := sum.Sub(b)
	assert.Equal(t, a.Symmetric(), diff.Symmetric())

	// a + (-a) == 0
	zero := a.Add(a.Neg())
	for i := range zero {
		assert.Equal(t, int32(0), zero[i].Symmetric())
	}
}

func TestInfinityNorm(t *testing.T) {
	qHalf := int32(params.Q / 2)
	var z ring.Rz
	// include a range of values, with maximum absolute value q/2
	z[0] = -1
	z[1] = 0
	z[2] = 12345
	z[3] = -6789
	z[4] = qHalf
	z[5] = -qHalf
	a := ring.FromSymmetric(z)
	assert.Equal(t, uint32(qHalf), a.InfinityNorm())
}

func TestPower2RoundRoundTrip(t *testing.T) {
	// construct coefficients with a spread across small values to avoid overflow in reconstruction math
	z := makeSymVec()
	a := ring.FromSymmetric(z)
	r1, r0 := a.Power2Round()

	// For each coefficient, verify a = r1*2^d + r0 (mod q) by comparing symmetric reps
	d := int32(params.D)
	q := int64(params.Q)
	back := a.Symmetric()
	for i := range back {
		v := (int64(r1[i])<<d + int64(r0[i])) % q
		if v < 0 {
			v += q
		}
		// map to symmetric in [-q/2, q/2]
		sym := int32(v)
		if v > int64(params.Q/2) {
			sym -= int32(params.Q)
		}
		assert.Equal(t, sym, back[i])
	}
}

func TestHighLowBitsRoundTrip(t *testing.T) {
	z := makeSymVec()
	a := ring.FromSymmetric(z)
	gamma2 := params.MLDSA44Cfg.Gamma2

	high := a.HighBits(gamma2)
	low := a.LowBits(gamma2)

	q := int64(params.Q)
	back := a.Symmetric()
	for i := range back {
		v := (int64(high[i])*int64(2*gamma2) + int64(low[i])) % q
		if v < 0 {
			v += q
		}
		sym := int32(v)
		if v > int64(params.Q/2) {
			sym -= int32(params.Q)
		}
		assert.Equal(t, sym, back[i])
	}
}

func TestVectorHelpers(t *testing.T) {
	// InfinityNormVec: max across vectors
	z1 := makeSymVec()
	z2 := makeSymVec()
	// amplify one entry in z2 to increase its norm
	z2[7] = int32(params.Q / 2)

	a1 := ring.FromSymmetric(z1)
	a2 := ring.FromSymmetric(z2)

	max := ring.InfinityNormVec([]ring.Rq{a1, a2})
	assert.Equal(t, uint32(params.Q/2), max)

	// HighBitsVec should equal elementwise HighBits
	gamma2 := params.MLDSA65Cfg.Gamma2
	hv := ring.HighBitsVec([]ring.Rq{a1, a2}, gamma2)
	assert.Equal(t, a1.HighBits(gamma2), hv[0])
	assert.Equal(t, a2.HighBits(gamma2), hv[1])

	// ScalarMul sanity: multiply by 0 yields 0, by 1 yields same
	zero := field.NewFromReduced(0)
	one := field.NewFromReduced(1)
	zeroVec := a1.ScalarMul(zero)
	oneVec := a1.ScalarMul(one)
	for i := range zeroVec {
		assert.Equal(t, int32(0), zeroVec[i].Symmetric())
		assert.Equal(t, a1[i].Symmetric(), oneVec[i].Symmetric())
	}
}
