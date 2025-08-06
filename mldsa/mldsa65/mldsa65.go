// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mldsa65 implements the ML-DSA-65 parameter set of the ML-DSA algorithm.
//
// See [FIPS 204](https://nvlpubs.nist.gov/nistpubs/fips/nist.fips.204.pdf) for more details.
package mldsa65

import (
	"crypto"
	"io"

	"trailofbits.com/ml-dsa/mldsa/internal"
	"trailofbits.com/ml-dsa/mldsa/internal/params"
)

// Package mldsa65 implements the ML-DSA-65 parameter set of the ML-DSA algorithm.

// PublicKey is the type of ML-DSA public keys. Implements crypto.PublicKey.
type PublicKey internal.VerifyingKey

// PrivateKey is the type of ML-DSA private keys. It implements crypto.Signer.
type PrivateKey internal.SigningKey

// GenerateKeyPair generates a key pair for the ML-DSA algorithm.
func GenerateKeyPair(rng io.Reader) (*PublicKey, *PrivateKey, error) {
	sk, pk, err := internal.GenerateKeyPair(params.MLDSA65Cfg, rng)

	if err != nil {
		return nil, nil, err
	}
	return (*PublicKey)(pk), (*PrivateKey)(sk), nil
}

// Public returns the public key corresponding to the ML-DSA private key.
func (sk *PrivateKey) Public() crypto.PublicKey {
	return (*internal.SigningKey)(sk).Public()
}
