// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mldsa87 implements the ML-DSA-87 parameter set of the ML-DSA algorithm.
// See [FIPS 204] for more details.
//
// [FIPS 204]: https://nvlpubs.nist.gov/nistpubs/fips/nist.fips.204.pdf
package mldsa87

import (
	"crypto"
	"io"

	internal "trailofbits.com/ml-dsa/internal"
	"trailofbits.com/ml-dsa/internal/params"
	options "trailofbits.com/ml-dsa/options"
)

// Package mldsa87 implements the ML-DSA-87 parameter set of the ML-DSA algorithm.

// PublicKey is the type of ML-DSA public keys. Implements [crypto.PublicKey].
type PublicKey struct {
	pk internal.VerifyingKey
}

// PrivateKey is the type of ML-DSA private keys. It implements [crypto.Signer].
type PrivateKey struct {
	sk internal.SigningKey
}

// GenerateKeyPair generates a key pair for the ML-DSA algorithm.
// If rng is nil, [crypto/rand] is used.
//
// [crypto/rand]: https://pkg.go.dev/crypto/rand
func GenerateKeyPair(rng io.Reader) (*PublicKey, *PrivateKey, error) {
	sk, pk, err := internal.GenerateKeyPair(params.MLDSA87Cfg, rng)

	if err != nil {
		return nil, nil, err
	}
	return &PublicKey{*pk}, &PrivateKey{*sk}, nil
}

func (pub *PublicKey) Verify(msg, sig []byte) bool {
	return pub.pk.Verify(msg, sig, nil)
}

func (pub *PublicKey) VerifyWithOptions(msg, sig []byte, opts *options.Options) bool {
	return pub.pk.Verify(msg, sig, opts)
}

// Public returns the public key corresponding to the ML-DSA private key.
func (priv *PrivateKey) Public() crypto.PublicKey {
	return priv.sk.Public()
}

// Returns the 2592-byte public key as defined in FIPS 204.
func (pub *PublicKey) Bytes() []byte {
	return pub.pk.Bytes()
}

// Decodes a 2592-byte public key as defined in FIPS 204.
func PublicKeyFromBytes(bytes []byte) (*PublicKey, error) {
	pk, err := internal.PkDecode(params.MLDSA87Cfg, bytes)
	if err != nil {
		return nil, err
	}
	return &PublicKey{*pk}, nil
}

// Signs the given message with priv. If rand is nil, [crypto/rand] is used.
// For deterministic signing, you may explicitly pass in a reader that always returns zeros.
//
// Only pure ML-DSA is supported. opts.HashFuc() must return 0.
// opts may be nil, in which case empty context is used.
//
// [crypto/rand]: https://pkg.go.dev/crypto/rand
func (priv *PrivateKey) Sign(rand io.Reader, message []byte, opts crypto.SignerOpts) ([]byte, error) {
	return priv.sk.Sign(rand, message, opts)
}

// Returns the seed used to generate the private key.
// This is the recommended way to store the private key.
// Note that this is not the fully expanded private key defined in FIPS 204.
// For compatibility with other implementations, you may need to use EncodeExpanded instead.
func (priv *PrivateKey) Seed() ([]byte, error) {
	return priv.sk.Bytes()
}

// Reads a private key from a 32-byte seed.
// Returns an error if the seed is not 32 bytes.
func PrivateKeyFromSeed(seed []byte) (*PrivateKey, error) {
	sk, err := internal.FromSeed(params.MLDSA87Cfg, seed)
	if err != nil {
		return nil, err
	}
	return &PrivateKey{*sk}, nil
}

// Returns the expanded 4896-byte private key defined in FIPS 204.
// This is not the recommended way to store the private key, unless necessary for compatibility.
func (priv *PrivateKey) EncodeExpanded() []byte {
	return priv.sk.EncodeExpanded()
}

// Decodes an expanded 4896-byte private key as defined in FIPS 204.
// This is not the recommended way to store the private key, unless necessary for compatibility.
func PrivateKeyFromExpanded(expanded []byte) (*PrivateKey, error) {
	sk, err := internal.SkDecode(params.MLDSA87Cfg, expanded)
	if err != nil {
		return nil, err
	}
	return &PrivateKey{*sk}, nil
}
