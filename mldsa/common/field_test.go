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

func TestFieldReuce(t *testing.T) {
	tests := []struct {
		a uint64
		b uint32
	}{
		{uint64(q) + 1, 1},
		{uint64(q) + 321, 321},
		{uint64(q * 2), 0},
		{uint64(q*2) + 1, 1},
		{uint64(q) * 100, 0},
		{uint64(q)*100 + 1, 1},
		{uint64(q)*10000 + 1, 1},
		{uint64(q)*100000 + 1, 1},
		{uint64(q)*100000 + 2, 2},
	}

	for _, tt := range tests {
		a := common.FieldReduce(tt.a)
		b := common.Uint32ToFieldElement(tt.b)
		assert.Equal(t, tt.b, uint32(tt.a%q))
		assert.Equal(t, tt.b, uint32(a%q))
		assert.Equal(t, b, a)
	}
}

func TestFieldMul(t *testing.T) {
	tests := []struct {
		a uint32
		b uint32
		c uint32
	}{
		{0, 1, 0},
		{2, 5, 10},
		{q >> 1, 2, q - 1},
		{q - 1, 3, q - 3},
		{q, 100, 0},
		{33, 1000, 33000},
		{q, 100000, 0},
	}

	for _, tt := range tests {
		a := common.Uint32ToFieldElement(tt.a)
		b := common.Uint32ToFieldElement(tt.b)
		c := common.Uint32ToFieldElement(tt.c)
		result := common.FieldMul(a, b)
		assert.Equal(t, c, result)
		result = common.FieldMul(b, a)
		assert.Equal(t, c, result)
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
