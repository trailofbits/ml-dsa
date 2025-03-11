// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package common implements the generic underlying algorithms from [NIST FIPS 204].
//
// The types.go file defines the following data types:
//
// POLYNOMIAL:
// - RingCoeff (integers mod q)
// - RingElement ([n]RingCoeff)
// - RingVector ([k]RingVector, a.k.a. [k][n]RingCoeff)
//
// NTT DOMAIN:
// - FieldElement (integers mod q, in the NTT domain)
// - NttElement ([n]FieldElement)
// - NttVector ([]NttElement)
// - NttMatrix ([]NttVector)
package common

// Field elements are in the range [0, q-1]. This fits in a uint32.
type FieldElement uint32

// [n]FieldElement (integers mod q)
type NttElement [n]FieldElement

// [k]NttElement
type NttVector []NttElement

// [l]NttVector
type NttMatrix []NttVector

// We can simplify the code a lot and use uint32 for both components
type RingCoeff uint32
type RingElement [n]RingCoeff
type RingVector []RingElement

func Uint32ToFieldElement(x uint32) FieldElement {
	return FieldElement(x)
}

func NewRingElement() RingElement {
	var x RingElement
	for i := range n {
		x[i] = RingCoeff(0)
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
