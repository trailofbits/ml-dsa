// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package util implements the generic underlying algorithms from [NIST FIPS 204].
//
// This aux.go file contains the auxiliary internal functions needed to implement ML-DSA.
//
// The implementations here have slightly more verbose function prototypes than FIPS-204,
// due to the need to parametrize the functions based on the actual parameter sets for the
// ML-DSA algorithm being used. The first parameters to all functions are constants to the
// specific instantiation of ML-DSA (e.g., k, l, or omega2).
//
// [NIST FIPS 204]: https://doi.org/10.6028/NIST.FIPS.204
package util

import (
	"crypto/subtle"

	"golang.org/x/crypto/sha3"
	"trailofbits.com/ml-dsa/mldsa/internal/field"
	"trailofbits.com/ml-dsa/mldsa/internal/params"
	"trailofbits.com/ml-dsa/mldsa/internal/ring"
)

// Algorithm 22, parametrized by `k` from each ML-DSA parameter set
func PKEncode(k uint8, rho []byte, t1 []ring.Rz) []byte {
	var pk []byte
	pk = append(pk, rho...)
	for i := range k {
		packed := SimpleBitPack(t1[i], params.D)
		pk = append(pk, packed...)
	}
	return pk
}

// Algorithm 23, parametrized by `k` from each ML-DSA parameter set
func PKDecode(k uint8, pk []byte) ([]byte, []ring.Rz) {
	rho := pk[0:32]
	z := pk[32:]
	t1 := make([]ring.Rz, k)
	j := 0
	for i := range k {
		t1[i] = SimpleBitUnpack(z[j:], params.D-1)
		j += 40
	}
	return rho, t1
}

// Algorithm 24
func SKEncode(k, l, log_eta uint8, rho, K, tr []byte, s1, s2, t0 []ring.Rz) []byte {
	sk := rho[:]
	sk = append(sk, K...)
	sk = append(sk, tr...)
	for i := range l {
		packed := BitPackClosed(s1[i], log_eta)
		sk = append(sk, packed[:]...)
	}
	for i := range k {
		packed := BitPackClosed(s2[i], log_eta)
		sk = append(sk, packed[:]...)
	}

	for i := range k {
		packed := BitPack(t0[i], params.D-1)
		sk = append(sk, packed[:]...)
	}
	return sk
}

// Algorithm 26
func SigEncode(cfg *params.Cfg, c []byte, z []ring.Rq, h []ring.R2) []byte {
	sigma := c[:]
	for i := range cfg.L {
		packed := BitPack(z[i].Symmetric(), cfg.LogGamma1)
		sigma = append(sigma, packed...)
	}
	packed := HintBitPack(cfg.K, cfg.Omega, h)
	sigma = append(sigma, packed...)
	return sigma[:]
}

// Algorithm 27
func SigDecode(cfg *params.Cfg, sig []byte) ([]byte, []ring.Rz, []ring.R2, error) {
	z := make([]ring.Rz, cfg.K)

	length := cfg.Lambda / 4
	c, sigma := sig[:length], sig[length:]
	var x []byte
	elemLen := 32 * (1 + int(cfg.LogGamma1))
	for i := range cfg.L {
		x, sigma = sigma[:elemLen], sigma[elemLen:]
		z[i] = BitUnpack(x, cfg.LogGamma1)
	}
	h, err := HintBitUnpack(cfg.K, cfg.Omega, sigma)
	if err != nil {
		return nil, nil, nil, err
	}
	return c, z, h, nil
}

// Algorithm 28
// This is just SimpleBitPack with a precomputed per-paramset length
func W1Encode(cfg *params.Cfg, w1 []ring.Rz) []byte {
	var w []byte
	for i := range cfg.K {
		packed := SimpleBitPack(w1[i], cfg.W1Bits)
		w = append(w, packed...)
	}
	return w
}

// Algorithm 29
func SampleInBall(cfg *params.Cfg, seed []byte) (c ring.Rz) {
	var s [8]byte
	ctx := sha3.NewShake256()
	ctx.Write(seed)
	ctx.Read(s[:])

	var j [1]byte
	// The NIST specification says "to 255" which means "< 256"
	// for i from (256 - tau) to 255 do
	for i := uint16(256 - cfg.Tau); i < 256; i++ {
		ctx.Read(j[:])
		for uint16(j[0]) > i {
			ctx.Read(j[:])
		}
		j0 := j[0]
		c[i] = c[j0]

		// Extract bit (i + tau - 256)
		idx := uint16(i + cfg.Tau - 256)
		h := (s[idx/8] >> (idx & 7)) & 1

		// Swap between 1 and -1 without side-channels
		c[j0] = 1 - int32(h<<1)
	}
	return c
}

// Algorithm 30
func RejNTTPoly(seed []byte) (ah ring.Tq) {
	ctx := sha3.NewShake128()
	var s [3]byte
	ctx.Write(seed)
	for j := 0; j < 256; j++ {
		var tmp *field.T
		for tmp == nil {
			ctx.Read(s[:])
			tmp = field.FromThreeBytes(s[0], s[1], s[2])
		}
		ah[j] = *tmp
	}
	return ah
}

