package common

import (
	"crypto/subtle"
	"errors"
)

func SignInternal(k, l, beta, tau, omega uint8, eta int, lambda uint16, gamma1 uint32, gamma2 uint32, t0 RingVector, seed, K, tr, Mprime, rnd []byte) ([]byte, error) {
	hashed := H(append(seed[:], byte(k), byte(l)), 128)
	rho := make([]byte, 32)
	rhoprime := make([]byte, 64)
	copy(rho, hashed[0:32])
	copy(rhoprime, hashed[32:96])
	s1, s2 := ExpandS(k, l, eta, rhoprime)

	s1hat := NttVec(l, s1)
	s2hat := NttVec(k, s2)
	t0hat := NttVec(k, t0[:])
	Ahat := ExpandA(k, l, rho)

	// mu <- H(BytesToBits(tr) || M', 64)
	mu := H(append(tr[:], Mprime...), 64)

	// rhopp <- H(K || rnd || mu, 64)
	tmp := append(K[:], rnd...)
	tmp = append(tmp, mu...)
	rhopp := H(tmp, 64)

	// Rejection sampling loop
	iterations := 0
	kappa := uint16(0)
	for {
		fail := false
		y := ExpandMask(l, gamma1, rhopp, kappa)
		w := InvNttVec(k, MatrixVectorNTT(k, l, Ahat, NttVec(k, y)))
		w1 := HighBitsVec(k, gamma2, w)
		w1_encoded := W1Encode(k, gamma2, w1)
		c_tilde := H(append(mu, w1_encoded...), uint32(lambda>>2))
		c := SampleInBall(tau, c_tilde[:])
		c_hat := NTT(c)

		cs1 := InvNttVec(l, ScalarVectorNTT(l, c_hat, s1hat))
		cs2 := InvNttVec(k, ScalarVectorNTT(k, c_hat, s2hat))
		z := RingVectorAdd(l, y, cs1)
		r0 := LowBitsVec(k, gamma2, RingVectorSub(k, w, cs2))
		z_inf := InfinityNormRingVector(k, z)
		r0_inf := InfinityNormRingVector(k, r0)
		// return nil, errors.New("Unfinished")

		// Validity checks
		gamma1_beta := gamma1 - uint32(beta)
		gamma2_beta := gamma2 - uint32(beta)
		if z_inf >= gamma1_beta || r0_inf >= gamma2_beta {
			fail = true
		}
		// <<ct0>> <- NTT^-1(c_hat o t0_hat)
		ct0 := InvNttVec(k, ScalarVectorNTT(l, c_hat, t0hat))
		minus_ct0 := NegateRingVector(k, ct0)
		// w - cs2
		// w_cs2 := RingVectorSub(k, w, cs2)
		// w - cs2 + ct0
		// w_cs2_ct0 := RingVectorAdd(k, w_cs2, ct0)
		w_cs2_ct0 := RingVectorAdd(k, RingVectorSub(k, w, cs2), ct0)
		h := MakeHintRingVec(k, gamma2, minus_ct0, w_cs2_ct0)
		ct0_inf := InfinityNormRingVector(k, ct0)
		if ct0_inf >= gamma2 || CountOnesHint(k, h) > uint32(omega) {
			fail = true
		}
		kappa = kappa + uint16(l)

		// Loop termination via return
		if !fail {
			// Convert to the expected data type
			h_vec := NewRingVector(k)
			for i := range k {
				for j := range 256 {
					h_vec[i][j] = RingCoeff(uint32(h[i][j]))
				}
			}
			return SigEncode(k, l, omega, gamma1, c_tilde[:], z[:], h_vec), nil
		}

		// Appendix C - Loop Bounds
		iterations++
		if iterations > 814 {
			return nil, errors.New("too many rejections in common.SignInternal()")
		}
	}
}

func VerifyInternal(k, l, beta, tau, omega uint8, lambda uint16, gamma1, gamma2 uint32, rho []byte, t1 RingVector, Mprime, sigma []byte) bool {
	c_tilde, z, h, err := SigDecode(k, l, omega, lambda, gamma1, sigma)
	if err != nil {
		return false
	}
	if h == nil {
		return false
	}
	Ahat := ExpandA(k, l, rho)
	tr := H(PKEncode(k, rho, t1), 64)
	mu := H(append(tr, Mprime...), 64)
	c := SampleInBall(tau, c_tilde)

	z_hat := NttVec(k, z)
	c_hat := NTT(c)
	t1_2d := NewRingVector(k)
	for i := range k {
		for j := range 256 {
			ti := uint32(t1[i][j]) << d
			t1_2d[i][j] = CoeffReduceOnce(ti % q)
		}
	}
	t1_2d_hat := NttVec(k, t1_2d)
	ct1_2d_hat := ScalarVectorNTT(k, c_hat, t1_2d_hat)
	Azhat := MatrixVectorNTT(k, l, Ahat, z_hat)
	// w_approx := InvNttVec(k, Azhat - ct1_2d_hat)
	w_approx := InvNttVec(k, SubVectorNTT(k, Azhat, ct1_2d_hat))

	// Convert back
	h8 := make([][]uint8, k)
	for i := range k {
		h8[i] = make([]uint8, 256)
		for j := range 256 {
			h8[i][j] = uint8(h[i][j])
		}
	}

	w1 := UseHintRingVector(k, gamma2, h8, w_approx)
	w1_encoded := W1Encode(k, gamma2, w1)
	c_tilde_prime := H(append(mu, w1_encoded...), uint32(lambda>>2))
	z_inf := InfinityNormRingVector(k, z)
	b32 := uint32(beta)
	return z_inf <= (gamma1-b32) && subtle.ConstantTimeCompare(c_tilde, c_tilde_prime) == 1
}
