package common_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"trailofbits.com/ml-dsa/mldsa/common"
)

func TestInt16ToBits(t *testing.T) {
	expected := []byte{0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0}
	assert.Equal(t, expected, common.Int16ToBits(int16(0b01100000_10100000), 16))
}

func TestInt32ToBits(t *testing.T) {
	expected := []byte{0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0}
	assert.Equal(t, expected, common.Int32ToBits(int32(0b01100000_10100000), 16))
}

func TestBitsToInteger(t *testing.T) {
	y := []byte{1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0}
	x := common.BitsToInteger(y, 32)
	assert.Equal(t, uint32(0b01101001_01111001_11111101_00000101), x)
	x = common.BitsToInteger(y, 16)
	assert.Equal(t, uint32(0b11111101_00000101), x)
}

func TestBitsToBytes(t *testing.T) {
	y := []byte{1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0}
	actual := common.BitsToBytes(y)
	expected := []byte{0b00000101, 0b11111101, 0b01111001, 0b01101001}
	assert.Equal(t, expected, actual)
}

func TestBytesToBits(t *testing.T) {
	expected := []byte{1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0}
	b := []byte{0b00000101, 0b11111101, 0b01111001, 0b01101001}
	actual := common.BytesToBits(b)
	assert.Equal(t, expected, actual)
}

func TestCoeffFromThreeBytes(t *testing.T) {
	coeff, err := common.CoeffFromThreeBytes(byte(0xff), byte(0xff), byte(0x01))
	if err != nil {
		panic(err)
	}
	assert.Equal(t, common.Uint32ToFieldElement(uint32(0x01ffff)), coeff)

	// This should be rejection sampled, so expect an error:
	_, err = common.CoeffFromThreeBytes(byte(0xff), byte(0xff), byte(0xff))
	assert.Error(t, err)
}

func TestCoeffFromHalfByte(t *testing.T) {
	// Test that we get an error for an invalid input
	_, err := common.CoeffFromHalfByte(2, byte(15))
	assert.Error(t, err)
	_, err = common.CoeffFromHalfByte(4, byte(9))
	assert.Error(t, err)
	// Test that errors are only produced outside of the expected range
	for i := range 15 {
		c, err := common.CoeffFromHalfByte(2, byte(i))
		assert.Empty(t, err)
		x := int16(2 - (i % 5))
		assert.Equal(t, common.Int16ToRingCoeff(x), c)
	}
	for i := range 9 {
		c, err := common.CoeffFromHalfByte(4, byte(i))
		assert.Empty(t, err)
		x := int16(4 - i)
		assert.Equal(t, common.Int16ToRingCoeff(x), c)
	}
}

// Test SimpleBitPack and SimpleBitUnpack
func TestSimpleBits(t *testing.T) {
	var ringElement common.RingElement
	for i := range 256 {
		ringElement[i] = common.Int16ToRingCoeff(int16(1023 - i))
	}
	packed := common.SimpleBitPack(ringElement, 1023)
	asHex := hex.EncodeToString(packed)
	expected := "fffbdf3ffffbeb9f3ffef7db5f3ffdf3cb1f3ffcefbbdf3efbebab9f3efae79b5f3ef9e38b1f3ef8df7bdf3df7db6b9f3df6d75b5f3df5d34b1f3df4cf3bdf3cf3cb2b9f3cf2c71b5f3cf1c30b1f3cf0bffbde3befbbeb9e3beeb7db5e3bedb3cb1e3becafbbde3aebabab9e3aeaa79b5e3ae9a38b1e3ae89f7bde39e79b6b9e39e6975b5e39e5934b1e39e48f3bde38e38b2b9e38e2871b5e38e1830b1e38e07ffbdd37df7beb9d37de77db5d37dd73cb1d37dc6fbbdd36db6bab9d36da679b5d36d9638b1d36d85f7bdd35d75b6b9d35d6575b5d35d5534b1d35d44f3bdd34d34b2b9d34d2471b5d34d1430b1d34d03ffbdc33cf3beb9c33ce37db5c33cd33cb1c33cc2fbbdc32cb2bab9c32ca279b5c32c9238b1c32c81f7bdc31c71b6b9c31c6175b5c31c5134b1c31c40f3bdc30c30b2b9c30c2071b5c30c1030b1c30c0"
	assert.Equal(t, expected, asHex)

	unpacked := common.SimpleBitUnpack(packed, 1023)
	for i := range 256 {
		expect := common.Int16ToRingCoeff(int16(1023 - i))
		assert.Equal(t, expect, unpacked[i])
	}
}

