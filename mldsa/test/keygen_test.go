package mldsa_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"trailofbits.com/ml-dsa/mldsa/mldsa44"
	"trailofbits.com/ml-dsa/mldsa/mldsa65"
	"trailofbits.com/ml-dsa/mldsa/mldsa87"
)

func TestKeyGeneration(t *testing.T) {
	testVectors, err := ParseTestVectorFile[keyGenTestCase]("keygen_test.json")
	if err != nil {
		t.Fatalf("failed to parse test vector file: %v", err)
	}
	for _, testGroup := range testVectors.TestGroups {
		// Select key generation function based on the parameter set
		var generateKeyPair keyGenFunc
		switch testGroup.ParameterSet {
		case ML_DSA_44:
			generateKeyPair = generateKeyPair44
		case ML_DSA_65:
			generateKeyPair = generateKeyPair65
		case ML_DSA_87:
			generateKeyPair = generateKeyPair87
		default:
			t.Fatalf("unknown parameter set: %s", testGroup.ParameterSet)
		}
		// Run key generation tests
		for _, test := range testGroup.Tests {
			sk, vk := generateKeyPair(test.Seed)
			// Verify correctness via serialization
			assert.Equal(t, test.SK.Bytes(), sk.ExpandedBytesForTesting(), "signing keys differ in test case %d", test.Id)
			assert.Equal(t, test.VK.Bytes(), vk.Bytes(), "verifying keys differ in test case %d", test.Id)
		}
	}
}

// keyGenFunc defines the function signature for the internal, seed-based key
// generation functions
type keyGenFunc func(seed Seed) (SigningKey, VerifyingKey)

func generateKeyPair44(seed Seed) (SigningKey, VerifyingKey) {
	return mldsa44.KeyGenInternal(seed)
}

func generateKeyPair65(seed Seed) (SigningKey, VerifyingKey) {
	return mldsa65.KeyGenInternal(seed)
}

func generateKeyPair87(seed Seed) (SigningKey, VerifyingKey) {
	return mldsa87.KeyGenInternal(seed)
}

type keyGenTestCase struct {
	Id   int      `json:"tcId"`
	Seed Seed     `json:"seed"`
	SK   HexBytes `json:"sk"`
	VK   HexBytes `json:"pk"`
}
