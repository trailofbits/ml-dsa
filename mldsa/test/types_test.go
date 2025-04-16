package mldsa_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

const (
	ML_DSA_44 = ParameterSet("ML-DSA-44")
	ML_DSA_65 = ParameterSet("ML-DSA-65")
	ML_DSA_87 = ParameterSet("ML-DSA-87")
)

// A wrapper type used to distinguish between different ML-DSA parameter sets
type ParameterSet string

func (p *ParameterSet) UnmarshalJSON(data []byte) error {
	var paramString string
	if err := json.Unmarshal(data, &paramString); err != nil {
		return err
	}
	switch paramString {
	case "ML-DSA-44":
		*p = ML_DSA_44
	case "ML-DSA-65":
		*p = ML_DSA_65
	case "ML-DSA-87":
		*p = ML_DSA_87
	default:
		return fmt.Errorf("unknown parameter set: %q", paramString)
	}
	return nil
}

// A wrapper type used to unmarshal hexadecimal strings
type HexBytes []byte

func (h *HexBytes) UnmarshalJSON(data []byte) error {
	var hexString string
	if err := json.Unmarshal(data, &hexString); err != nil {
		return err
	}
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}
	*h = bytes
	return nil
}

func (h *HexBytes) Bytes() []byte {
	return *h
}

type Seed [32]byte

func (s *Seed) UnmarshalJSON(data []byte) error {
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

// SigningKey defines the tested interface for signing keys in ML-DSA
type SigningKey interface {
	ExpandedBytesForTesting() []byte
	SignInternal(message, ctx []byte) ([]byte, error)
}

// VerifyingKey defines the tested interface for verifying keys in ML-DSA
type VerifyingKey interface {
	Bytes() []byte
	Verify(message, ctx, signature []byte) bool
}

// TestVectorFile represents the structure of the test vector file
type TestVectorFile[TestCase any] struct {
	TestGroups []*TestGroup[TestCase] `json:"testGroups"`
}

func ParseTestVectorFile[TestCase any](path string) (TestVectorFile[TestCase], error) {
	file, err := os.Open(path)
	if err != nil {
		return TestVectorFile[TestCase]{}, err
	}
	defer file.Close()

	var testVectors TestVectorFile[TestCase]
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&testVectors); err != nil {
		return TestVectorFile[TestCase]{}, err
	}

	return testVectors, nil
}

type TestGroup[TestCase any] struct {
	Id           int          `json:"tgId"`
	ParameterSet ParameterSet `json:"parameterSet"`
	Tests        []TestCase   `json:"tests"`
}
