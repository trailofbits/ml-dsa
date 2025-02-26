package common

// Since ring coefficients are between [-1024, 1023], we can use bitwise operators
// and avoid a more expensive Barrett reduction, while also avoiding side-channels
func CoeffReduceOnce(a int16) RingCoeff {
	// If the highest bit is set, this is a signed number
	sign := int16(1 & (a >> 15))
	// If sign == 1, mask = -1; else, mask = 0
	mask := -sign
	// Reduce mod 1024 (equivalent to an AND mask with 1023)
	reduced := int16(a & 1023)
	// Constant-time conditional swap
	left := int16((-1024 | reduced) & mask)
	right := int16(reduced & ^mask)
	with_sign := left ^ right
	return RingCoeff(with_sign)
}

// TODO - get rid of this garbage
func CoeffReduceUint32(a uint32) RingCoeff {
	return CoeffReduceOnce(int16(a & 0xffff))
}
func CoeffReduceInt32(a int32) RingCoeff {
	return CoeffReduceOnce(int16(a & 0xffff))
}
func Int16ToRingCoeff(a int16) RingCoeff {
	return RingCoeff(a)
}

// Adding coefficients, mod q.
func CoeffAdd(a, b RingCoeff) RingCoeff {
	x := int16(a + b)
	return CoeffReduceOnce(x)
}

// Subtracting coefficients, mod q.
func CoeffSub(a, b RingCoeff) RingCoeff {
	x := int16(a - b)
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
	for i := range k {
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