// Algorithm 31
func RejBoundedPoly(eta int, seed []byte) (a ring.Rq) {
	var z [1]byte
	ctx := sha3.NewShake256()
	ctx.Write(seed)
	for j := 0; j < 256; {
		ctx.Read(z[:])
		z0 := field.FromHalfByte(eta, z[0]&0xf)
		if z0 != nil {
			a[j] = *z0
			j++
		}
		z1 := field.FromHalfByte(eta, z[0]>>4)
		if z1 != nil {
			if j < 256 {
				a[j] = *z1
				j++
			}
		}
	}
	return a
}

// TODO - make cfg usage more consistent
// Algorithm 32
func ExpandA(cfg *params.Cfg, rho []byte) [][]ring.Tq {
	k, l := cfg.K, cfg.L
	Ahat := make([][]ring.Tq, k)
	for r := range k {
		Ahat[r] = make([]ring.Tq, l)
		for s := range l {
			rhoprime := append(rho, byte(s), byte(r))
			Ahat[r][s] = RejNTTPoly(rhoprime)
		}
	}
	return Ahat
}

// Appends a uint16 as a []byte of length 2, in little-endian order
// Modifies rho in place, if rho has sufficient capacity.
func tweakUint16(rho []byte, x uint16) []byte {
	return append(rho, byte(x&0xff), byte(x>>8))
}

// Algorithm 33
func ExpandS(cfg *params.Cfg, rho []byte) ([]ring.Rq, []ring.Rq) {
	k, l, eta := cfg.K, cfg.L, 1<<cfg.LogEta
	s1 := make([]ring.Rq, l)
	s2 := make([]ring.Rq, k)

	// copy rho so that we can tweak in-place
	packed := make([]byte, 0, len(rho)+2)
	rho = append(packed, rho...)

	for r := range l {
		s1[r] = RejBoundedPoly(eta, tweakUint16(rho, uint16(r)))
	}
	for r := range k {
		s2[r] = RejBoundedPoly(eta, tweakUint16(rho, uint16(r+l)))
	}
	return s1, s2
}

// H(str, l) -> SHAKE256(str, 8l)
func H(out []byte, data []byte) {
	ctx := sha3.NewShake256()
	ctx.Write(data)
	ctx.Read(out)
}

// Algorithm 34
func ExpandMask(cfg *params.Cfg, rho []byte, mu uint16) []ring.Rz {
	y := make([]ring.Rz, cfg.L)
	c := uint32(1 + cfg.LogGamma1)

	// copy rho so that we can tweak in-place
	packed := make([]byte, 0, len(rho)+2)
	rho = append(packed, rho...)

	v := make([]byte, c<<5)
	for r := range cfg.L {
		// rho' <- rho || IntegerToBytes(mu + r, 2)
		packed := tweakUint16(rho, uint16(r)+mu)
		// v <- H(rho', 32c)
		H(v, packed)
		// y[r] = BitUnpack(v, gamma1 - 1, gamma1)
		y[r] = BitUnpack(v, cfg.LogGamma1)
	}
	return y
}

func makeHint(cfg *params.Cfg, z, r field.T) uint8 {
	r1 := r.HighBits(cfg.Gamma2)
	v1 := r.Add(z).HighBits(cfg.Gamma2)
	return uint8(1 - subtle.ConstantTimeEq(r1, v1))
}

// Algorithm 39
// Not constant time - inputs and outputs are public
// Returns nil when the number of 1s in the hint is greater than omega
func MakeHint(cfg *params.Cfg, z, r []ring.Rq) []ring.R2 {
	hints := make([]ring.R2, cfg.K)
	weight := 0
	for i := range cfg.K {
		for j := range params.N {
			hints[i][j] = makeHint(cfg, z[i][j], r[i][j])
			weight += int(hints[i][j])
		}
	}
	if weight > int(cfg.Omega) {
		return nil
	}
	return hints
}

// Algorithm 40
// Not constant time - inputs and outputs are public
func UseHint(cfg *params.Cfg, h []ring.R2, r []ring.Rq) []ring.Rz {
	m := int32((params.Q - 1) / (2 * cfg.Gamma2))
	v := make([]ring.Rz, cfg.K)
	for i := range cfg.K {
		for j := range params.N {
			r1, r0 := r[i][j].Decompose(cfg.Gamma2)
			v[i][j] = r1
			if h[i][j] == 1 {
				if r0 > 0 {
					v[i][j] = (r1 + 1) % m
				} else {
					v[i][j] = (r1 - 1 + m) % m
				}
			}
		}
	}
	return v
}

// Multiplies each element of a vector by a scalar
func ScalarVector(c field.T, v []ring.Rq) []ring.Rq {
	w := make([]ring.Rq, len(v))
	for i := range len(v) {
		w[i] = v[i].ScalarMul(c)
	}
	return w
}
