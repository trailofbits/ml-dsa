package common_test

import (
	// "encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"trailofbits.com/ml-dsa/mldsa/common"
)

const (
	q = 8380417
)

func TestFieldReduceOnce(t *testing.T) {
	for i := range 256 {
		j := common.FieldReduceOnce(uint32(i))
		assert.Equal(t, common.FieldElement(i), j)

		j = common.FieldReduceOnce(uint32(i + q))
		assert.Equal(t, common.FieldElement(i), j)
	}
}

func TestFieldAdd(t *testing.T) {
	fe_q := common.Uint32ToFieldElement(q)
	for i := range uint32(256) {
		fe_i := common.Uint32ToFieldElement(i)
		k := common.FieldAdd(fe_i, fe_q)
		assert.Equal(t, uint32(k), uint32(i%q))
		for j := range uint32(1024) {
			fe_j := common.Uint32ToFieldElement(j)
			k = common.FieldAdd(fe_i, fe_j)
			assert.Less(t, uint32(k), uint32(q))
			fe_jq := common.Uint32ToFieldElement(j + (q - 1024))
			k = common.FieldAdd(fe_i, fe_jq)
			assert.Less(t, uint32(k), uint32(q))
		}
	}
}

func TestFieldSub(t *testing.T) {
	fe_q := common.Uint32ToFieldElement(q)
	for i := range uint32(256) {
		fe_i := common.Uint32ToFieldElement(i)
		k := common.FieldSub(fe_i, fe_q)
		assert.Equal(t, uint32(k), uint32(i))
		for j := range uint32(1024) {
			fe_j := common.Uint32ToFieldElement(j)
			k = common.FieldSub(fe_i, fe_j)
			assert.Less(t, uint32(k), uint32(q))
			fe_jq := common.Uint32ToFieldElement(j + (q - 1024))
			k = common.FieldSub(fe_i, fe_jq)
			assert.Less(t, uint32(k), uint32(q))
		}
	}
}

/*
func TestCoeffReduceOnce(t *testing.T) {
	for i := range 256 {
		j := CoeffReduceOnce(int16(i))
		assert.Equal(t, i, j)

		j := CoeffReduceOnce(int16(i))
		assert.Equal(t, i, j)
	}
}

func TestPKTreeKAT1(t *testing.T) {
	seed, err := hex.DecodeString("558b8966c48ae9cb898b423c83443aae014a72f1b1ab5cc85cf1d892903b5439")
	if err != nil {
		panic(err)
	}
}
*/