// Test BitPack and BitUnpack
func TestBits(t *testing.T) {
	var ringElement common.RingElement
	for i := range 256 {
		ringElement[i] = common.Int16ToRingCoeff(int16(i))
	}
	packed := common.BitPack(ringElement, 0, 255)
	asHex := hex.EncodeToString(packed)
	expected := "fffefdfcfbfaf9f8f7f6f5f4f3f2f1f0efeeedecebeae9e8e7e6e5e4e3e2e1e0dfdedddcdbdad9d8d7d6d5d4d3d2d1d0cfcecdcccbcac9c8c7c6c5c4c3c2c1c0bfbebdbcbbbab9b8b7b6b5b4b3b2b1b0afaeadacabaaa9a8a7a6a5a4a3a2a1a09f9e9d9c9b9a999897969594939291908f8e8d8c8b8a898887868584838281807f7e7d7c7b7a797877767574737271706f6e6d6c6b6a696867666564636261605f5e5d5c5b5a595857565554535251504f4e4d4c4b4a494847464544434241403f3e3d3c3b3a393837363534333231302f2e2d2c2b2a292827262524232221201f1e1d1c1b1a191817161514131211100f0e0d0c0b0a09080706050403020100"
	assert.Equal(t, expected, asHex)

	unpacked := common.BitUnpack(packed, 0, 255)
	for i := range 256 {
		expect := common.Int16ToRingCoeff(int16(i))
		assert.Equal(t, expect, unpacked[i])
	}
}

/*
func TestHintPacking(t *testing.T) {
	hintPackingForK(t, uint8(4), uint8(80))
		hintPackingForK(t, uint8(6), uint8(55))
		hintPackingForK(t, uint8(8), uint8(75))
}

func hintPackingForK(t *testing.T, k, omega uint8) {
	vec := common.NewRingVector(k)
	for i := range k {
		x := uint8(0)
		for j := range 256 {
			// Ensure at most omega are nonzero
			if x%64 == k {
				vec[i][j] = common.Int16ToRingCoeff(1)
			}
			x++
		}
	}
	packed := common.HintBitPack(k, omega, vec)
	fmt.Println(hex.EncodeToString(packed))
	unpacked, err := common.HintBitUnpack(k, omega, packed)
	if err != nil {
		panic(err)
	}

	for i := range k {
		x := uint8(0)
		for j := range 256 {
			val := int16(0)
			if x%64 == k {
				val++
			}
			expected := common.Int16ToRingCoeff(val)
			assert.Equal(t, expected, unpacked[i][j])
			x++
		}
	}
}

func TestPKEncode(t *testing.T) {
	expected, err := hex.DecodeString("01")
	if err != nil {
		panic(err)
	}
	testPkEncodeInternal(t, uint8(4), expected)
	expected, err = hex.DecodeString("02")
	if err != nil {
		panic(err)
	}
	testPkEncodeInternal(t, uint8(6), expected)
	expected, err = hex.DecodeString("03")
	if err != nil {
		panic(err)
	}
	testPkEncodeInternal(t, uint8(8), expected)
}

func testPkEncodeInternal(t *testing.T, k uint8, expected []byte) {
	seed, err := hex.DecodeString("f696484048ec21f96cf50a56d0759c448f3779752f0383d37449690694cf7a68")
	if err != nil {
		panic(err)
	}
	rv := common.NewRingVector(k)
	x := int16(0)
	for i := range k {
		for j := range 256 {
			rv[i][j] = common.Int16ToRingCoeff(x)
			x++
		}
	}
	actual := common.PKEncode(k, seed, rv)
	decoded, newRV := common.PKDecode(k, actual)
	assert.Equal(t, seed, decoded)
	for i := range k {
		for j := range 256 {
			assert.Equal(t, rv[i][j], newRV[i][j])
		}
	}

	assert.Equal(t, hex.EncodeToString(expected), hex.EncodeToString(actual))
	// assert.Equal(t, expected, actual)
}
*/

