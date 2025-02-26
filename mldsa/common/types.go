package common

// The coefficients for the polynomial coefficients are: [0, 2^((bitlen(q-1)-d) - 1]
// Since q and d are constant for all parameter sets, this results in:
// [0, 2^10 - 1] or [0, 1023] for ring coefficients. This fits in a int16.
//
// We use a signed integer here because some calculations result in negative numbers
// for intermediary values.

// Integers in the range [0, 2^10 - 1]
type RingCoeff int16
type RingElement [n]RingCoeff
type RingVector []RingElement

// Field elements are in the range [0, q-1]. This fits in a uint32.
type FieldElement uint32

// [n]FieldElement (integers mod q)
type NttElement [n]FieldElement

// [k]NttElement
type NttVector []NttElement

// [l]NttVector
type NttMatrix []NttVector

func Uint32ToFieldElement(x uint32) FieldElement {
	return FieldElement(x)
}

func NewRingElement() RingElement {
	var x RingElement
	for i := range n {
		x[i] = RingCoeff(int16(0))
	}
	return x
}

func NewRingVector(k uint8) RingVector {
	x := make([]RingElement, k)
	for i := range k {
		x[i] = NewRingElement()
	}
	return x
}

func NewNttElement() NttElement {
	x := make([]FieldElement, n)
	for i := range n {
		x[i] = FieldElement(0)
	}
	return NttElement(x)
}

func NewNttVector(k uint8) NttVector {
	x := make([]NttElement, k)
	for i := range k {
		x[i] = NewNttElement()
	}
	return x
}

func NewNttMatrix(k, l uint8) NttMatrix {
	x := make([]NttVector, k)
	for i := range k {
		x[i] = NewNttVector(l)
	}
	return x
}
