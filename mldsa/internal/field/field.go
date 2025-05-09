// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package field implements arithmetic in the field Z_q, where q = 8380417.
package field

import (
	"crypto/subtle"

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
	// mod+/- is defined in the spec to be in the range
	// -ceil(m/2) < mod <= floor(m/2)
	// For even moduli, this means that m/2 is a canonical representative.
	// TODO: Bit-twiddle this to make constant-time
	if r0 > (1 << (d - 1)) {
		r0 -= (1 << d)
	}
	r1 := (int32(a.reduced) - r0) >> d
	return r1, r0
}

// Decompose (Algorithm 36) decomposes a field element x into two components:
// (r1, r0) such that x = r1 * (2 * gamma2) + r0 (mod q)
// -gamma2 < r0 <= gamma2
// 0 <= r1 < (q-1)/(2 * gamma2)
func (a T) Decompose(gamma2 uint32) (r1 int32, r0 int32) {
	// TODO - make this constant-time
	rPlus := int32(a.Reduced())
	r0 = rPlus % int32(2*gamma2)
	if r0 > int32(gamma2) {
		r0 -= 2 * int32(gamma2)
	}
	if int32(rPlus)-r0 == int32(q-1) {
		r1 = 0
		r0 = r0 - 1
	} else {
		r1 = (rPlus - r0) / int32(2*gamma2)
	}
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