func TestSampleInBallExactNonzero(t *testing.T) {
	seed, err := hex.DecodeString("f696484048ec21f96cf50a56d0759c448f3779752f0383d37449690694cf7a68")
	if err != nil {
		panic(err)
	}
	tauMap := []int{39, 49, 60}
	for _, tau := range tauMap {
		sampled := common.SampleInBall(uint8(tau), seed)
		found := 0
		for i := range 256 {
			if sampled[i] != 0 {
				found++
			}
		}
		assert.Equal(t, tau, found)
	}
}

func TestRejNTTPoly(t *testing.T) {
	seed, err := hex.DecodeString("f696484048ec21f96cf50a56d0759c448f3779752f0383d37449690694cf7a68")
	if err != nil {
		panic(err)
	}
	poly := common.RejNTTPoly(seed)
	for i := range 256 {
		assert.Less(t, uint32(poly[i]), uint32(q))
	}
}

func TestRejBoundedPoly(t *testing.T) {
	seed, err := hex.DecodeString("f696484048ec21f96cf50a56d0759c448f3779752f0383d37449690694cf7a68")
	if err != nil {
		panic(err)
	}
	etaMap := []uint8{2, 4}
	for _, eta := range etaMap {
		poly := common.RejBoundedPoly(int(eta), seed)
		plus := int32(eta) + 1
		minus := -plus

		// Ensure all values are between -eta and +eta
		for i := range 256 {
			assert.Greater(t, int32(poly[i]), minus)
			assert.Less(t, int32(poly[i]), plus)
		}
	}
}

func TestExpandA(t *testing.T) {
	testExpandAParametrized(t, uint8(4), uint8(4))
	testExpandAParametrized(t, uint8(6), uint8(5))
	testExpandAParametrized(t, uint8(8), uint8(7))
}

func testExpandAParametrized(t *testing.T, k, l uint8) {
	seed, err := hex.DecodeString("f696484048ec21f96cf50a56d0759c448f3779752f0383d37449690694cf7a68")
	if err != nil {
		panic(err)
	}
	expanded := common.ExpandA(k, l, seed)
	assert.Equal(t, int(k), len(expanded))
	for i := range k {
		assert.Equal(t, int(l), len(expanded[i]))
		for j := range l {
			assert.Equal(t, 256, len(expanded[i][j]))
			for x := range 256 {
				assert.Less(t, uint32(expanded[i][j][x]), uint32(q))
			}
		}
	}
}

func TestExpandS(t *testing.T) {
	testExpandSParametrized(t, uint8(4), uint8(4), int(2))
	testExpandSParametrized(t, uint8(6), uint8(5), int(4))
	testExpandSParametrized(t, uint8(8), uint8(7), int(2))
}

func testExpandSParametrized(t *testing.T, k, l uint8, eta int) {
	tr, err := hex.DecodeString("97e92fdc87c9f838ea6d1c24a38b07b3a7adf78662495a2dcb64f05bc5f3031332254755fc6090c4c4c3dc08d1d5fab033a2635fd94b71e953f5a58b3695313d")
	if err != nil {
		panic(err)
	}
	s1, s2 := common.ExpandS(k, l, eta, tr)
	plus := eta + 1
	minus := -1 * plus
	for r := range l {
		for i := range 256 {
			assert.Greater(t, int(s1[r][i]), minus)
			assert.Less(t, int(s1[r][i]), plus)
		}
	}
	for s := range k {
		for i := range 256 {
			assert.Greater(t, int(s2[s][i]), minus)
			assert.Less(t, int(s2[s][i]), plus)
		}
	}
}

func TestExpandMask(t *testing.T) {
	testExpandMaskParametrized(t, uint8(4), uint32(131072), uint64(0xFFFEFDFC_FBFAF9F8))
	testExpandMaskParametrized(t, uint8(5), uint32(524288), uint64(0xF7F6F5F4_F3F2F1F0))
	testExpandMaskParametrized(t, uint8(7), uint32(524288), uint64(0x12345678_9ABCDEF0))
}

