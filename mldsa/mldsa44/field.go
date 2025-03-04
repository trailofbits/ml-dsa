package mldsa44

import (
	"trailofbits.com/ml-dsa/mldsa/common"
)

func pkEncode(rho []byte, t1 common.RingVector) []byte {
	return common.PKEncode(k, rho, t1)
}

func skEncode(rho, K, tr []byte, s1, s2, t0 common.RingVector) []byte {
	return common.SKEncode(k, l, η, rho, K, tr, s1, s2, t0)
}

func skDecode(sk []byte) ([]byte, []byte, []byte, common.RingVector, common.RingVector, common.RingVector) {
	return common.SKDecode(k, l, η, sk)
}

func sigEncode(c []byte, z, h common.RingVector) []byte {
	return common.SigEncode(k, l, ω, γ1, c, z, h)
}

func sigDecode(sig []byte) ([]byte, common.RingVector, common.RingVector, error) {
	return common.SigDecode(k, l, ω, λ, γ1, sig)
}

func w1Encode(w1 common.RingVector) []byte {
	return common.W1Encode(k, γ2, w1)
}

func sampleInBall(seed []byte) common.RingElement {
	return common.SampleInBall(т, seed)
}

func expandA(rho []byte) common.NttMatrix {
	return common.ExpandA(k, l, rho)
}

func expandS(rho []byte) (common.RingVector, common.RingVector) {
	return common.ExpandS(k, l, int(η), rho)
}

func ringVecPower2Round(r common.RingVector) (common.RingVector, common.RingVector) {
	return common.RingVecPower2Round(k, r)
}
