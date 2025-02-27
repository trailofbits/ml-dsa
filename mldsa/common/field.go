package common

const (
	// These are consistent across all parameter sets:
	q = 8380417
	n = 256
	ζ = 1753
	d = 13

	// For field elements:
	barrettMultiplier = 8396807 // 2²³ * 2²³ / q
	barrettShift      = 46      // log₂(2²³ * 2²³)
)

// Reduce a field element once, mod q.
func FieldReduceOnce(a uint32) FieldElement {
	x := uint32(a - q)
	x += (x >> 31) * q
	return FieldElement(x)
}

// Add two field elements, mod q.
func FieldAdd(a, b FieldElement) FieldElement {
	x := uint32(a + b)
	return FieldReduceOnce(x)
}

// Subtract two field elements, mod q.
func FieldSub(a, b FieldElement) FieldElement {
	x := uint32(a - b + q)
	return FieldReduceOnce(x)
}

// Use barrett reduction to calculate a mod q without division.
func FieldReduce(a uint64) FieldElement {
	quotient := (a >> 23) * barrettMultiplier >> (barrettShift - 23)
	// Compute the remainder
	remainder := uint32(a - quotient*uint64(q))
	// Ensure result is fully reduced
	remainder += q & -((remainder - q) >> 31) // Add q back if negative
	return FieldReduceOnce(remainder)
}

// Multiply two field elements, mod q.
func FieldMul(a, b FieldElement) FieldElement {
	x := uint64(a) * uint64(b)
	return FieldReduce(x)
}
