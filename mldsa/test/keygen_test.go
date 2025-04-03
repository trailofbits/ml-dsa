package mldsa_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"trailofbits.com/ml-dsa/mldsa/mldsa44"
	"trailofbits.com/ml-dsa/mldsa/mldsa65"
	"trailofbits.com/ml-dsa/mldsa/mldsa87"
)

func TestKeyGeneration(t *testing.T) {
	testVectors, err := parseTestVectorFile("keygen_test.json")
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
			if test.Id != 13 {
				// TODO: Skip all but the first failing test case
				continue
			}
			_, vk := generateKeyPair(test.Seed)
			// Verify correctness via serialization
			// TODO: Skip signing key comparison for now
			// assert.Equal(t, test.SK.Bytes(), sk.ExpandedBytesForTesting(), "signing keys differ in test case %d", test.Id)
			assert.Equal(t, test.VK.Bytes(), vk.Bytes(), "verifying keys differ in test case %d", test.Id)
		}
	}
}

// keyGenFunc defines the function signature for the internal, seed-based key
// generation functions
type keyGenFunc func(seed seed) (SigningKey, VerifyingKey)

func generateKeyPair44(seed seed) (SigningKey, VerifyingKey) {
	return mldsa44.KeyGenInternal(seed)
}

func generateKeyPair65(seed seed) (SigningKey, VerifyingKey) {
	return mldsa65.KeyGenInternal(seed)
}

func generateKeyPair87(seed seed) (SigningKey, VerifyingKey) {
	return mldsa87.KeyGenInternal(seed)
}

type testVectorFile struct {
	TestGroups []testGroup `json:"testGroups"`
}

func parseTestVectorFile(path string) (testVectorFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return testVectorFile{}, err
	}
	defer file.Close()

	var testVectors testVectorFile
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&testVectors); err != nil {
		return testVectorFile{}, err
	}

	return testVectors, nil
}

type testGroup struct {
	Id           int          `json:"tgId"`
	ParameterSet ParameterSet `json:"parameterSet"`
	Tests        []testCase   `json:"tests"`
}

type testCase struct {
	Id   int      `json:"tcId"`
	Seed seed     `json:"seed"`
	SK   HexBytes `json:"sk"`
	VK   HexBytes `json:"pk"`
}

type seed [32]byte

func (s *seed) UnmarshalJSON(data []byte) error {
	var hexString string
	if err := json.Unmarshal(data, &hexString); err != nil {
		return err
	}
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}
	if len(bytes) != 32 {
		return fmt.Errorf("invalid length for HexArray: expected 32 bytes, got %d", len(bytes))
	}
	copy(s[:], bytes)
	return nil
}
