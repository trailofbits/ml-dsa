// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mldsa implements the quantum-resistant module-lattice based digital
// signature algorithm ML-DSA (formerly known as Dilithium), as specified in [NIST FIPS 204].
//
// Most applications should use the ML-DSA-44 parameter set, as implemented by
// [SigningKey44] and [VerifyingKey44].
//
// [NIST FIPS 204]: https://doi.org/10.6028/NIST.FIPS.204
package mldsa

const (
	// We use seeds instead of expanded private keys
	SeedSize = 32

	// Public Key size (bytes) for ML-DSA-44
	PublicKeySize44 = 1312

	// Signature size (bytes) for ML-DSA-44
	SignatureSize44 = 2420

	// Public Key size (bytes) for ML-DSA-65
	PublicKeySize65 = 1952

	// Signature size (bytes) for ML-DSA-65
	SignatureSize65 = 3309

	// Public Key size (bytes) for ML-DSA-87
	PublicKeySize87 = 2592

	// Signature size (bytes) for ML-DSA-87
	SignatureSize87 = 4627
)

type SigningKey struct {
	seed [32]byte // ξ from the specification
	ρ    [32]byte // Rho is the public seed
	K    [32]byte
	tr   [64]byte
	t0   [k]common.RingElement
	t1   [k]common.RingElement
}

type VerifyingKey struct {
	ρ  [32]byte // Rho is the public seed
	t1 [k]common.RingElement
}
