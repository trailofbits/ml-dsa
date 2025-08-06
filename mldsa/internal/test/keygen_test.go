package mldsa_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"trailofbits.com/ml-dsa/mldsa/internal"
	"trailofbits.com/ml-dsa/mldsa/internal/params"
)

func TestKeyGeneration(t *testing.T) {
	testVectors, err := ParseTestVectorFile[*keyGenTestGroup]("keygen_test.json")
	if err != nil {
		t.Fatalf("failed to parse test vector file: %v", err)
	}
	for _, testGroup := range testVectors.TestGroups {
		name := fmt.Sprintf("TestGroup-%d", testGroup.id)
		t.Run(name, func(t *testing.T) {
			if len(testGroup.tests) == 0 {
				panic("no test cases found")
			}
			for _, test := range testGroup.tests {
				sk, err := internal.FromSeed(testGroup.parameterSet, test.Seed)
				assert.NoError(t, err, "failed to generate signing key in test case %d", test.Id)
				// Verify correctness via serialization
				assert.Equal(t, test.SK, sk.EncodeExpanded(), "signing keys differ in test case %d", test.Id)
				assert.Equal(t, test.VK, sk.Public().Bytes(), "verifying keys differ in test case %d", test.Id)
			}
		})
	}
}

type keyGenTestCaseMarshaller struct {
	Id   int    `json:"tcId"`
	Seed string `json:"seed"`
	SK   string `json:"sk"`
	VK   string `json:"pk"`
}

type keyGenTestCase struct {
	Id   int
	Seed []byte
	SK   []byte
	VK   []byte
}

func (t *keyGenTestCase) UnmarshalJSON(data []byte) error {
	var tRaw keyGenTestCaseMarshaller
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return err
	}
	t.Id = tRaw.Id
	t.Seed, _ = hex.DecodeString(tRaw.Seed)
	t.SK, _ = hex.DecodeString(tRaw.SK)
	t.VK, _ = hex.DecodeString(tRaw.VK)
	return nil
}

type keyGenTestGroup struct {
	id           int
	parameterSet *params.Cfg
	tests        []keyGenTestCase
}

type keyGenTestGroupUnmarshaler struct {
	TgId         int              `json:"tgId"`
	ParameterSet string           `json:"parameterSet"`
	Tests        []keyGenTestCase `json:"tests"`
}

func (tg *keyGenTestGroup) UnmarshalJSON(data []byte) error {
	tgRaw := new(keyGenTestGroupUnmarshaler)
	if err := json.Unmarshal(data, &tgRaw); err != nil {
		return err
	}
	tg.id = tgRaw.TgId
	tg.parameterSet = parseCfg(tgRaw.ParameterSet)
	tg.tests = tgRaw.Tests
	return nil
}
