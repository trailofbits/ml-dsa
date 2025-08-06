package mldsa_test

import (
	"encoding/json"
	"fmt"
	"os"

	"trailofbits.com/ml-dsa/internal/params"
)

func parseCfg(paramString string) *params.Cfg {
	switch paramString {
	case "ML-DSA-44":
		return params.MLDSA44Cfg
	case "ML-DSA-65":
		return params.MLDSA65Cfg
	case "ML-DSA-87":
		return params.MLDSA87Cfg
	default:
		panic(fmt.Sprintf("unknown parameter set: %q", paramString))
	}
}

// TestVectorFile represents the structure of the test vector file
type TestVectorFile[TestGroup json.Unmarshaler] struct {
	TestGroups []TestGroup `json:"testGroups"`
}

func ParseTestVectorFile[TestGroup json.Unmarshaler](path string) (*TestVectorFile[TestGroup], error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck

	testVectors := new(TestVectorFile[TestGroup])
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&testVectors); err != nil {
		return nil, err
	}

	return testVectors, nil
}
