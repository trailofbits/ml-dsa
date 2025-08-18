// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ring implements arithmetic in the degree-256 cyclotomic polynomial ring R_q
// where q = 8380417.
package ring

import (
	"github.com/trailofbits/ml-dsa/internal/field"
	"github.com/trailofbits/ml-dsa/internal/params"
)

const n = params.N

// TODO -  consider specializing types to smaller value bounds

// Rq is the type of cyclotomic ring elements R_q, where q = 8380417.
type Rq [n]field.T

// Tq is the type of NTT ring elements T_q, where q = 8380417.
type Tq [n]field.T

// Rz is the type of cyclotomic ring elements R_Z, where Z is the integers.
type Rz [n]int32

// R2 is the type cyclotomic ring elements R_2, over the booleans.
type R2 [n]uint8

/*
// The type of low-order bits from Power2Round and Compress.
// Values are in the range -2^(d-1) x < <= 2^(d-1).
type T0 [n]int32

// The type of high-order bits from Power2Round and Decompress
// In practice these are 10-bit integers.
type T1 [n]uint16

// The type of secret noise. Values in -eta < x <= eta.
type S [n]int8

// The type of hints used in the compression algorithm. Values in {0,1}
type Hint [n]uint8
*/

// Consider making generic over a base ring

// Add two Ring elements
func (a Rq) Add(b Rq) Rq {
	var s Rq
	for i := range s {
		s[i] = a[i].Add(b[i])
	}
	return s
}

// Subtract two Ring elements
func (a Rq) Sub(b Rq) Rq {
	var s Rq
	for i := range s {
		s[i] = a[i].Sub(b[i])
	}
	return s
}

func (a Rq) Neg() Rq {
	var s Rq
	for i := range s {
		s[i] = a[i].Neg()
	}
	return s
}

// Coefficient-wise decomposition using Power2Round
func (a Rq) Power2Round() (r1 Rz, r0 Rz) {
	for i := range n {
		r1[i], r0[i] = a[i].Power2Round()
	}
	return r1, r0
}

// Coefficient-wise decomposition using Decompose
func (a Rq) HighBits(gamma2 uint32) (r1 Rz) {
	for i := range n {
		r1[i] = a[i].HighBits(gamma2)
	}
	return r1
}

func HighBitsVec(a []Rq, gamma2 uint32) (r1 []Rz) {
	r1 = make([]Rz, len(a))
	for i := range a {
		r1[i] = a[i].HighBits(gamma2)
	}
	return r1
}

// Coefficient-wise decomposition using Decompose
func (a Rq) LowBits(gamma2 uint32) (r0 Rz) {
	for i := range n {
		_, r0[i] = a[i].Decompose(gamma2)
	}
	return r0
}

func LowBitsVec(a []Rq, gamma2 uint32) (r0 []Rz) {
	r0 = make([]Rz, len(a))
	for i := range a {
		r0[i] = a[i].LowBits(gamma2)
	}
	return r0
}

func (a Rq) InfinityNorm() (norm uint32) {
	for i := range n {
		norm = max(norm, a[i].InfinityNorm())
	}
	return norm
}

func InfinityNormVec(a []Rq) (norm uint32) {
	for i := range a {
		norm = max(norm, a[i].InfinityNorm())
	}
	return norm
}

func (a Rq) Symmetric() (z Rz) {
	for i := range a {
		z[i] = a[i].Symmetric()
	}
	return z
}

func (a Rq) ScalarMul(c field.T) Rq {
	var s Rq
	for i := range s {
		s[i] = a[i].Mul(c)
	}
	return s
}

func FromSymmetric(z Rz) (a Rq) {
	for i := range z {
		a[i] = field.NewFromSymmetric(z[i])
	}
	return a
}

func FromSymmetricVec(z []Rz) []Rq {
	v := make([]Rq, len(z))
	for i := range z {
		v[i] = FromSymmetric(z[i])
	}
	return v
}

func (a Tq) Add(b Tq) Tq {
	var s Tq
	for i := range s {
		s[i] = a[i].Add(b[i])
	}
	return s
}

func (a Tq) Sub(b Tq) Tq {
	var s Tq
	for i := range s {
		s[i] = a[i].Sub(b[i])
	}
	return s
}

func (a Tq) Mul(b Tq) Tq {
	var s Tq
	for i := range s {
		s[i] = a[i].Mul(b[i])
	}
	return s
}
