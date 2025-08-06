// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package field implements arithmetic in the field Z_q, where q = 8380417.
package field

import (
	"crypto/subtle"
	"math/bits"

	"trailofbits.com/ml-dsa/mldsa/internal/params"
)

// T is the type of field elements F_q, where q = 8380417.
type T struct {
	reduced uint32 // TODO: implement Montgomery representation
}

const (
	q                 = params.Q
	d                 = params.D
	barrettMultiplier = 8396807 // 2²³ * 2²³ / q
	barrettShift      = 46      // log₂(2²³ * 2²³)
)

// NewFromReduced creates a new field element from a reduced non-negative value.
// Panic if the value is not reduced (0 <= reduced < q).
func NewFromReduced(reduced uint32) T {
	if reduced >= q {
		panic("NewFromReduced: value out of range")
	}
	return T{reduced: reduced}
}

// NewFromSymmetric creates a new field element from a reduced signed value.
func NewFromSymmetric(x int32) T {
	// TODO: remove this panic once we are sure that all inputs are reduced, to avoid timing leaks
	if x < -int32(q/2) || x > int32(q/2) {
		panic("NewFromSymmetric: value out of range")
	}
	y := uint32(x) + (uint32(x)>>31)*q
	return NewFromReduced(y)
}

// Reduce a field element once, mod q.
func reduceOnce(a uint32) T {
	x := uint32(a - q)
	x += (x >> 31) * q
	// TODO: remove this panic once we are sure that all inputs are reduced
	return NewFromReduced(x)
	//return Element(x)
}

// Add two field elements, mod q.
func (a T) Add(b T) T {
	return reduceOnce(a.reduced + b.reduced)
}

// Subtract two field elements, mod q.
func (a T) Sub(b T) T {
	return reduceOnce(a.reduced - b.reduced + q)
}

// Negate a field element, mod q.
func (a T) Neg() T {
	return reduceOnce(q - a.reduced)
}

// Multiply two field elements, mod q.
func (a T) Mul(b T) T {
	// TODO - Use an efficient constant-time implementation
	product := uint64(a.reduced) * uint64(b.reduced)
	return T{uint32(product % q)}
}

// Power2Round (Algorithm 35) decomposes a field element x into two components:
// (r1, r0) such that x = r1 * 2^d + r0 (mod q)
// and r0 \in (-2^(d-1), 2^(d-1)].
// r1 is then in the range [0, q/2^d), which is 10 bits
func (a T) Power2Round() (int32, int32) {
	r0 := int32(a.reduced) & ((1 << d) - 1)

	// Constant-time conditional subtraction to avoid timing side-channels
	// Check if r0 > 2^(d-1) without branching
	threshold := int32(1 << (d - 1))

	// Create a mask: all 1s if r0 > threshold, all 0s otherwise
	// We want the mask to be -1 when r0 > threshold, 0 otherwise
	// Since we want r0 > threshold (not >=), we check if (threshold - r0) < 0
	mask := (threshold - r0) >> 31

	// Conditionally subtract 2^d from r0 using the mask
	// If mask is -1 (all 1s), we subtract 2^d
	// If mask is 0, we subtract 0
	adjustment := mask & (1 << d)
	r0 -= adjustment

	r1 := (int32(a.reduced) - r0) >> d
	return r1, r0
}

// Decompose (Algorithm 36) decomposes a field element x into two components:
// (r1, r0) such that x = r1 * (2 * gamma2) + r0 (mod q)
// -gamma2 < r0 <= gamma2
// 0 <= r1 < (q-1)/(2 * gamma2)
func (a T) Decompose(gamma2 uint32) (r1 int32, r0 int32) {

	rPlus := int32(a.Reduced())
	twoGamma2 := int32(2 * gamma2)
	gamma2Int := int32(gamma2)

	// Constant-time modulo: r0 = rPlus % (2*gamma2)
	tmp, _ := divBarrettSigned(rPlus, twoGamma2)
	r0 = rPlus - (tmp*twoGamma2)

	// Constant-time conditional: if r0 > gamma2, subtract 2*gamma2
	// Create mask: -1 if r0 > gamma2, 0 otherwise
	mask1 := (gamma2Int - r0) >> 31
	adjustment1 := mask1 & twoGamma2
	r0 -= adjustment1

	// Constant-time conditional for the special case
	// if rPlus - r0 == q-1, then r1 = 0 and r0 = r0 - 1
	// otherwise r1 = (rPlus - r0) / (2*gamma2)

	diff := rPlus - r0
	qMinus1 := int32(q - 1)

	// Create mask: -1 if diff == q-1, 0 otherwise
	// We use the fact that (a-b) | (b-a) has MSB set iff a != b
	temp := (diff - qMinus1) | (qMinus1 - diff)
	mask2 := ^(temp >> 31) // Invert to get -1 when equal, 0 when not equal

	// Use a constant-time function to compute integer division
	quotient, _ := DivBarrett(uint32(diff), uint32(twoGamma2))

	// Calculate both possible values
	normalR1 := int32(quotient)
	specialR1 := int32(0)
	normalR0 := r0
	specialR0 := r0 - 1

	// Select between normal and special case using mask2
	r1 = (mask2 & specialR1) | (^mask2 & normalR1)
	r0 = (mask2 & specialR0) | (^mask2 & normalR0)

	return r1, r0
}

