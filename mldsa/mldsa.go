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

/*
import (
	"trailofbits.com/ml-dsa/mldsa/common"
	"trailofbits.com/ml-dsa/mldsa/mldsa44"
	"trailofbits.com/ml-dsa/mldsa/mldsa65"
	"trailofbits.com/ml-dsa/mldsa/mldsa87"
)
*/

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
