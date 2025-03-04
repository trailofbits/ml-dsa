package mldsa65

import (
	"trailofbits.com/ml-dsa/mldsa/common"
)

func pkEncode(rho []byte, t1 common.RingVector) []byte {
	return common.PKEncode(k, rho, t1)
}

func hintBitPack(h common.RingVector) []byte {
	return common.HintBitPack(k, ω, h)
}

func hintBitUnpack(y []byte) (common.RingVector, error) {
	return common.HintBitUnpack(k, ω, y)
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

func expandMask(rho []byte, mu uint16) common.RingVector {
	return common.ExpandMask(l, γ1, rho, mu)
}

func decompose(r uint32) (uint32, uint32) {
	return common.Decompose(γ2, r)
}

func highBits(r uint32) uint32 {
	return common.HighBits(γ2, r)
}

func lowBits(r uint32) uint32 {
	return common.LowBits(γ2, r)
}

func makeHint(z, r common.FieldElement) uint8 {
	return common.MakeHint(γ2, z, r)
}

func useHint(h uint8, r common.FieldElement) common.FieldElement {
	return common.UseHint(γ2, h, r)
}
func addVectorNTT(v, w common.NttVector) common.NttVector {
	return common.AddVectorNTT(l, v, w)
}

func scalarVectorNTT(c_hat common.NttElement, v_hat common.NttVector) common.NttVector {
	return common.ScalarVectorNTT(l, c_hat, v_hat)
}

func matrixVectorNTT(M_hat common.NttMatrix, v_hat common.NttVector) common.NttVector {
	return common.MatrixVectorNTT(k, l, M_hat, v_hat)
}

func ringPower2Round(r common.RingElement) (common.RingElement, common.RingElement) {
	return common.RingPower2Round(k, r)
}

func ringVecPower2Round(r common.RingVector) (common.RingVector, common.RingVector) {
	return common.RingVecPower2Round(k, r)
}
