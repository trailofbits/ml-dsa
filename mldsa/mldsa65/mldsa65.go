package mldsa65

const (
	// ML-DSA-65 specific parameters:
	т  = uint8(49)
	λ  = uint16(192)
	γ1 = uint32(524288)
	γ2 = uint32(261888) // (q-1)/32
	k  = uint8(6)
	l  = uint8(5)
	η  = uint8(4)
	β  = uint8(196) // т * η
	ω  = uint8(55)

	ChallengeEntropy = uint16(225)
)
