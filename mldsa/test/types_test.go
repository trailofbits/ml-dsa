package mldsa_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
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
