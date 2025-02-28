// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mldsa44

import (
	"crypto/rand"

	"trailofbits.com/ml-dsa/mldsa/common"
)

const (
	// ML-DSA-44 specific parameters:
	т  = uint8(39)
	λ  = uint16(128)
	γ1 = uint32(131072)
	γ2 = uint32(95232) // (q-1)/88
	k  = uint8(4)
	l  = uint8(4)
	η  = uint8(2)
	β  = uint8(78) // т * η
	ω  = uint8(80)

	ChallengeEntropy = uint16(192)

	// We use seeds instead of expanded private keys
	SeedSize = 32
	// Expanded secret key (only used for testing)
	ExpandedSecretKeySize = 2560
	// Public Key size (bytes) for ML-DSA-44
	PublicKeySize = 1312
	// Signature size (bytes) for ML-DSA-44
	SignatureSize = 2420
)

type SigningKey struct {
	seed [32]byte // ξ from the specification
	ρ    [32]byte // Rho is the public seed
	K    [32]byte
	tr   [64]byte
	t1   [k]common.RingElement
}

// 32 + (32 * k * (bitlen(q-1)-d))
// 32 + (32 * k * (23 - 13))
// 32 + 320 * k
type VerifyingKey struct {
	ρ  [32]byte // Rho is the public seed
	t1 [k]common.RingElement
}

func GenerateKeyPair() (SigningKey, VerifyingKey, error) {
	seed := make([]byte, SeedSize)
	_, err := rand.Read(seed)
	if err != nil {
		return SigningKey{}, VerifyingKey{}, err
	}
	var seed_copy [SeedSize]byte
	copy(seed_copy[:], seed[0:SeedSize])
	sk, pk := KeyGenInternal(seed_copy)
	return sk, pk, nil
}

// Algorithm 7
func KeyGenInternal(seed [32]byte) (SigningKey, VerifyingKey) {
	var rho_copy [32]byte
	var K_copy [32]byte
	var tr_copy [64]byte
	var t1_copy [k]common.RingElement
	var t0_copy [k]common.RingElement

	hashed := common.H(append(seed[:], byte(k), byte(l)), 128)
	rho, rhoprime, K := hashed[0:32], hashed[32:96], hashed[96:]
	copy(rho_copy[:], rho[:])
	copy(K_copy[:], K[:])

	Ahat := expandA(rho)
	s1, s2 := expandS(rhoprime)

	// t = (A * s1) + s2
	s1hat := common.NttVec(l, s1)
	multiplied := common.MatrixVectorNTT(k, l, Ahat, s1hat)
	inverted := common.InvNttVec(l, multiplied)
	t := common.RingVectorAdd(k, inverted, s2)

	// multiplied := nttMul(Ahat, common.NTT(s1))
	// polynomial := inverseNTT(multiplied)
	// t := ringAdd(polynomial, s2)
	t1, t0 := ringVecPower2Round(t)

	pke := pkEncode(rho, t1[:])
	tr := common.H(pke, 64)
	copy(tr_copy[:], tr[:])
	for i := range int(k) {
		t1_copy[i] = common.NewRingElement()
		for j := range 256 {
			t1_copy[i][j] = t1[i][j]
			t0_copy[i][j] = t0[i][j]
		}
	}
	sk := SigningKey{seed, rho_copy, K_copy, tr_copy, t0_copy}
	vk := VerifyingKey{rho_copy, t1_copy}
	return sk, vk
}

func (sk SigningKey) Bytes() []byte {
	var b [SeedSize]byte
	copy(b[:], sk.seed[:])
	return b[:]
}

// We do not recommend actually ever using this. Store the seed instead.
func (sk SigningKey) ExpandedBytesForTesting() []byte {
	hashed := common.H(append(sk.seed[:], byte(k), byte(l)), 128)
	rho, rhoprime := hashed[0:32], hashed[32:96]
	Ahat := expandA(rho)
	s1, s2 := expandS(rhoprime)

	// t = (A * s1) + s2
	s1hat := common.NttVec(l, s1)
	multiplied := common.MatrixVectorNTT(k, l, Ahat, s1hat)
	inverted := common.InvNttVec(l, multiplied)
	t := common.RingVectorAdd(k, inverted, s2)

	// multiplied := nttMul(Ahat, common.NTT(s1))
	// polynomial := inverseNTT(multiplied)
	// t := ringAdd(polynomial, s2)
	_, t0 := ringVecPower2Round(t)
	// END DEBUG

	return skEncode(sk.ρ[:], sk.K[:], sk.tr[:], s1, s2, t0)
}

func (sk *SigningKey) VerificationKey() VerifyingKey {
	return VerifyingKey{sk.ρ, sk.t1}
}

func (vk *VerifyingKey) Bytes() []byte {
	var b [PublicKeySize]byte
	encoded := pkEncode(vk.ρ[:], vk.t1[:])
	copy(b[:], encoded[:])
	return b[:]
}
