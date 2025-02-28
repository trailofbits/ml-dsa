package common

func CoeffReduceOnce(a uint32) RingCoeff {
	x := uint32(a - q)
	x += (x >> 31) * q
	return RingCoeff(x)
}

// Adding coefficients, mod q.
func CoeffAdd(a, b RingCoeff) RingCoeff {
	x := uint32(a + b)
	return CoeffReduceOnce(x)
}

// Subtracting coefficients, mod q.
func CoeffSub(a, b RingCoeff) RingCoeff {
	x := uint32(a - b + q)
	return CoeffReduceOnce(x)
}

// Add two Ring elements ([]int16)
func RingAdd(a, b RingElement) (s RingElement) {
	for i := range s {
		s[i] = CoeffAdd(a[i], b[i])
	}
	return s
}

// Subtract two Ring elements ([]int16)
func RingSub(a, b RingElement) (s RingElement) {
	for i := range s {
		s[i] = CoeffSub(a[i], b[i])
	}
	return s
}

func RingPower2Round(k uint8, r RingElement) (RingElement, RingElement) {
	var r1, r0 RingElement
	for i := range 256 {
		round0, round1 := Power2Round(uint32(r[i]))
		r1[i], r0[i] = CoeffReduceOnce(round0), CoeffReduceOnce(round1)
	}
	return r1, r0
}

func RingVecPower2Round(k uint8, r RingVector) (RingVector, RingVector) {
	r1 := NewRingVector(k)
	r0 := NewRingVector(k)
	for i := range k {
		round1, round0 := RingPower2Round(k, r[i])
		r1[i] = round1
		r0[i] = round0
	}
	return r1, r0
}

func RingVectorAdd(k uint8, a RingVector, b RingVector) RingVector {
	c := NewRingVector(k)
	for i := range k {
		c[i] = RingAdd(a[i], b[i])
	}
	return c
}
