package mldsa_test

import (
	"testing"

	"trailofbits.com/ml-dsa/mldsa/common"
	"trailofbits.com/ml-dsa/mldsa/mldsa44"
)

func TestSignatureGeneration(t *testing.T) {
	testVectors, err := ParseTestVectorFile[sigGenTestCase]("siggen_test.json")
	if err != nil {
		t.Fatalf("failed to parse test vector file: %v", err)
	}
	for _, testGroup := range testVectors.TestGroups {
		// Select key generation function based on the parameter set
		var decodeSigningKey skDecodeFunc
		switch testGroup.ParameterSet {
		case ML_DSA_44:
			decodeSigningKey = generateKeyPair44
		case ML_DSA_65:
			decodeSigningKey = generateKeyPair65
		case ML_DSA_87:
			decodeSigningKey = generateKeyPair87
		default:
			t.Fatalf("unknown parameter set: %s", testGroup.ParameterSet)
		}
		// Run key generation tests
		for _, test := range testGroup.Tests {
			// Verify correctness via serialization
		}
	}
}

// skDecodeFunc defines the function signature for the function that decodes
// a secret key from an extended byte representation
type skDecodeFunc func(bytes HexBytes) (SigningKey, error)

func decodeSigningKey44(bytes HexBytes) (SigningKey, error) {
	rho, K, tr, s1, s2, t0 := common.SKDecode(mldsa44.k, mldsa44.l, mldsa44.η, bytes)
	return mldsa44.SigningKey{
		ρ:  rho,
		K:  K,
		tr: tr,
		t0: s1,
		t1: s2,
	}, nil
}

type sigGenTestCase struct {
	Id  int      `json:"tcId"`
	SK  HexBytes `json:"sk"`
	VK  HexBytes `json:"pk"`
	Msg HexBytes `json:"message"`
	Ctx HexBytes `json:"context"`
	Sig HexBytes `json:"signature"`
}