func testExpandMaskParametrized(t *testing.T, l uint8, gamma1 uint32, mu uint64) {
	seed, err := hex.DecodeString("97e92fdc87c9f838ea6d1c24a38b07b3a7adf78662495a2dcb64f05bc5f3031332254755fc6090c4c4c3dc08d1d5fab033a2635fd94b71e953f5a58b3695313d")
	if err != nil {
		panic(err)
	}
	mask := common.ExpandMask(l, gamma1, seed, mu)
	upper := int32(gamma1 + 1)
	lower := -upper + 1
	for i := range l {
		for j := range 256 {
			assert.Greater(t, int32(mask[i][j]), lower)
			assert.Less(t, int32(mask[i][j]), upper)
		}
	}
}

func TestPower2Round(t *testing.T) {
	tests := []struct {
		a      uint32
		wantA1 int16
		wantA0 int16
	}{
		{123456, 15, 576},
		{8192, 1, 0},    // Exact multiple of 2^13
		{8193, 1, 1},    // Just above 2^13
		{4095, 0, 4095}, // Below rounding threshold
		{122880, 15, 0}, // Exact multiple of 2^13
	}
	for _, tt := range tests {
		gotA1, gotA0 := common.Power2Round(tt.a)
		assert.Equal(t, tt.wantA1, gotA1)
		assert.Equal(t, tt.wantA0, gotA0)
	}
}

func TestDecompose(t *testing.T) {
	tests := []struct {
		name      string
		gamma2    uint32
		r         int32
		expected1 int32
		expected2 int32
	}{
		{"Zero Input (gamma2=95232)", 95232, 0, 0, 0},
		{"Positive r (gamma2=95232)", 95232, 5, 0, 5},
		{"r = 2 * gamma2 (gamma2=95232)", 95232, 190464, 1, 0},
		{"Zero Input (gamma2=261888)", 261888, 0, 0, 0},
		{"Positive r (gamma2=261888)", 261888, 5, 0, 5},
		{"r = 2 * gamma2 (gamma2=261888)", 261888, 261888 << 1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res1, res2 := common.DecomposeVarTime(tt.gamma2, tt.r)
			if res1 != tt.expected1 || res2 != tt.expected2 {
				t.Errorf("Decompose(%d, %d) = (%d, %d); want (%d, %d)",
					tt.gamma2, tt.r, res1, res2, tt.expected1, tt.expected2)
			}
			res1, res2 = common.Decompose(tt.gamma2, tt.r)
			if res1 != tt.expected1 || res2 != tt.expected2 {
				t.Errorf("Decompose(%d, %d) = (%d, %d); want (%d, %d)",
					tt.gamma2, tt.r, res1, res2, tt.expected1, tt.expected2)
			}
		})
	}
}

func TestMakeUseHint(t *testing.T) {
	tests := []struct {
		gamma2    uint32
		z         uint32
		r         uint32
		h         uint8
		recovered uint32
	}{
		{95232, 0, 0, 0, 0},
		{95232, 47616, 0, 0, 0},
		{95232, 190463, 1, 1, 0},
		{95232, 100, 190364, 1, 0},
		{95232, 0, 190464, 0, 1},
		{261888, 0, 0, 0, 0},
	}

	for _, tt := range tests {
		r := common.Uint32ToFieldElement(tt.r)
		z := common.Uint32ToFieldElement(tt.z)
		gotH := common.MakeHint(tt.gamma2, z, r)
		if tt.h != gotH {
			fmt.Printf("%d %d | %d != %d\n", tt.z, tt.r, tt.h, gotH)
		}
		assert.Equal(t, tt.h, gotH)
		used := common.UseHint(tt.gamma2, tt.h, r+z)
		recovered := common.Uint32ToFieldElement(tt.recovered)
		if used != recovered {
			fmt.Printf("%d %d - recovered: %d != %d\n", tt.z, tt.r, used, recovered)
		}
		assert.Equal(t, recovered, used)
	}
}