// Algorithm 37
func (a T) HighBits(gamma2 uint32) int32 {
	r1, _ := a.Decompose(gamma2)
	return r1
}

// InfinityNorm computes the absolute value of the
// signed symmetric representation of a field element.
func (a T) InfinityNorm() uint32 {
	return min(a.reduced, q-a.reduced)
}

// Reduced returns the [0, q-1] representative
func (a T) Reduced() uint32 {
	return a.reduced
}

// Symmetric returns the [-q/2, q/2]
func (a T) Symmetric() int32 {
	mask := subtle.ConstantTimeLessOrEq(q>>1+1, int(a.reduced))
	return int32(a.reduced) - (int32(mask) * int32(q))
}

// Algorithm 15
// Input: eta, byte
// Output: integer between -eta and eta or nil for rejection
func FromHalfByte(eta int, b byte) *T {
	var r T
	if eta == 2 && b < 15 {
		r = NewFromSymmetric(2 - (int32(b) % 5))
		return &r
	}
	if eta == 4 && b < 9 {
		r = NewFromSymmetric(4 - int32(b))
		return &r
	}
	return nil
}

// Algorithm 14
// Inputs: 3 bytes
// Output: integer mod q or nil for rejection
func FromThreeBytes(b0, b1, b2 byte) *T {
	var r T
	bp2 := uint32(b2 & 0x7f)
	z := (bp2 << 16) | uint32(b1)<<8 | uint32(b0)

	// if q >= z, return an error
	invalid := (q - (z + 1)) >> 31
	if invalid != 0 {
		return nil
	}
	r = reduceOnce(z)
	return &r
}

// Alternate to integer division by using Barrett Reduction
// Calculates (n/d, n%d) given (n, d)
func DivBarrett(numerator, denominator uint32) (uint32, uint32) {
    // Since d is always 2 * gamma2, we can precompute (2^64 / d) and use it
    var reciprocal uint64
    switch denominator {
    case 95232:
        reciprocal = 193703209779376
    case 261888:
        reciprocal = 70368744177664
    case 190464:
        reciprocal = 96851604889688
    case 523776:
        reciprocal = 35184372088832
    default:
        // Fallback to slow division
		return DivConstTime32(numerator, denominator)
    }
    
    // Barrett reduction
    hi, _ := bits.Mul64(uint64(numerator), reciprocal)
    quo := uint32(hi)
    r := numerator - quo * denominator
    
    // Two correction steps using bits.Sub32 (constant-time)
    for i := 0; i < 2; i++ {
        newR, borrow := bits.Sub32(r, denominator, 0)
        correction := borrow ^ 1  // 1 if r >= d, 0 if r < d
        mask := uint32(-correction)
        quo += mask & 1
        r ^= mask & (newR ^ r)  // Conditional swap using XOR
    }
    
    return quo, r
}

// For signed integers:
func divBarrettSigned(numerator, denominator int32) (int32, int32) {
	un := uint32(numerator)
	ud := uint32(denominator)
	quo, r := DivBarrett(un, ud)
	return int32(quo), int32(r)
}

// Modification of https://en.wikipedia.org/wiki/Division_algorithm#Integer_division_(unsigned)_with_remainder
// Except with branchless, conditional swaps
//
// This function works for arbitrary values for d, but is slower than DivBarrett.
func DivConstTime32(n uint32, d uint32) (uint32, uint32) {
	quotient := uint32(0)
	R := uint32(0)

	// We are dealing with 32-bit integers, so we iterate 32 times
	b := uint32(32)
	i := b
	for range b {
		i--
		R <<= 1

		// R(0) := N(i)
		R |= ((n >> i) & 1)

		// swap from Sub32() will look like this:
		// if remainder > d,  swap == 0
		// if remainder == d, swap == 0
		// if remainder < d,  swap == 1
		Rprime, swap := bits.Sub32(R, d, 0)

		// invert logic of sub32 for conditional swap
		swap ^= 1
		/*
			Desired:
				if R > D  then swap = 1
				if R == D then swap = 1
				if R < D  then swap = 0
		*/

		// Qprime := Q
		// Qprime(i) := 1
		Qprime := quotient
		Qprime |= (1 << i)

		// Conditional swap:
		mask := uint32(-swap)
		R ^= ((Rprime ^ R) & mask)
		quotient ^= ((Qprime ^ quotient) & mask)
	}
	return quotient, R
}
