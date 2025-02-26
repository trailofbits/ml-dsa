package mldsa87

const (
	// ML-DSA-87 specific parameters:
	т  = uint8(60)
	λ  = uint16(256)
	γ1 = uint32(524288)
	γ2 = uint32(261888) // (q-1)/32
	k  = uint8(8)
	l  = uint8(7)
	η  = uint8(2)
	β  = uint8(120) // т * η
	ω  = uint8(75)

	ChallengeEntropy = uint16(257)
)
