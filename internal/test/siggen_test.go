package mldsa_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	internal "trailofbits.com/ml-dsa/internal"
	"trailofbits.com/ml-dsa/internal/params"
	options "trailofbits.com/ml-dsa/options"
)

func TestSignatureGeneration(t *testing.T) {
	testVectors, err := ParseTestVectorFile[*sigGenTestGroup]("siggen_test.json")
	if err != nil {
		t.Fatalf("failed to parse test vector file: %v", err)
	}

	for _, testGroup := range testVectors.TestGroups {
		name := fmt.Sprintf("TestGroup-%d", testGroup.id)
		t.Run(name, func(t *testing.T) {
			// Skip external mu test groups, we don't support the feature
			// Skip prehash test groups, we don't support the feature
			if testGroup.externalMu || testGroup.preHash == "preHash" {
				t.Log("skipping test group with external mu or preHash")
				return
			}

			if len(testGroup.tests) == 0 {
				panic("no test cases found")
			}
			for _, test := range testGroup.tests {
				var sig []byte
				var err error
				var rnd []byte

				if testGroup.deterministic {
					rnd = make([]byte, 32)
				} else {
					rnd = test.rnd
				}

				sk, err := internal.SkDecode(testGroup.parameterSet, test.sk)
				assert.NoError(t, err, "failed to parse signing key in test case %d", test.id)

				if testGroup.signatureInterface == "internal" {
					sig = sk.SignInternal(test.msg, rnd[:])
				} else {
					reader := bytes.NewReader(rnd[:])
					sig, err = sk.Sign(reader, test.msg, &options.Options{Context: string(test.ctx)})
					assert.NoError(t, err, "failed to sign message in test case %d", test.id)
				}
				assert.NoError(t, err, "failed to sign message in test case %d", test.id)
				assert.Equal(t, test.sig, sig, "signatures differ in test case %d", test.id)
			}
		})
	}
}

type sigGenTestCase struct {
	id  int
	sk  []byte
	vk  []byte
	msg []byte
	rnd []byte
	ctx []byte
	sig []byte
}

type sigGenTestCaseMarshaller struct {
	Id  int     `json:"tcId"`
	SK  string  `json:"sk"`
	VK  string  `json:"pk"`
	Msg *string `json:"message"`
	Ctx *string `json:"context"`
	Rnd *string `json:"rnd"`
	Sig string  `json:"signature"`
}

func (t *sigGenTestCase) UnmarshalJSON(data []byte) error {
	var tRaw sigGenTestCaseMarshaller
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return err
	}
	t.id = tRaw.Id
	t.sk, _ = hex.DecodeString(tRaw.SK)
	t.vk, _ = hex.DecodeString(tRaw.VK)
	if tRaw.Msg != nil {
		t.msg, _ = hex.DecodeString(*tRaw.Msg)
	}
	if tRaw.Rnd != nil {
		t.rnd, _ = hex.DecodeString(*tRaw.Rnd)
	}
	if tRaw.Ctx != nil {
		t.ctx, _ = hex.DecodeString(*tRaw.Ctx)
	}
	t.sig, _ = hex.DecodeString(tRaw.Sig)
	return nil
}

type sigGenTestGroup struct {
	id                 int
	parameterSet       *params.Cfg
	deterministic      bool
	signatureInterface string
	preHash            string
	externalMu         bool
	tests              []sigGenTestCase
}

type sigGenTestGroupUnmarshaler struct {
	TgId               int              `json:"tgId"`
	ParameterSet       string           `json:"parameterSet"`
	Deterministic      bool             `json:"deterministic"`
	SignatureInterface string           `json:"signatureInterface"`
	PreHash            string           `json:"preHash"`
	ExternalMu         bool             `json:"externalMu"`
	Tests              []sigGenTestCase `json:"tests"`
}

func (tg *sigGenTestGroup) UnmarshalJSON(data []byte) error {
	tgRaw := new(sigGenTestGroupUnmarshaler)
	if err := json.Unmarshal(data, &tgRaw); err != nil {
		return err
	}
	tg.id = tgRaw.TgId
	tg.parameterSet = parseCfg(tgRaw.ParameterSet)
	tg.deterministic = tgRaw.Deterministic
	tg.signatureInterface = tgRaw.SignatureInterface
	tg.preHash = tgRaw.PreHash
	tg.externalMu = tgRaw.ExternalMu
	tg.tests = tgRaw.Tests
	return nil
}
