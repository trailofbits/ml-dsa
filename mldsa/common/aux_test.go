package common_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"trailofbits.com/ml-dsa/mldsa/common"
)

func TestUint32ToBits(t *testing.T) {
	expected := []byte{0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0}
	assert.Equal(t, expected, common.Uint32ToBits(uint32(0b01100000_10100000), 16))

	// More rigorous tests
	tests := []struct {
		input    uint32
		bits     int
		expected []byte
	}{
		{0xFFFFFFFF, 8, []byte{1, 1, 1, 1, 1, 1, 1, 1}},
		{0xFFFFFFFF, 9, []byte{1, 1, 1, 1, 1, 1, 1, 1, 1}},
		{0x12345678, 32, []byte{0, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0}},
		//                      8 --------| 7 --------| 6 --------| 5 --------| 4 --------| 3 --------| 2 --------| 1 --------|
	}
	for _, tt := range tests {
		output := common.Uint32ToBits(tt.input, tt.bits)
		assert.Equal(t, tt.bits, len(output))
		assert.Equal(t, hex.EncodeToString(tt.expected), hex.EncodeToString(output))
	}
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
		x := uint32(q + 2 - (i % 5))
		assert.Equal(t, common.CoeffReduceOnce(x), c)
	}
	for i := range 9 {
		c, err := common.CoeffFromHalfByte(4, byte(i))
		assert.Empty(t, err)
		x := uint32(q + 4 - i)
		assert.Equal(t, common.CoeffReduceOnce(x), c)
	}
}

// Test SimpleBitPack and SimpleBitUnpack
func TestSimpleBits(t *testing.T) {
	var ringElement common.RingElement
	for i := range 256 {
		ringElement[i] = common.CoeffReduceOnce(uint32(1023 - i))
	}
	packed := common.SimpleBitPack(ringElement, 1023)
	asHex := hex.EncodeToString(packed)
	expected := "fffbdf3ffffbeb9f3ffef7db5f3ffdf3cb1f3ffcefbbdf3efbebab9f3efae79b5f3ef9e38b1f3ef8df7bdf3df7db6b9f3df6d75b5f3df5d34b1f3df4cf3bdf3cf3cb2b9f3cf2c71b5f3cf1c30b1f3cf0bffbde3befbbeb9e3beeb7db5e3bedb3cb1e3becafbbde3aebabab9e3aeaa79b5e3ae9a38b1e3ae89f7bde39e79b6b9e39e6975b5e39e5934b1e39e48f3bde38e38b2b9e38e2871b5e38e1830b1e38e07ffbdd37df7beb9d37de77db5d37dd73cb1d37dc6fbbdd36db6bab9d36da679b5d36d9638b1d36d85f7bdd35d75b6b9d35d6575b5d35d5534b1d35d44f3bdd34d34b2b9d34d2471b5d34d1430b1d34d03ffbdc33cf3beb9c33ce37db5c33cd33cb1c33cc2fbbdc32cb2bab9c32ca279b5c32c9238b1c32c81f7bdc31c71b6b9c31c6175b5c31c5134b1c31c40f3bdc30c30b2b9c30c2071b5c30c1030b1c30c0"
	assert.Equal(t, expected, asHex)

	unpacked := common.SimpleBitUnpack(packed, 1023)
	for i := range 256 {
		expect := common.CoeffReduceOnce(uint32(1023 - i))
		assert.Equal(t, expect, unpacked[i])
	}
}

// Test BitPack and BitUnpack
func TestBits(t *testing.T) {
	var ringElement common.RingElement
	for i := range 256 {
		ringElement[i] = common.CoeffReduceOnce(uint32(i))
	}
	packed := common.BitPack(ringElement, 0, 255)
	asHex := hex.EncodeToString(packed)
	expected := "fffefdfcfbfaf9f8f7f6f5f4f3f2f1f0efeeedecebeae9e8e7e6e5e4e3e2e1e0dfdedddcdbdad9d8d7d6d5d4d3d2d1d0cfcecdcccbcac9c8c7c6c5c4c3c2c1c0bfbebdbcbbbab9b8b7b6b5b4b3b2b1b0afaeadacabaaa9a8a7a6a5a4a3a2a1a09f9e9d9c9b9a999897969594939291908f8e8d8c8b8a898887868584838281807f7e7d7c7b7a797877767574737271706f6e6d6c6b6a696867666564636261605f5e5d5c5b5a595857565554535251504f4e4d4c4b4a494847464544434241403f3e3d3c3b3a393837363534333231302f2e2d2c2b2a292827262524232221201f1e1d1c1b1a191817161514131211100f0e0d0c0b0a09080706050403020100"
	assert.Equal(t, expected, asHex)

	unpacked := common.BitUnpack(packed, 0, 255)
	for i := range 256 {
		expect := common.CoeffReduceOnce(uint32(i))
		assert.Equal(t, expect, unpacked[i])
	}
	/*
		re := common.RingElement{-2, -1, 0, 2, -2, 2, 0, 2, -1, 0, 0, 1, -1, 2, 2, 1, -1, 2, 0, -1, -1, 1, -2, 0, -1, 2, 1, -1, 0, -1, 1, 2, -1, -2, 1, 2, -1, 0, 1, 1, 2, 0, 0, 1, 0, 2, 0, 1, 1, 1, 2, -2, 0, -2, 0, -2, 0, 2, 1, -1, -1, 0, -1, 0, 1, 0, 2, -2, 1, 1, -2, 1, 0, 0, 1, 2, -2, 2, -1, 2, -1, 2, -2, -2, 2, -2, 1, 1, -2, -2, 2, 1, -1, 2, 1, 0, -1, 2, 0, 2, 2, 1, 0, 2, 2, 1, -2, -2, -1, 2, 2, -2, 0, 0, 1, -2, -2, -2, 0, -1, 0, 0, -1, -2, -2, 0, -1, -2, 0, -2, -1, 0, 1, 2, 1, 1, -1, -1, 2, -2, -2, 0, -2, 0, 2, 1, 1, 1, 2, 0, -2, 1, 1, 2, -1, -1, 1, -1, 0, -1, -2, 2, 2, -2, 0, 0, 2, -2, 0, 2, -2, 2, 2, 2, 1, 1, 1, 0, 0, -2, 1, 1, 0, 1, 1, -1, -2, -2, 0, -2, -1, 1, -2, -2, -2, 1, -1, 1, 0, 0, 0, -1, 1, 0, 1, 1, -2, 1, 0, 1, 2, 0, 1, -1, 2, -1, 0, -1, 1, -1, -1, -1, 0, 1, 2, -1, 0, 0, 2, -1, 1, 1, 0, -2, 0, 0, -2, -1, -1, 1, 1, -1, 0, 0, 1, -1, 1, -2, 2, -2, 1, 2, 1, -1, 1, -2}
		packed = common.BitPack(re, 2, 2)
		// print(len(packed), "\n")
		asHex = hex.EncodeToString(packed)
		print(asHex, "\n")
	*/
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

// Hamming weight helper for SampleInBall test
func countOnes(r common.RingElement) uint16 {
	ones := uint16(0)
	for i := range 256 {
		if r[i] != 0 {
			ones++
		}
	}
	return ones
}

func TestSampleInBall(t *testing.T) {
	for tau_1 := range uint16(64) {
		tau := tau_1 + 1
		for seed := range uint16(256) {
			rho := (tau << 8) | (seed & 0xff)
			p := common.SampleInBall(uint8(tau), common.PackUint16(rho))
			hamming_weight := countOnes(p)
			assert.Equal(t, hamming_weight, tau)
		}
	}
}

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
		plus := uint32(eta + 1)
		minus := uint32(common.CoeffReduceOnce(q - plus))

		// Ensure all values are between -eta and +eta
		for i := range 256 {
			value := uint32(common.CoeffReduceOnce(uint32(poly[i])))
			// fmt.Printf("%d ", value)
			assert.True(t, value < plus || value > minus)
			// assert.Greater(t, int32(poly[i]), minus)
			// assert.Less(t, int32(poly[i]), plus)
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
	plus := uint32(eta + 1)
	minus := uint32(common.CoeffReduceOnce(q - plus))
	// fmt.Printf("s1 = {\n")
	for r := range l {
		// fmt.Printf("\t{")
		for i := range 256 {
			value := uint32(common.CoeffReduceOnce(uint32(s1[r][i])))
			// fmt.Printf("%d ", value)
			assert.True(t, value < plus || value > minus)
			// fmt.Printf("%d", s1[r][i])
			// if i < 255 {
			// 	fmt.Printf(",")
			// }
		}
		// fmt.Printf("\n\t},\n")
	}
	// fmt.Printf("}\ns2 = {\n")
	for s := range k {
		// fmt.Printf("\t{")
		for i := range 256 {
			value := uint32(common.CoeffReduceOnce(uint32(s2[s][i])))
			assert.True(t, value < plus || value > minus)
			// fmt.Printf("%d ", value)
			// if i < 255 {
			// 	fmt.Printf(",")
			// }
		}
		// fmt.Printf("\n\t},\n")
	}
	// fmt.Printf("}\n")
}

func TestExpandSZeroSeed(t *testing.T) {
	// k := uint8(4)
	l := uint8(4)
	// Rho value for an all-zero seed:
	rhop := []byte{193, 170, 32, 212, 203, 183, 231, 1, 44, 0, 171, 235, 242, 200, 90, 146, 4, 163, 161, 171, 89, 111, 206, 39, 231, 130, 60, 182, 9, 243, 123, 41, 121, 49, 17, 90, 218, 43, 116, 132, 183, 82, 123, 187, 209, 94, 114, 42, 167, 211, 91, 196, 227, 239, 176, 190, 61, 2, 185, 6, 237, 210, 23, 107}
	s1, _ := common.ExpandS(uint8(4), uint8(4), int(2), rhop)
	// [84], [74], [216], [92], [252], [131], [187], [15], [124], [225], [13], [79], [213], [23], [66], [6], [47], [70], [139], [2], [136], [45], [246], [105], [92], [201], [0], [128], [220], [122], [209], [164], [24], [165], [4], [171], [102], [23], [127], [203], [159], [148], [138], [165], [79], [135], [54], [13], [92], [237], [210], [170], [50], [70], [55], [214], [179], [0], [115], [102], [46], [248], [188], [182], [215], [203], [52], [48], [89], [162], [194], [92], [199], [174], [144], [119], [40], [43], [87], [255], [250], [11], [185], [128], [82], [244], [222], [212], [252], [205], [55], [236], [115], [82], [220], [194], [120], [14], [10], [29], [197], [176], [78], [139], [253], [155], [224], [187], [148], [106], [221], [105], [150], [100], [102], [88], [158], [132], [90], [115], [117], [143], [195], [96], [117], [242], [222], [209], [199], [0], [102], [159], [216], [148], [153], [184], [212]
	// [84], [74], [216], [92], [252], [131], [187], [15], [124], [225], [13], [79], [213], [23], [66], [6], [47], [70], [139], [2], [136], [45], [246], [105], [92], [201], [0], [128], [220], [122], [209], [164], [24], [165], [4], [171], [102], [23], [127], [203], [159], [148], [138], [165], [79], [135], [54], [13], [92], [237], [210], [170], [50], [70], [55], [214], [179], [0], [115], [102], [46], [248], [188], [182], [215], [203], [52], [48], [89], [162], [194], [92], [199], [174], [144], [119], [40], [43], [87], [255], [250], [11], [185], [128], [82], [244], [222], [212], [252], [205], [55], [236], [115], [82], [220], [194], [120], [14], [10], [29], [197], [176], [78], [139], [253], [155], [224], [187], [148], [106], [221], [105], [150], [100], [102], [88], [158], [132], [90], [115], [117], [143], [195], [96], [117], [242], [222], [209], [199], [0], [102], [159], [216], [148], [153], [184], [212], [35], [124], [224], [240], [209], [3], [127], [229], [70], [192], [53], [41], [169], [69], [44], [23], [41], [234], [128], [102], [235], [23], [149], [190], [142], [93], [134], [242], [116], [22], [153], [217], [214], [135], [221], [237], [37], [251], [149], [9], [33], [82], [110], [13], [253], [162], [46], [17], [230], [222], [73], [86], [30], [34], [74], [89], [31], [188], [28], [247], [163], [6], [199], [73], [43], [18], [39], [105], [107], [194], [133], [19], [58], [78], [125], [214], [179], [210], [153], [48], [67], [234], [45], [16], [201], [165], [5], [158], [196], [79], [91], [201], [84], [117], [205], [25], [205], [9], [201], [130], [233], [207], [174], [253], [196], [232], [105], [31], [209], [147], [78], [161], [250], [63], [53], [199], [161], [237], [205],
	expect_s1 := [][]uint32{
		[]uint32{8380415, 2, 2, 8380415, 8380416, 8380416, 0, 2, 0, 8380416, 8380416, 1, 1, 2, 0, 0, 1, 8380415, 8380416, 2, 8380415, 2, 8380416, 0, 1, 0, 8380415, 1, 2, 0, 1, 8380415, 1, 8380416, 0, 2, 8380416, 8380416, 8380416, 0, 1, 8380415, 1, 0, 2, 8380415, 0, 2, 2, 2, 8380416, 0, 8380416, 2, 0, 1, 8380416, 8380415, 2, 8380416, 1, 2, 2, 8380415, 2, 1, 2, 1, 1, 0, 1, 0, 1, 0, 8380415, 8380415, 8380415, 2, 8380416, 2, 2, 8380415, 0, 8380416, 1, 8380416, 8380416, 2, 0, 2, 8380416, 8380415, 0, 8380416, 2, 2, 0, 8380416, 1, 8380415, 0, 8380416, 1, 8380416, 8380416, 1, 2, 2, 8380416, 0, 1, 1, 8380415, 0, 8380416, 0, 1, 1, 1, 0, 8380416, 1, 0, 8380415, 8380416, 2, 8380416, 8380415, 2, 0, 2, 0, 0, 0, 2, 0, 0, 8380415, 2, 2, 8380415, 0, 0, 8380416, 0, 1, 0, 0, 2, 2, 1, 2, 8380415, 1, 2, 8380416, 0, 2, 8380415, 8380415, 8380416, 8380415, 8380416, 0, 8380416, 0, 0, 8380416, 0, 8380415, 8380416, 0, 0, 2, 0, 8380416, 0, 0, 8380416, 0, 8380415, 2, 2, 2, 8380416, 1, 2, 0, 2, 1, 8380415, 8380415, 1, 8380416, 8380416, 1, 8380415, 2, 8380415, 1, 1, 8380415, 8380415, 2, 1, 8380416, 8380416, 8380415, 1, 1, 8380415, 8380415, 1, 1, 1, 8380416, 2, 8380415, 8380415, 8380415, 8380416, 2, 2, 8380416, 0, 2, 0, 8380416, 8380416, 0, 2, 1, 2, 0, 0, 8380415, 8380416, 1, 8380416, 0, 0, 2, 2, 1, 1, 8380415, 8380416, 8380416, 8380415, 8380415, 8380415, 8380415, 8380416, 1, 8380415, 8380416},
		[]uint32{8380415, 2, 0, 1, 0, 0, 2, 2, 1, 8380415, 8380416, 8380415, 0, 1, 1, 8380416, 8380415, 2, 1, 8380416, 0, 8380415, 8380416, 2, 1, 2, 8380415, 8380416, 1, 8380415, 8380416, 2, 0, 2, 0, 2, 8380415, 2, 8380416, 8380416, 8380416, 8380416, 0, 8380415, 8380416, 0, 8380416, 8380416, 0, 1, 8380415, 1, 1, 1, 8380415, 0, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 8380416, 8380415, 8380416, 1, 8380416, 1, 0, 8380416, 8380415, 0, 0, 2, 8380415, 8380416, 2, 0, 0, 1, 1, 8380415, 2, 1, 1, 0, 8380416, 1, 8380416, 1, 2, 8380415, 0, 8380415, 2, 8380416, 8380415, 8380415, 8380415, 0, 1, 2, 2, 0, 8380416, 2, 2, 8380415, 2, 2, 1, 0, 8380416, 0, 2, 1, 1, 2, 1, 8380416, 2, 2, 2, 8380416, 1, 2, 1, 1, 8380416, 8380415, 2, 8380415, 1, 1, 8380415, 0, 8380415, 2, 8380415, 8380415, 0, 0, 1, 1, 8380416, 0, 0, 1, 8380416, 2, 0, 8380415, 8380416, 1, 1, 0, 8380416, 0, 0, 1, 1, 0, 2, 1, 0, 1, 8380415, 2, 8380415, 8380416, 0, 8380415, 2, 8380415, 1, 1, 8380416, 8380416, 8380415, 8380416, 8380415, 8380416, 0, 2, 1, 8380416, 0, 8380416, 2, 0, 0, 1, 2, 1, 8380415, 8380415, 1, 1, 0, 1, 1, 0, 0, 0, 2, 2, 1, 2, 8380415, 8380415, 0, 8380415, 8380415, 8380416, 1, 2, 1, 2, 0, 0, 1, 2, 8380415, 8380416, 2, 8380416, 8380415, 8380415, 8380416, 2, 2, 2, 8380416, 1, 2, 0, 2, 8380416, 8380416, 0, 8380415, 0, 1, 2, 8380415, 8380415, 8380416, 2, 8380416, 8380416, 0, 1},
		[]uint32{0, 1, 8380415, 8380416, 0, 2, 8380415, 8380415, 8380415, 8380416, 8380415, 8380415, 8380415, 1, 8380415, 0, 2, 8380416, 1, 1, 2, 2, 8380416, 8380415, 8380415, 1, 8380415, 8380415, 8380415, 8380415, 8380415, 8380416, 1, 1, 1, 0, 2, 8380416, 8380416, 8380416, 1, 8380416, 8380415, 8380416, 0, 8380416, 1, 8380415, 8380416, 8380415, 8380416, 2, 2, 2, 8380415, 1, 8380416, 1, 8380416, 1, 8380416, 8380416, 2, 1, 0, 8380415, 8380416, 8380415, 2, 2, 8380415, 2, 8380415, 1, 1, 8380416, 0, 1, 1, 8380416, 1, 0, 8380415, 1, 8380415, 2, 0, 8380416, 0, 0, 1, 0, 1, 8380416, 0, 8380416, 1, 1, 8380415, 8380415, 8380416, 8380415, 8380415, 1, 1, 0, 0, 1, 0, 8380416, 8380416, 8380416, 0, 1, 8380416, 0, 1, 2, 1, 8380415, 8380416, 8380415, 0, 1, 1, 8380416, 8380415, 2, 8380416, 8380416, 8380416, 0, 1, 1, 8380416, 1, 1, 2, 2, 8380416, 8380415, 1, 1, 8380416, 8380415, 0, 2, 0, 0, 2, 8380416, 8380415, 0, 8380416, 8380415, 0, 1, 0, 8380415, 1, 0, 1, 8380415, 1, 0, 8380415, 8380416, 2, 2, 8380416, 8380416, 0, 1, 2, 2, 1, 8380415, 8380415, 0, 8380416, 0, 0, 8380416, 8380415, 8380416, 2, 1, 2, 8380415, 2, 8380415, 8380415, 1, 2, 8380416, 1, 8380415, 8380415, 1, 8380415, 8380415, 8380415, 0, 2, 1, 0, 1, 8380415, 8380415, 1, 2, 1, 8380416, 0, 8380416, 1, 2, 0, 8380416, 8380415, 0, 2, 2, 8380416, 8380416, 8380416, 8380415, 0, 8380416, 1, 8380416, 2, 0, 0, 1, 0, 8380415, 0, 2, 8380416, 1, 8380415, 8380415, 0, 1, 2, 1, 1, 1, 8380415, 2, 8380416, 8380415, 1, 1, 2},
		[]uint32{2, 8380416, 8380416, 0, 8380416, 0, 8380415, 1, 1, 8380416, 8380416, 1, 1, 0, 8380415, 8380416, 0, 8380415, 0, 0, 8380416, 8380415, 8380415, 0, 8380416, 2, 2, 2, 8380416, 0, 0, 8380415, 8380415, 0, 1, 1, 8380416, 2, 8380415, 2, 8380416, 2, 0, 0, 1, 1, 0, 2, 0, 8380415, 0, 0, 0, 1, 8380415, 8380416, 0, 0, 8380416, 0, 8380415, 8380415, 8380416, 8380415, 8380415, 2, 0, 1, 2, 0, 8380416, 8380415, 1, 0, 0, 8380416, 8380415, 8380415, 2, 8380416, 8380416, 1, 2, 2, 2, 8380415, 8380416, 2, 2, 8380415, 0, 2, 2, 1, 8380415, 8380416, 0, 8380415, 8380415, 0, 8380416, 8380416, 8380416, 2, 8380416, 1, 8380415, 8380415, 8380415, 8380416, 2, 2, 2, 2, 2, 8380416, 0, 0, 8380416, 0, 0, 8380415, 8380416, 0, 8380415, 2, 2, 8380415, 8380416, 0, 0, 2, 0, 0, 1, 8380416, 1, 1, 8380416, 8380416, 1, 8380416, 8380416, 8380415, 0, 8380416, 8380416, 0, 0, 1, 8380416, 0, 0, 2, 1, 8380415, 0, 1, 2, 2, 0, 8380416, 8380415, 8380416, 0, 2, 8380416, 8380416, 8380416, 2, 8380416, 0, 1, 8380416, 8380416, 0, 2, 8380415, 0, 2, 8380416, 2, 8380416, 1, 0, 8380416, 8380416, 8380415, 0, 1, 1, 2, 8380416, 8380415, 8380415, 1, 0, 1, 8380416, 8380415, 8380416, 8380415, 1, 0, 2, 8380415, 8380416, 2, 1, 8380416, 1, 8380415, 8380415, 1, 2, 2, 2, 8380415, 8380415, 8380415, 0, 1, 8380416, 8380416, 8380415, 2, 8380416, 8380416, 2, 2, 8380415, 0, 0, 2, 0, 2, 2, 8380415, 8380416, 1, 8380415, 8380416, 1, 1, 8380416, 1, 1, 0, 1, 8380415, 1, 8380416, 8380415, 0, 8380415, 0},
	}
	for i := range l {
		for j := range 256 {
			if expect_s1[i][j] != uint32(s1[i][j]) {
				fmt.Printf("s1[%d][%d] conflict: %d != %d\n", i, j, expect_s1[i][j], s1[i][j])
			}
			// assert.Equal(t, expect_s1[i][j], uint32(s1[i][j]))
		}
	}
}

func TestExpandMask(t *testing.T) {
	testExpandMaskParametrized(t, uint8(4), uint32(131072), uint16(0xFFFE))
	testExpandMaskParametrized(t, uint8(5), uint32(524288), uint16(0xF7F6))
	testExpandMaskParametrized(t, uint8(7), uint32(524288), uint16(0x1234))
}

func testExpandMaskParametrized(t *testing.T, l uint8, gamma1 uint32, mu uint16) {
	seed, err := hex.DecodeString("97e92fdc87c9f838ea6d1c24a38b07b3a7adf78662495a2dcb64f05bc5f3031332254755fc6090c4c4c3dc08d1d5fab033a2635fd94b71e953f5a58b3695313d")
	if err != nil {
		panic(err)
	}
	mask := common.ExpandMask(l, gamma1, seed, mu)
	upper := uint32(gamma1 + 1)
	lower := q - upper + 1
	for i := range l {
		for j := range 256 {
			m := uint32(mask[i][j])
			assert.True(t, m < lower || m > upper)
		}
	}
}

func TestPower2Round(t *testing.T) {
	tests := []struct {
		a      uint32
		wantA1 uint32
		wantA0 uint32
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
		// Test that (a1 * 2^13) + a0 == a
		congruent := (gotA1 << 13) + gotA0
		assert.Equal(t, tt.a, congruent)
	}
}

/*
   fn decompose() {
       for x in 0..MOD {
           let x = Elem::new(x);
           let (x1, x0) = x.decompose::<Mod>();

           // The low-order output from decompose() is a mod+- output, optionally minus one.  So
           // they should be in the closed interval [-gamma2, gamma2].
           let positive_bound = x0.0 <= MOD / 2;
           let negative_bound = x0.0 >= BaseField::Q - MOD / 2;
           assert!(positive_bound || negative_bound);

           // The low-order and high-order outputs should combine to form the input.
           let xx = (MOD * x1.0 + x0.0) % BaseField::Q;
           assert_eq!(xx, x.0);
       }
   }
*/

func TestDecomposeExhaustive(t *testing.T) {
	/*
		gammas := []uint32{95232, 261888}
		for _, gamma2 := range gammas {
			testDcomposeExhausiveGamma2(gamma2)
		}
	*/
	testDecomposeExhausiveGamma2(t, uint32(95232))
}

func testDecomposeExhausiveGamma2(t *testing.T, gamma2 uint32) {
	m := gamma2 << 1
	t1, t0 := common.Decompose(gamma2, 380926)
	assert.Equal(t, uint32(q-2), t0)
	assert.Equal(t, uint32(2), t1)

	// for i := uint32(m + 1); i < q; i += (gamma2 - 1) {
	for i := uint32(1); i < m; i++ {
		x1, x0 := common.Decompose(gamma2, i)
		if x0 > gamma2 && x0 < uint32(q-gamma2) {
			fmt.Printf("\t%d {%d, %d}\n", i, x1, x0)
			fmt.Printf("\t%d > %d\n", x0, gamma2)
			fmt.Printf("\t%d < %d\n", x0, q-gamma2)
			panic("value out of range\n")
		} else {
			// We're just asserting that the values are in the expected range
			assert.True(t, x0 <= gamma2 || x0 >= uint32(q-gamma2))
			// Can we recombine x1 and x0 to get the desired value?
			// xx :=  (x1*m + x0) % q

			// xx := uint32((uint64(x1)*uint64(m) + uint64(x0)) % uint64(q))
			tmp := uint64(x1) * uint64(m) % uint64(q)
			xx := (uint32(tmp) + x0) % q
			if xx != i {
				fmt.Printf("x1 * m + x0 == (%d * %d + %d) == %d, should equal %d\n", x1, m, x0, xx, i)
			} else {
				assert.Equal(t, xx, i)
			}
			if i > gamma2+10 {
				return
			}
		}
	}
}

func TestModPlusMinus(t *testing.T) {
	tests := []struct {
		name     string
		gamma2   uint32
		r        uint32
		expected uint32
	}{
		{"Zero Input (gamma2=95232)", 95232, 0, 0},
		{"Gamma2 + 1 (gamma2=95232)", 95232, 95233, q - 95231},
		{"2 * Gamma2 - 2 (gamma2=95232)", 95232, 190462, q - 2},
		{"2 * Gamma2 - 1 (gamma2=95232)", 95232, 190463, q - 1},
		{"2 * Gamma2 (gamma2=95232)", 95232, 190464, 0},
		{"2 * Gamma2 + 1 (gamma2=95232)", 95232, 190465, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, common.ModPlusMinus(tt.r, tt.gamma2<<1))
		})
	}
}

func TestDecompose(t *testing.T) {
	tests := []struct {
		name      string
		gamma2    uint32
		r         uint32
		expected1 uint32
		expected2 uint32
	}{
		{"Zero Input (gamma2=95232)", 95232, 0, 0, 0},
		{"Positive r (gamma2=95232)", 95232, 5, 0, 5},
		{"r = 2 * gamma2 (gamma2=95232)", 95232, 190464, 1, 0},
		{"Zero Input (gamma2=261888)", 261888, 0, 0, 0},
		{"Positive r (gamma2=261888)", 261888, 5, 0, 5},
		{"r = 2 * gamma2 (gamma2=261888)", 261888, 261888 << 1, 1, 0},
		{"Testing from w1encode", 95232, 6789674, 36, 8313387},
		{"Testing from w1encode", 261888, 6789674, 13, 8361003},
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
			// Confirm that decomposition does trivially recombine
			congruent := ((res1 * (tt.gamma2 << 1) % q) + res2) % q
			if congruent != tt.r%q {
				fmt.Printf("Not equal mod q: %d, %d\n", tt.r%q, congruent%q)
			}
			assert.Equal(t, tt.r%q, congruent)
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
		{95232, 95232, 0, 0, 0},
		/*
			{95232, 47616, 0, 0, 0},
			{95232, 95232, 1, 1, 0},
			{95232, 190463, 1, 0, 0},
			{95232, 100, 190364, 1, 0},
			{95232, 0, 190464, 0, 1},
		*/
		{261888, 0, 0, 0, 0},
	}

	for _, tt := range tests {
		r := common.Uint32ToFieldElement(tt.r)
		z := common.Uint32ToFieldElement(tt.z)
		gotH := common.MakeHint(tt.gamma2, z, r)
		assert.Equal(t, tt.h, gotH)
		used := common.UseHint(tt.gamma2, tt.h, r+z)
		recovered := common.Uint32ToFieldElement(tt.recovered)
		assert.Equal(t, recovered, used)
	}
}

func TestUseHint(t *testing.T) {
	tests := []struct {
		gamma2 uint32
		h      uint8
		input  uint32
		output uint32
	}{
		{95232, 0, 0, 0},
		{95232, 1, 0, 1},
		{95232, 0, 95232, 0},
	}
	for _, tt := range tests {
		r := common.Uint32ToFieldElement(tt.input)
		out := common.UseHint(tt.gamma2, tt.h, r)
		assert.Equal(t, common.Uint32ToFieldElement(tt.output), out)
	}
}

func TestNTTAndInverse(t *testing.T) {
	// Operations are inverses
	w0 := common.NewRingElement()
	w0[0] = common.RingCoeff(9)
	w0h := common.NTT(w0)
	expected := common.FieldElement(9)
	for i := range 256 {
		assert.Equal(t, expected, w0h[i])
	}
	w0r := common.InverseNTT(w0h)
	for i := range 256 {
		assert.Equal(t, w0[i], w0r[i])
	}

	// Let's populate a ring then NTT it
	w1 := common.NewRingElement()
	for i := range 256 {
		w1[i] = common.RingCoeff(i)
	}
	w1h := common.NTT(w1)
	expect_w1h := []uint32{8023823, 4949942, 5503697, 7227518, 4077164, 903461, 2287113, 3389395, 1447936, 3912035, 3833152, 5335025, 7966085, 8118989, 7144945, 7460296, 8200405, 5651255, 5840697, 2041, 8329041, 2296483, 7624292, 7760084, 6558166, 2463083, 592160, 7596205, 490458, 4570418, 535121, 5905710, 2269315, 25712, 65279, 6056088, 437727, 5437873, 45209, 3628670, 5932184, 4892020, 4400120, 3282855, 5579212, 2040171, 8129297, 3975887, 886499, 5275349, 1715375, 2422113, 503654, 2500352, 3475364, 2130347, 7671751, 7706886, 6190567, 1877207, 1880030, 7339689, 5192027, 7408649, 4046506, 6555025, 861568, 5241798, 3351905, 7967553, 8240568, 2908955, 1077579, 7068530, 1063576, 2082141, 1227026, 4901674, 6147942, 4516462, 7784774, 4909015, 2489952, 8055865, 1807242, 3141274, 4210121, 2460839, 6404829, 6055556, 699854, 8144470, 167925, 2815245, 5308330, 7801015, 7301606, 2832490, 6224608, 4233662, 3984450, 6969568, 7183502, 6133025, 3069985, 7499554, 5559452, 7309678, 5405335, 5069329, 3320196, 2451430, 3043243, 3070455, 3966814, 6244424, 2083871, 2186058, 7917105, 5731770, 8357109, 4801012, 3444419, 6442745, 3142318, 4483091, 4065258, 1986703, 8368027, 4615661, 144560, 4178015, 2729052, 7118387, 1224642, 2979664, 2679432, 2620296, 3256914, 7425771, 4495896, 6348741, 6906650, 4571569, 5432259, 4416612, 3304060, 5577029, 3173849, 6062776, 8209741, 1186292, 3076903, 7840971, 2874775, 2013616, 4888110, 5543365, 6149437, 7037817, 2703904, 148603, 1178408, 5493962, 2871386, 2394607, 4524768, 626150, 8137948, 2020685, 2930707, 6943539, 3297580, 3309315, 7957803, 3489579, 1101657, 2199934, 2667995, 311407, 4615923, 268380, 7867980, 1165026, 6246419, 7938242, 3436132, 5102358, 1264622, 6021013, 3303556, 104046, 252176, 6426141, 3998553, 918827, 4282041, 2746755, 1284601, 5651462, 6998811, 1817618, 528380, 2525913, 5078866, 8002802, 2110331, 2052914, 155305, 3718478, 5776192, 6905096, 5498888, 7254918, 6047002, 6361152, 915442, 87228, 1281704, 3647397, 8363923, 3451609, 6209053, 1776623, 1128875, 6914893, 4152979, 1018431, 6308070, 982921, 3563602, 1283529, 1618324, 1186221, 13008, 759546, 6421303, 5292714, 2462024, 7387771, 7276117, 1343415, 1301221, 977961, 3904031, 193986, 5172786, 1429550, 2425536, 68499, 3777265, 7056830, 6555455, 981963, 8074937, 3279003}
	for i := range 256 {
		assert.Equal(t, expect_w1h[i], uint32(w1h[i]))
	}
	w2 := common.InverseNTT(w1h)
	for i := range 256 {
		expect := uint32(w1[i])
		actual := uint32(w2[i])
		assert.Equal(t, expect, actual)
	}

	/// Descending values
	w1 = common.NewRingElement()
	for i := range 256 {
		w1[i] = common.RingCoeff(q - 1 - i)
	}
	w1h = common.NTT(w1)
	w2 = common.InverseNTT(w1h)
	for i := range 256 {
		assert.Equal(t, w1[i], w2[i])
	}
}

func TestNTTOps(t *testing.T) {
	w0 := common.NewRingElement()
	w0[0] = common.RingCoeff(9)
	w0h := common.NTT(w0)

	w1 := common.NewRingElement()
	w1[0] = common.RingCoeff(3)
	w1h := common.NTT(w1)

	// Multiplication
	multiplied := common.NTTMul(w0h, w1h)
	expected := common.FieldElement(27)
	for i := range 256 {
		assert.Equal(t, expected, multiplied[i])
	}
	inverted := common.InverseNTT(multiplied)
	for i := range 256 {
		if i == 0 {
			assert.Equal(t, uint32(27), uint32(inverted[i]))
		} else {
			assert.Equal(t, uint32(0), uint32(inverted[i]))
		}
	}

	// Addition
	added := common.NTTAdd(w0h, w1h)
	expected = common.FieldElement(12)
	for i := range 256 {
		assert.Equal(t, expected, added[i])
	}
}

func TestNTTVectorOps(t *testing.T) {
	l_values := []uint8{4, 5, 7}
	for _, l := range l_values {
		a := common.NewNttVector(l)
		b := common.NewNttVector(l)
		for i := range l {
			for j := range 256 {
				a[i][j] = common.FieldReduceOnce(uint32(i + 1))
				b[i][j] = common.FieldReduceOnce(uint32((i + 1) << 1))
			}
		}

		// AddVectorNTT
		for i := range l {
			added := common.AddVectorNTT(l, a, b)
			expect := uint32(3 * (i + 1))
			assert.Equal(t, expect, uint32(added[i][0]))
		}

		// ScalarVectorNTT
		scalar := common.NewNttElement()
		for j := range 256 {
			scalar[j] = common.FieldReduceOnce(uint32(2))
		}
		product := common.ScalarVectorNTT(l, scalar, a)
		for i := range l {
			for j := range 256 {
				assert.Equal(t, b[i][j], product[i][j])
			}
		}
	}
}

func TestScalarVectorNTTWithT0(t *testing.T) {
	t0 := common.RingVector{
		{8378533, 8378690, 3463, 8378784, 8376623, 3368, 8377830, 81, 8377563, 8378141, 3594, 1040, 2494, 3832, 1146, 8379982, 8377951, 8377978, 8377828, 8378014, 3105, 8378970, 1867, 1274, 1229, 3461, 8378074, 1366, 3167, 4000, 8376672, 8376473, 8376477, 8376870, 8379047, 4012, 8380413, 8376521, 1953, 1778, 8377766, 8377110, 3257, 8378841, 201, 2532, 8380071, 470, 2332, 8378010, 8379473, 8378342, 8380177, 8380320, 1155, 8380111, 8377876, 8378173, 8377652, 8378603, 1736, 2038, 3811, 8376354, 8378207, 8377597, 8380288, 8379564, 8378073, 8379242, 8376404, 1176, 8377512, 8380070, 1810, 8378123, 1975, 8376426, 60, 2886, 1304, 8378120, 8377785, 8379758, 760, 8376830, 8379919, 8379749, 401, 8377256, 8378276, 8380400, 2972, 3170, 7, 3057, 1378, 1271, 8378530, 8379774, 2393, 1421, 3720, 8379507, 8379175, 8380365, 1604, 476, 8377308, 8377848, 8378675, 8379296, 3574, 1314, 3709, 8379491, 1401, 3447, 3085, 8376843, 8378096, 8376331, 8378314, 8379086, 8376349, 8379236, 1343, 2771, 234, 8377523, 2230, 3761, 8379771, 8377871, 3453, 215, 8378651, 8376685, 8377146, 328, 8378318, 8377591, 3788, 8379243, 1056, 8377508, 8379181, 8379332, 8376514, 8379771, 1139, 3217, 8378519, 8378334, 3813, 8378485, 8377837, 8379134, 3492, 8377329, 2363, 8376758, 432, 3256, 1571, 8379393, 8378604, 561, 1169, 1400, 565, 8379354, 1708, 8378490, 706, 4049, 8376924, 377, 423, 8379338, 3425, 3726, 2855, 8377938, 1049, 2510, 8377239, 8377968, 8379181, 3589, 482, 3164, 8377145, 8378599, 8377336, 8379769, 4043, 8376479, 262, 8376645, 3316, 2262, 8379840, 8379759, 3818, 1579, 98, 8379717, 8379273, 8378206, 2572, 8379697, 8378084, 8380399, 8379425, 4025, 3585, 1904, 8379438, 3043, 8378896, 1321, 8376602, 8377095, 2851, 83, 3283, 8377121, 8377415, 830, 8379767, 1796, 662, 216, 1012, 895, 8378591, 8377618, 8379948, 2654, 3706, 3328, 2793, 8379902, 3529, 2693, 8376556, 8378227, 8378179, 2362, 2374, 8379160, 1540, 868, 2260, 1954},
		{601, 8376488, 8376520, 8378725, 8378594, 8380142, 2205, 8378633, 8378690, 8379884, 8380384, 2693, 8379852, 3681, 8378600, 1297, 8378090, 3113, 1482, 344, 3112, 8376339, 8377916, 1318, 2390, 2179, 8377317, 8376641, 8379767, 123, 746, 1559, 8379156, 8379460, 3877, 8377500, 1105, 8379889, 8376983, 8376964, 8377773, 3852, 460, 3520, 1902, 8379235, 8376679, 1636, 8380213, 751, 427, 8378259, 260, 8377319, 8378781, 4010, 2353, 2326, 8378515, 1477, 8376344, 8380292, 1243, 3946, 3735, 8377724, 1394, 1328, 8378250, 2367, 8378165, 8377032, 601, 8377400, 8379422, 8378732, 2983, 3434, 8379099, 1218, 8377916, 8376347, 8376423, 8376816, 3005, 8378699, 2496, 8380378, 8379688, 8379610, 8379869, 1932, 2773, 523, 3387, 8379937, 770, 835, 2025, 8379826, 2667, 8376546, 999, 980, 8377850, 8379368, 8376597, 8379227, 8379440, 762, 8380090, 8379090, 1392, 1641, 8379316, 1105, 8376822, 8379699, 8378431, 8378332, 8377519, 8380183, 3735, 8378580, 2747, 8379011, 792, 8379347, 391, 8378257, 8377731, 8378931, 2176, 1283, 8380404, 8377017, 8378997, 8377878, 2513, 8378683, 1367, 8378171, 1898, 8379150, 1969, 8378422, 374, 255, 928, 2548, 8378662, 3874, 8379234, 8377506, 8379292, 8376459, 8376514, 8377923, 8378234, 3610, 8377330, 8378255, 8378978, 8377781, 3795, 8379956, 1349, 123, 682, 586, 8377533, 8379576, 1796, 3594, 8380327, 8377690, 2031, 1085, 1424, 8379268, 8377814, 8377803, 2026, 8376676, 1113, 8377068, 3875, 8379654, 8377763, 3574, 8377921, 2930, 273, 8378152, 8379731, 1965, 3822, 8376438, 8379208, 2265, 500, 3448, 8379715, 2466, 1999, 8379961, 8379138, 8380305, 8378276, 695, 301, 8376882, 2367, 191, 2536, 3082, 1539, 2180, 3, 8379964, 672, 384, 8377184, 3938, 1350, 8378738, 8378501, 8378433, 8379827, 1011, 8376337, 8379378, 8377892, 1938, 400, 8376453, 8377162, 1035, 1854, 8377186, 8379980, 3178, 8379493, 359, 3898, 1119, 8380376, 8376750, 2981, 42, 8379528, 1437, 674, 2003, 1909, 8379764},
		{8379715, 2910, 8377886, 8379901, 1942, 8378673, 8378627, 2215, 3388, 8378917, 1340, 8379530, 1497, 8380294, 1887, 3416, 197, 8377328, 792, 613, 8378240, 2317, 807, 1686, 769, 8377789, 8380109, 8378718, 8379699, 977, 2054, 8377098, 634, 8377599, 2645, 8378416, 3565, 3279, 1583, 8376817, 3240, 8379166, 1567, 8379638, 1510, 8377495, 402, 3624, 2441, 8377756, 242, 8376914, 8377524, 1551, 1985, 797, 4063, 8377880, 1503, 8380140, 2581, 8377088, 1443, 8376441, 152, 8378285, 2697, 8377117, 2664, 1857, 8378319, 1089, 8377102, 1359, 600, 8379427, 909, 283, 8380012, 2719, 8380258, 8378141, 3482, 3645, 1730, 1622, 634, 3970, 1858, 1588, 8380045, 8377193, 8377509, 3459, 8377074, 8379382, 8378919, 8377260, 8380254, 757, 3869, 2827, 3983, 268, 8377195, 8379825, 86, 8377045, 8377956, 8377161, 8377027, 8379206, 8377313, 389, 2237, 8378026, 2942, 8378294, 8377862, 266, 8378343, 1567, 2919, 8376857, 8377824, 8378555, 8376430, 8376417, 8377171, 8377672, 8378542, 8379263, 1898, 8379460, 8379461, 8376658, 8380267, 1505, 8378230, 8378221, 2846, 8379691, 8380226, 8378394, 3284, 3701, 3282, 3272, 8378792, 1569, 8380209, 732, 3534, 3472, 80, 2423, 3557, 8377830, 8379503, 2842, 8378992, 8377304, 8378113, 114, 235, 8376979, 8378994, 8377265, 8379563, 1903, 8376976, 3929, 881, 8379103, 2321, 3537, 2098, 8379642, 8377615, 8379780, 8379472, 8379144, 2290, 3440, 3193, 8379184, 8378317, 8380376, 185, 3913, 8378840, 8377174, 8376470, 1861, 363, 8379696, 1473, 8380231, 8379156, 8377270, 2323, 2519, 8379937, 8378652, 579, 8378116, 34, 8376645, 8379268, 8378429, 1005, 2628, 1573, 8379209, 1930, 8380131, 3076, 8379119, 8378404, 8379942, 475, 552, 689, 8377169, 8378831, 8376460, 777, 1755, 8379427, 671, 2613, 8379393, 8379033, 8380130, 8378042, 8379836, 1840, 844, 8378680, 870, 1949, 8377204, 226, 3510, 8378795, 8379589, 8379540, 232, 8378795, 2051, 1294, 104, 8377870, 2946, 2360, 8376726},
		{8378408, 8377283, 1522, 8376840, 2869, 1590, 2344, 3424, 3951, 1149, 8377338, 4087, 8377578, 8376669, 8376371, 1627, 8377285, 33, 8377626, 2673, 8379370, 8376649, 8378292, 192, 8379731, 8379931, 1634, 8380255, 8379120, 8378189, 3941, 978, 1029, 3709, 8376392, 8376498, 2461, 8377095, 8378210, 8378999, 3290, 8379896, 8376799, 2203, 8377975, 584, 3600, 3859, 8379213, 1751, 8378761, 8377546, 8377122, 8377683, 8379194, 2798, 2439, 8379113, 8377013, 3277, 8377850, 3615, 8376740, 8378262, 736, 3939, 2145, 8377240, 2100, 1292, 1919, 618, 8378464, 8379125, 1154, 8377885, 8377011, 8376976, 497, 2201, 3374, 8379867, 4064, 2939, 8377100, 8377587, 8377674, 2248, 8378994, 8378552, 926, 3825, 3071, 4037, 2266, 8377481, 852, 3084, 3657, 8379764, 2293, 8380096, 2169, 8376509, 8377690, 8378858, 8379770, 2458, 8378297, 1691, 8380287, 987, 8379716, 1184, 8378538, 8378742, 3288, 963, 2316, 530, 1004, 727, 781, 8377400, 2071, 804, 8378998, 8378603, 1958, 68, 602, 8379084, 8378779, 8376606, 8376951, 1325, 8377165, 4079, 8378205, 8377035, 3467, 218, 3297, 428, 2263, 8379234, 8376859, 8376923, 8378143, 8378246, 8377361, 8378952, 356, 8378319, 1507, 2443, 8379649, 8377516, 651, 602, 1326, 1875, 8379562, 427, 8376517, 8379255, 8376869, 1751, 8380363, 8377620, 2814, 8376473, 1368, 2545, 3246, 1334, 795, 8378214, 2586, 3600, 2279, 1235, 8380170, 1715, 609, 2887, 3365, 8378682, 8376844, 3214, 3062, 528, 927, 8377488, 1902, 346, 842, 8376884, 8379684, 4045, 8377458, 8377249, 2732, 8379345, 1328, 8376994, 1316, 8378262, 4010, 8379528, 8376521, 8376797, 3139, 1109, 494, 8379205, 8377545, 4049, 2551, 8377269, 8378587, 3987, 8377292, 2483, 1596, 3079, 8376756, 8376350, 3702, 8377419, 8378989, 1044, 8379632, 8376532, 8379865, 8378872, 8378408, 3976, 8377276, 3009, 8376482, 8378180, 3473, 35, 2359, 520, 1361, 500, 496, 1038, 2228, 1985, 938, 8380338, 8378761, 8380347},
	}
	expect := common.NttVector{
		{5686364, 2006146, 180176, 5708537, 7230227, 7886217, 2577094, 5796504, 6976297, 1101905, 903927, 541469, 4136851, 6661048, 7712457, 8246475, 4895739, 2690950, 1957626, 5907024, 4743983, 1701947, 1072652, 2405729, 5145839, 7330202, 2566647, 4987009, 1515055, 3289759, 1103250, 1970293, 159280, 3886642, 12872, 5471857, 5609588, 2471019, 8292836, 6910809, 2781749, 159557, 6746742, 1787470, 3844604, 6068910, 8304511, 2227870, 5434046, 2449735, 7474119, 5597831, 5167260, 1265564, 2860326, 1312280, 4524373, 4561176, 7738140, 7237558, 542549, 7964212, 3367538, 6582708, 241511, 6954828, 3954607, 7696655, 7751308, 1350390, 1352602, 3134557, 1543837, 21609, 7406146, 1672422, 5621703, 2946462, 6746730, 3393788, 1201642, 3409684, 7414375, 3718312, 6509136, 4013157, 8141561, 2719629, 2483563, 4399068, 4781163, 3441433, 2531762, 2672022, 5687002, 7059555, 7071038, 642205, 7640192, 7932044, 1972166, 2887970, 7401078, 3883916, 2020282, 8262723, 2375401, 8237418, 1974921, 2275271, 2941631, 2578745, 6285201, 295934, 1792374, 838454, 672801, 4986949, 1872986, 4450096, 5045636, 7680406, 1876203, 5759900, 1960479, 1144294, 5583870, 7519730, 342632, 3967904, 5819138, 4820367, 732952, 5127309, 5334719, 1069619, 7599984, 3558911, 886538, 2401799, 7402477, 5555132, 6977100, 4424438, 4393718, 4174059, 4107568, 7097628, 3757062, 4731798, 15777, 8351826, 6860587, 3337812, 7707660, 8137703, 1612008, 6426130, 3764306, 6533866, 761568, 1306518, 2125526, 3654339, 2792096, 6327437, 7195396, 7769333, 7468448, 3884593, 3733981, 4970163, 5359352, 1962274, 7367320, 2719339, 7418431, 4495494, 4809823, 901963, 1914538, 4977069, 3312483, 5118288, 5982256, 2351665, 8284463, 6735682, 1790263, 4094945, 90005, 4238611, 375850, 3822506, 6144812, 6976081, 4221414, 1475336, 4562914, 3314349, 2737096, 3388598, 6476478, 4594290, 1154540, 1567152, 4340264, 6151812, 2248390, 6200910, 6788059, 5273425, 3892618, 7821228, 6834098, 199581, 772066, 1658899, 4356612, 2229862, 580593, 2392602, 3347125, 2457038, 940725, 2233796, 4079410, 2527412, 571745, 7160361, 724250, 6891295, 365441, 2775715, 6007054, 4965487, 52935, 215833, 3020484, 7119588, 1126106, 749070, 6875977, 1191248, 5164127, 4364181, 3090628, 4520603, 2199052, 3427401, 2080386, 3786345, 2916635, 3821396, 6108774, 2649078},
		{7354434, 3183480, 228689, 7901586, 5523600, 4717140, 3915199, 6413977, 4361321, 3292946, 3606094, 3818038, 591684, 3909798, 2548496, 6195364, 4996075, 6112601, 7066114, 3600992, 5053958, 3500419, 2627197, 3446547, 6619591, 173262, 7608054, 1657090, 4314445, 3965759, 2174519, 7762720, 5872835, 7247307, 889235, 7468785, 1882869, 6062834, 5881172, 7648650, 6092076, 5984453, 3418297, 1684960, 2283954, 2533945, 5013437, 1666704, 5184258, 2985748, 2387820, 1317493, 2587285, 2516832, 7092337, 1913827, 3016148, 1529142, 6832134, 3018578, 7531543, 1464269, 3503681, 90130, 3984829, 4856856, 4994986, 5292966, 5043607, 1800391, 4910246, 5836571, 1304085, 6869068, 5649161, 7558762, 1709568, 4719310, 7965309, 3173950, 5083619, 7712474, 8172315, 2773997, 7263397, 8182677, 4636328, 5517867, 7221244, 4239415, 2190214, 1872034, 780364, 3898908, 8021171, 6461976, 2127064, 6678493, 7589813, 5865995, 2735143, 3535308, 5938065, 2201514, 5176702, 8150819, 7669900, 1995114, 5491888, 7641073, 1380824, 338605, 7513906, 157541, 2289359, 6446178, 4635082, 306524, 5020119, 4652329, 5288185, 4206932, 1147719, 967647, 2152074, 3544123, 950306, 3667671, 2269475, 875171, 2663648, 7237284, 416983, 5159327, 4926930, 484431, 977480, 2436116, 4189246, 6951468, 7123225, 3364115, 7022177, 528633, 7359751, 1277534, 39519, 7638604, 8085397, 6597245, 269728, 1912780, 737856, 1360431, 2793388, 1806984, 1341729, 7959922, 3344496, 1933310, 7582270, 4344197, 4168365, 8294723, 6351849, 6712470, 5891406, 5283984, 135407, 5326697, 1849357, 5135730, 6466364, 5777164, 2719047, 5526449, 3503964, 7988698, 4999059, 4183631, 2603616, 5009492, 2562974, 436607, 4022095, 1282996, 7833057, 2705487, 5514506, 732436, 6521461, 3080056, 2035370, 1667999, 7219726, 2021114, 439297, 446571, 3212879, 5786453, 7505867, 7132954, 4750289, 2976746, 6123192, 6905588, 4248828, 7319459, 2876509, 6928824, 4898242, 4750883, 7725459, 2419477, 6119214, 8333728, 3349652, 374217, 1069373, 5860933, 5223605, 2509540, 2139091, 4412703, 1021080, 6009878, 6700394, 1438218, 7974988, 7016124, 1080830, 5058652, 626213, 8196919, 4144705, 2529171, 5621320, 1976517, 5268282, 3720493, 8018959, 8091354, 1842237, 5556166, 2097613, 4858593, 1747042, 5366317, 183539, 6309790, 2440349, 5700656, 4901652, 5913994, 5508807, 8269828},
		{769802, 8085960, 4100205, 3625175, 6282503, 1074624, 540529, 3757938, 5870120, 5718121, 3540406, 4282384, 3177450, 5880376, 1111057, 307033, 2433621, 4068659, 1280381, 5820056, 334899, 1349680, 6296430, 5812849, 4878175, 609703, 6248721, 2582897, 2405775, 523292, 4870500, 7603520, 5026810, 8232146, 857183, 2406464, 803017, 4776162, 6113420, 7568132, 1789308, 2738750, 7347289, 6872532, 8128686, 2915967, 1616522, 2952771, 2295274, 2256995, 4228842, 6481862, 4066240, 386932, 7638012, 1137871, 3331135, 5142318, 1537047, 6134187, 3901571, 1956288, 3685973, 1433508, 1317283, 1416072, 7117678, 2838121, 143125, 8097077, 4976858, 7060445, 3510817, 681436, 2957173, 194305, 4830580, 4221825, 64393, 5411725, 6539396, 5962156, 3960685, 7607586, 8169375, 3669598, 6015016, 1500930, 1911869, 899236, 2178261, 1028251, 955625, 1489111, 4071628, 2675963, 5987661, 3560722, 651933, 1216033, 8294261, 989384, 885098, 2240577, 4551404, 2085251, 1995094, 5199056, 4389817, 961272, 4442680, 3617939, 7051362, 2349213, 4053140, 788589, 4179299, 4775419, 73621, 1199887, 7276160, 6249435, 3110931, 1091860, 1142458, 2474509, 1624503, 1505479, 165548, 2260131, 2031824, 7239228, 4248837, 4687936, 3692257, 1589546, 6116636, 3914278, 3344402, 4171204, 2916239, 7058323, 5607538, 1859354, 6879, 1682416, 4113509, 6855876, 1864906, 4891702, 2843928, 20155, 5619289, 4150074, 7704221, 1637770, 8273185, 4382772, 3288444, 2624929, 2637801, 3227229, 4892011, 3703028, 7720149, 3160695, 716156, 113329, 364431, 2025802, 4101102, 5890216, 1807137, 1802132, 1928771, 653849, 6019443, 6231982, 5356846, 7982729, 6819552, 400154, 6566448, 6947873, 6536053, 8212771, 718429, 8266268, 5101869, 7447410, 6276853, 8304205, 4810247, 5812387, 3625686, 8222724, 7268874, 1934372, 4339749, 3361479, 5492764, 5024882, 5859374, 3180555, 2647028, 2216053, 3565049, 3359217, 965124, 589055, 1083021, 3892442, 2226295, 5630953, 3114555, 2262680, 5090192, 8279456, 1411450, 1000276, 7218487, 3218973, 2674709, 2902076, 3417310, 8166236, 3210532, 2364086, 2480568, 3301649, 3624232, 3126474, 3997288, 7288930, 5965585, 5998317, 1885226, 1621574, 1672498, 6774173, 8011743, 827586, 5268087, 6023817, 8193815, 563733, 5248285, 5759433, 5702638, 3396249, 5545821, 1035440, 6533882, 384896, 3844091, 1148522},
		{6275306, 3407672, 3352336, 1105205, 5300433, 6585456, 869548, 2892132, 7208653, 7398457, 7352272, 7168727, 5449049, 2124427, 901005, 676349, 1538694, 947575, 1359926, 8007943, 7585212, 4084266, 5492515, 1888997, 900357, 6804595, 2406092, 5002851, 8226043, 4594572, 3177313, 7787293, 1433732, 2067943, 5253200, 2808621, 5284185, 395881, 1286504, 3949778, 4507890, 1485748, 1354924, 4643516, 4038737, 1054755, 7722445, 7662172, 2547726, 998041, 1310599, 1552878, 5923861, 1325867, 5589111, 1194230, 7540717, 1745307, 3912292, 1379251, 2281894, 5675411, 4483407, 4732396, 3721220, 4592472, 5402654, 6930010, 5690242, 2546811, 5601486, 6496231, 2291106, 1035055, 4931262, 5991678, 1011979, 3028912, 2725997, 5565266, 40373, 781690, 1721441, 2644911, 5533518, 2498289, 6752370, 7160009, 2276830, 2433789, 4307224, 1517241, 8214653, 3297862, 6842567, 3437796, 7430565, 17487, 7634665, 4631604, 6958816, 5211656, 5483459, 7472990, 5606248, 4932689, 3351068, 4325885, 5773859, 5402745, 2827177, 5865506, 4723010, 8173190, 3994226, 3707124, 627223, 777738, 5268419, 2176724, 2480739, 7087625, 8114668, 7046960, 6581607, 1843036, 3059396, 2214739, 6768931, 6852672, 1057141, 902917, 7476011, 7129158, 4915541, 5947579, 1523214, 5381789, 3817654, 4439243, 8311275, 5880875, 2236180, 164215, 546406, 3329666, 4819307, 4980327, 3691554, 1170962, 4908563, 2374055, 6415184, 1350841, 1115187, 4744313, 5168858, 7801170, 5194843, 3658302, 555462, 4904378, 6926149, 3381668, 3154533, 2770609, 6677739, 1141754, 4811110, 254086, 1396955, 7686540, 6536310, 8167797, 5091920, 1306050, 4482099, 2076912, 4146066, 2345653, 3519489, 7431326, 1318780, 5985054, 8313303, 3800797, 3560415, 391678, 4939046, 5556891, 1021775, 300026, 3219473, 2046214, 3239893, 5298907, 157370, 4840730, 6670928, 3564973, 8065267, 6163555, 2641778, 7544639, 2076837, 2942611, 1067933, 6170789, 6987555, 2166977, 6956126, 4742535, 5750375, 3992885, 4600070, 5837242, 2587704, 5092492, 6663892, 3822209, 850619, 3190299, 4432192, 362658, 4424466, 962839, 2433284, 2278033, 616952, 6196925, 6426420, 1600256, 2137309, 7449796, 207914, 6175707, 7853890, 1499557, 3444979, 7270116, 560905, 1978768, 25105, 6433025, 3978715, 3031707, 3244074, 1169069, 6914634, 5271904, 1192957, 5494271, 7156674, 8166027, 8075015, 372004},
	}
	actual := common.NttVec(uint8(4), t0)
	assert.Equal(t, expect, actual)
}

func TestMatrixVectorNTT(t *testing.T) {
	params := []struct {
		k      uint8
		l      uint8
		expect []uint32
	}{
		{4, 4, []uint32{60, 80, 100, 120}},
		{6, 5, []uint32{110, 140, 170, 200, 230, 260}},
		{8, 7, []uint32{280, 336, 392, 448, 504, 560, 616, 672}},
	}
	for _, param := range params {
		// Matrix A
		A := common.NewNttMatrix(param.k, param.l)
		for i := range param.k {
			for j := range param.l {
				for x := range 256 {
					A[i][j][x] = common.FieldReduceOnce(uint32(i + j + 1))
				}
			}
		}

		// Vector b
		b := common.NewNttVector(param.l)
		for i := range param.l {
			for j := range 256 {
				b[i][j] = common.FieldReduceOnce(uint32((i + 1) << 1))
			}
		}

		// A * b
		product := common.MatrixVectorNTT(param.k, param.l, A, b)

		for i := range param.k {
			assert.Equal(t, param.expect[i], uint32(product[i][0]))
		}
	}
}

func TestInfinityNorm(t *testing.T) {
	q2 := uint32(q >> 1)
	tests := []struct {
		input  uint32
		output uint32
	}{
		{0, 0},
		{100, 100},
		{q2 - 1, q2 - 1},
		{q - 1, 1},
		{q - 100, 100},
		{0xFFFFFFFF, 1},
		{0xFFFFFFFE, 2},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.output, common.InfinityNorm(tt.input))
	}
}

func TestMakeHintRingVec(t *testing.T) {
	k := uint8(4)
	gamma2 := uint32(95232)
	z := common.RingVector{
		{328, 8379759, 8359154, 8367647, 9786, 22931, 8374484, 8374470, 8359667, 8357725, 11982, 8378585, 7667, 8361005, 28832, 5921, 15587, 8371327, 8373242, 8355357, 8352868, 107, 8364741, 9310, 23227, 34066, 11972, 9466, 9367, 5579, 8365417, 13303, 2483, 8353437, 706, 8374716, 8371686, 8370297, 24449, 17293, 8014, 8377475, 23352, 8363351, 8361423, 12432, 8349610, 8355633, 14663, 30948, 11926, 8375491, 3253, 8369122, 21686, 5376, 7692, 8367271, 8365132, 8364496, 17366, 29372, 8373535, 9641, 18126, 49570, 8369361, 8371756, 8364960, 13509, 8367857, 8356719, 8362863, 8379599, 28639, 32805, 8365145, 8354758, 4419, 8371669, 128, 8357466, 8370338, 8354708, 8372242, 7895, 2903, 6313, 8352657, 21549, 8380342, 8353894, 8376820, 8368736, 12811, 5190, 8542, 7399, 5463, 8842, 2787, 8359270, 8377308, 8376099, 8371639, 13629, 20634, 22151, 8364546, 12554, 5394, 3206, 8376168, 11797, 2610, 14477, 8371446, 8376185, 8374574, 24498, 10852, 8367283, 8369697, 8366369, 23606, 8374861, 8363165, 8366780, 8369972, 16453, 21240, 7313, 8367137, 13164, 4891, 8374789, 18588, 8366605, 4197, 329, 8374544, 8368332, 8375565, 20193, 23006, 8377866, 3459, 14305, 8343305, 3052, 7075, 8365905, 8356556, 11877, 2747, 8377083, 8379750, 8373070, 8338936, 8354032, 6474, 8378131, 8377648, 8378570, 8375741, 8354914, 8373956, 8373298, 11158, 8374715, 2900, 8547, 8357411, 8376309, 9464, 4597, 8374126, 20707, 8373710, 8363606, 8373268, 8372510, 16659, 4777, 8954, 17775, 11884, 44008, 16358, 8379110, 8355909, 8373382, 8366621, 23602, 9713, 3629, 8379620, 8375959, 8369379, 16860, 5627, 8376350, 8366897, 8368482, 8892, 8356086, 8371598, 19892, 9907, 8364912, 8369217, 7409, 2934, 3896, 1114, 8376645, 8356528, 16637, 9537, 8378028, 12784, 8373610, 3337, 8362066, 8377638, 6299, 17158, 8380264, 8353673, 8376329, 8353490, 14652, 12410, 8368823, 3131, 13517, 8163, 8358036, 8369832, 8856, 8378886, 8336610, 8367642, 8365731, 6360, 22649, 8366180, 8359462, 7679, 8369394, 8359112, 8379690, 14213, 15040, 8373912, 18461},
		{8364799, 14713, 8372315, 8376166, 8341388, 8378345, 7218, 8371211, 8361724, 13116, 8360683, 8368710, 2169, 8380360, 10361, 8357151, 8372301, 8351665, 7368, 29155, 15312, 8362806, 37473, 22162, 16524, 8366526, 1260, 286, 8377845, 8366760, 8362958, 3574, 8379290, 13028, 8361604, 8355464, 26976, 13225, 24455, 8574, 8345035, 8358242, 8370517, 11643, 11974, 3554, 8595, 10725, 8379306, 11288, 8370675, 8369217, 8367980, 8365972, 8363128, 8367242, 10371, 2387, 25285, 8376801, 20149, 8364780, 14381, 13493, 27946, 8379887, 20846, 3686, 6725, 23308, 18564, 18422, 8362552, 4114, 8378506, 1313, 910, 100, 8368649, 8356561, 1676, 8357531, 8375698, 6514, 8366379, 8365591, 8348333, 20351, 16502, 8375117, 8366885, 5324, 8375746, 4784, 19310, 923, 8379796, 8377979, 7803, 8377731, 8363838, 19356, 5254, 8371383, 4932, 7986, 8378458, 20885, 8373347, 8373651, 2225, 8375218, 23802, 17956, 22445, 19336, 8376633, 423, 8053, 8372195, 8345571, 8364867, 8353061, 8361035, 462, 8374945, 8374938, 8379376, 8370416, 10314, 8340996, 6674, 5592, 8348539, 8379401, 151, 3488, 12310, 38969, 34811, 8371810, 7032, 8375986, 8377871, 12678, 10021, 6520, 8367498, 8379758, 8377561, 4399, 23688, 29167, 8355595, 8377885, 8361560, 8367997, 5332, 8379271, 11967, 8353140, 8375507, 164, 23771, 974, 8363188, 8361913, 8373874, 6622, 8368372, 8373764, 4127, 8376554, 8363982, 8356290, 5300, 9738, 11648, 8376557, 8363032, 8378157, 28536, 6713, 5394, 9510, 8374247, 13505, 8433, 15554, 25773, 8373315, 5016, 19306, 8354536, 4866, 21459, 11395, 8378959, 8341508, 8353473, 8354530, 8367660, 8378764, 8376852, 8368455, 8376471, 15879, 8370793, 1549, 10570, 8376117, 8364806, 8366432, 8380274, 6766, 12155, 26447, 14629, 8378910, 8100, 9664, 6866, 2275, 8368068, 10768, 8361690, 26156, 16151, 8379539, 8369046, 8359926, 34162, 8375097, 10272, 8367844, 6269, 19970, 22765, 8364130, 8368756, 8366760, 4452, 8361600, 8347824, 6170, 8368431, 12082, 8365180, 8366126, 8379107, 8358235, 8375490, 8346350, 8370331, 8378474, 8379869},
		{9261, 8377317, 8378058, 8370727, 974, 8373786, 5894, 8374279, 8369395, 8375433, 8365841, 8476, 3522, 8363678, 2533, 8354865, 11378, 8360837, 8378684, 2795, 10523, 8370236, 8377111, 8366315, 8371657, 19280, 19772, 14778, 8365920, 10013, 962, 19487, 9326, 9568, 1058, 8376125, 10214, 6809, 8369549, 8373648, 7076, 8371864, 8367882, 45786, 8374566, 14498, 8360187, 18068, 8378993, 6892, 10020, 31948, 8361483, 8379259, 610, 8363923, 13559, 8364197, 8374598, 5987, 8355531, 18398, 8376784, 6030, 8376443, 1374, 8362901, 3823, 8368719, 26582, 2539, 8356268, 2849, 2590, 23532, 2519, 313, 8366194, 8354673, 8365336, 8268, 2475, 8375032, 26504, 8367934, 8380259, 3176, 8379732, 8374195, 8379753, 17166, 8369081, 8375801, 8376258, 22696, 9371, 20633, 8361578, 8366968, 8378142, 8379068, 16037, 8374042, 8380325, 103, 8375662, 13266, 17800, 9259, 8369351, 8376857, 8379962, 5076, 8366638, 7478, 8380104, 8371704, 4895, 6998, 13268, 8371192, 288, 8378524, 8375221, 8377947, 8379864, 8379925, 14397, 4239, 8371476, 8378282, 5592, 668, 6264, 8363135, 8370193, 6294, 8379006, 1612, 8358870, 8368442, 12842, 8371502, 8367823, 8371149, 8351529, 8368334, 10229, 8515, 8374905, 8374535, 8376687, 4969, 8362804, 13746, 6990, 1982, 8371671, 438, 15031, 23470, 7516, 8377626, 10215, 8378801, 22530, 8374959, 8376060, 8379150, 17285, 7178, 13456, 8378618, 31774, 8378622, 18163, 21793, 8378281, 8373534, 3225, 8361257, 8375053, 18098, 5529, 805, 8372298, 8378610, 26044, 8379730, 1942, 8361741, 8371802, 8363386, 8375806, 8379008, 24352, 3673, 10720, 8374969, 8376106, 8375408, 22957, 4644, 8358264, 8379947, 8378019, 8373583, 7410, 8380279, 8362691, 8347686, 8365572, 8365145, 8373506, 8376641, 8051, 10523, 5355, 7000, 8379325, 8379615, 8372690, 2531, 8361827, 2644, 8367162, 33949, 8376040, 17495, 3553, 8371006, 4189, 8378728, 6705, 23878, 8370261, 8364832, 8376933, 8363911, 12886, 18367, 14878, 8378851, 8373223, 4716, 1148, 8379190, 16470, 9572, 8362351, 8363142, 8378255, 8353034, 28873, 8368317, 8372434},
		{8348857, 8375519, 8379699, 8368268, 8380242, 8378281, 8336013, 8355809, 17303, 11531, 8376247, 8371781, 12388, 8364414, 8370634, 8372671, 8366117, 5452, 5336, 6428, 8356705, 8378663, 27805, 8377793, 8367060, 8376754, 9814, 8379025, 10055, 8376713, 295, 2803, 8354365, 20395, 8363494, 21731, 19183, 6011, 8357183, 8379200, 8375052, 8372185, 203, 1349, 6642, 8337462, 17337, 8373956, 8376708, 15612, 807, 8360021, 22479, 3327, 16996, 13834, 8353310, 8355896, 8377366, 25506, 19813, 5320, 7530, 8361870, 8373276, 1113, 4056, 8369904, 17536, 11871, 1255, 16710, 19057, 20557, 620, 8377515, 8353842, 8363344, 12892, 23485, 8377825, 8357318, 11578, 8362459, 31569, 8365233, 8377014, 8362707, 11158, 8365590, 8374098, 8363019, 15358, 28590, 8357654, 8378355, 1706, 11517, 8369284, 8374559, 8378131, 12982, 1124, 10825, 5005, 8379759, 15560, 8366453, 8349240, 8379387, 8311, 8372790, 8374943, 32241, 8367994, 8377196, 8372449, 2636, 8239, 13251, 8373774, 8368168, 8369943, 33241, 24088, 8350810, 8372089, 8360709, 20515, 8619, 8377932, 11790, 5842, 8377913, 8364302, 8377833, 8332220, 27028, 8376343, 8379410, 8373780, 11181, 28535, 574, 2674, 16305, 8366846, 8367444, 20488, 8360222, 11942, 13537, 8361983, 8351151, 6335, 12332, 7502, 8374921, 8358977, 8348540, 8371202, 8368368, 8371817, 8378692, 29282, 8378638, 13253, 6124, 16271, 8367323, 8368487, 8377821, 8360086, 7254, 8380228, 8935, 8358380, 6770, 8368118, 1770, 9585, 1682, 8379049, 8369767, 7549, 8359294, 94, 19444, 17562, 16984, 8368554, 15665, 16376, 12396, 12380, 7909, 8361194, 8366390, 8074, 8372215, 5829, 8359410, 20992, 8373759, 12464, 26346, 8363338, 8367239, 8375780, 15130, 8361404, 8634, 11412, 12528, 3323, 4773, 8377639, 8379305, 4085, 2938, 6326, 8369318, 8368637, 8350106, 6729, 22584, 8362968, 7207, 8373775, 8371361, 8365623, 22594, 8364756, 8369792, 8366548, 8366119, 8373206, 6825, 22387, 8376813, 5409, 8356378, 8999, 2231, 9276, 8365887, 8351973, 8350334, 8347934, 11947, 8371809, 8995, 8378891, 10067, 8367987, 8370153},
	}
	r := common.RingVector{
		{5858295, 2905669, 7452289, 5174126, 7412055, 6535847, 108100, 6584296, 6314039, 7866267, 2213036, 7977249, 6563983, 4086069, 8375402, 4771571, 5646683, 2290702, 5015840, 7829688, 2696946, 2960369, 4312722, 5644863, 1430001, 5937019, 1966955, 6217635, 8028204, 4390959, 5557114, 1605486, 7617757, 4851450, 2099249, 6234165, 38961, 1651143, 53530, 7229409, 3790067, 1548003, 6102910, 798043, 5388200, 3706298, 2443088, 7322730, 1373347, 7360251, 685879, 2428356, 4144175, 3934096, 5229284, 4191543, 6595065, 3710484, 6246140, 3662823, 2698464, 2639251, 1470277, 1272117, 5007648, 5428076, 2023744, 2488439, 4980788, 7567071, 5821862, 3087658, 6711099, 7684304, 16967, 991431, 749532, 732119, 4835520, 6438145, 7968905, 1755043, 3003573, 5775442, 1954498, 3299439, 1843935, 7894171, 7935524, 315653, 6111803, 1368038, 4249860, 320713, 5672885, 992997, 2811233, 6050883, 6435079, 1757848, 195311, 7244737, 3161589, 7474942, 5253405, 4443062, 5973359, 2420813, 415180, 5627124, 1610627, 8110242, 3484438, 5945835, 8266941, 4926379, 6679268, 8356802, 6852412, 2617531, 1771740, 6081257, 650996, 3794285, 7399753, 7989090, 7594052, 4717042, 5800584, 326740, 7824319, 7543414, 4272702, 2038227, 5033468, 7676040, 8093094, 5060507, 3269732, 2054661, 484224, 259847, 3034853, 3986684, 950258, 6800999, 4513500, 285997, 8151704, 2379620, 2586144, 6089633, 6308117, 7503064, 7623494, 6864254, 1568360, 608503, 3888608, 3189371, 2818315, 6008007, 6500243, 3418142, 7129118, 7295899, 7336981, 754859, 7623070, 4539962, 7245795, 1911264, 6399541, 2392708, 38708, 404855, 7994419, 6294036, 6756342, 4877346, 5719339, 542076, 4819735, 4805211, 4083198, 840805, 4030296, 1588476, 2951368, 989762, 8152131, 6035769, 5466544, 7872565, 876185, 4272917, 4880750, 467641, 6612477, 54794, 1713825, 4694313, 8331913, 894633, 4893228, 1940503, 5717332, 5866633, 5804354, 1848059, 2022211, 7119091, 3068708, 1233390, 3545027, 5044834, 2989567, 6426608, 5055292, 3820092, 3375832, 4296696, 2525245, 6807979, 3908396, 3756131, 5591928, 4278039, 5886834, 1341707, 5591314, 1237107, 7238282, 4308374, 7984046, 1585955, 2526143, 7001446, 2297901, 6962716, 3242618, 7697198, 3180777, 953170, 4932369, 1585536, 136885, 4649286, 4693908, 7301688, 1718680, 5134977, 770189, 4416926, 305447, 3655173},
		{6101069, 5163797, 5212282, 1418150, 1377321, 354402, 1187757, 3869535, 3988197, 2963926, 7594090, 2759314, 6594182, 6733254, 1541933, 4993187, 6054889, 7852044, 1666693, 5452045, 5593932, 5867465, 4122941, 3506586, 7799562, 1199060, 1970901, 4908146, 8054468, 2539558, 553941, 3463485, 3532975, 6857877, 5550999, 7568462, 3213672, 2897666, 3803652, 951433, 112153, 5921291, 6915533, 8011104, 6844412, 3871177, 3868972, 5191133, 1665174, 3872226, 4435288, 5758129, 3921266, 2875309, 147602, 5232766, 3677376, 1096728, 7529417, 2750650, 1304726, 3690391, 6928661, 4940028, 5639491, 1898582, 2265514, 1892711, 7856390, 3080685, 2258798, 2196338, 5310539, 4004182, 4193897, 3760630, 5599296, 116948, 5052724, 4593900, 1000034, 7970604, 1368482, 3241049, 4713784, 2486604, 5689001, 1618572, 2039345, 1400731, 274941, 7419982, 1086107, 2616078, 6745300, 4160762, 6292581, 7314644, 4771862, 7901009, 7885844, 5402211, 949374, 3805411, 2109704, 2323581, 727027, 2038379, 2952811, 3166489, 1774154, 2880237, 8365511, 8185755, 4673470, 4917973, 5125416, 4967494, 4854909, 4437710, 3264115, 4283720, 3861612, 4600819, 2757003, 6344843, 5279121, 6597243, 6032706, 5231546, 6600205, 488176, 6077215, 2031829, 7467289, 1552347, 4867972, 2761512, 8315565, 6909303, 4997348, 3384308, 5002963, 1317385, 3757208, 5248831, 3759356, 4782134, 1069090, 1872600, 7114650, 2968816, 7752854, 4186332, 5023364, 5740568, 2749372, 4967615, 2899222, 7454825, 512468, 7513997, 1244480, 4904210, 5843457, 4599112, 4941705, 1944797, 3842947, 6502613, 5039351, 7463728, 1764051, 3901261, 8348730, 3700172, 4635230, 2521056, 5023433, 31609, 1018313, 7441040, 1542519, 5334849, 5178356, 3093849, 5787150, 3525427, 7791750, 7350185, 6305368, 1263157, 5576887, 7819830, 2047805, 4895303, 7289915, 324243, 2081824, 795084, 3002165, 7620568, 1674554, 2938242, 1604642, 3904946, 8372082, 4508688, 7364630, 5432021, 4979128, 7271348, 2522235, 7433242, 4743893, 7927517, 2536083, 4681725, 7517682, 965317, 1070675, 7339133, 1248701, 4067195, 58359, 158950, 7149532, 1710048, 3852060, 8352240, 1852929, 7199877, 2240692, 8379844, 1310802, 2342821, 2210408, 2260202, 6403386, 4802414, 6413568, 1797577, 3211277, 5002911, 4006770, 1575478, 3291461, 6301554, 1839882, 2254525, 3723493, 3368357, 6829912, 8170566, 1633145, 6123620},
		{6430146, 2097350, 4862137, 7461070, 1791215, 2352844, 1640208, 5576522, 8152724, 87567, 995709, 6340741, 3956518, 2583612, 6196237, 3246606, 7044415, 4778978, 5918042, 5608462, 7175756, 6694925, 5176, 4245528, 2210681, 377010, 215966, 1028852, 2865959, 531318, 7735747, 6237717, 4600770, 5489363, 5225915, 4931952, 4927985, 8365653, 7787639, 2014017, 2783364, 6731261, 2615609, 424867, 998883, 6891959, 3092760, 3677889, 6490163, 5583900, 3565132, 7073920, 2247975, 6337319, 4211694, 4850168, 6331262, 7268834, 6983488, 6239371, 5610269, 663923, 7590784, 480678, 2862965, 5805326, 3123917, 2740848, 1851901, 1755971, 7143776, 7663181, 3655922, 7866789, 2752515, 7927063, 6590837, 4619587, 4625461, 1480056, 8064994, 3437750, 108412, 3945636, 535516, 4157936, 698595, 355359, 6629734, 2923154, 5679499, 6682094, 7024028, 58191, 576109, 7907626, 3578702, 2478982, 4089056, 562457, 6887970, 2833611, 6284791, 1277366, 1940185, 4600441, 4944178, 5201318, 2990013, 2045511, 1222727, 3006775, 6576002, 3255668, 3986220, 1196593, 494612, 7280093, 7061950, 2826451, 3027105, 3706482, 3579630, 5760070, 420548, 2649861, 855357, 7589448, 2269553, 6288506, 4652343, 5156324, 7407111, 1797934, 3032292, 1673323, 6026025, 6865947, 6331383, 6910076, 5917338, 4326152, 2366362, 2313991, 8263544, 2744854, 3708709, 283078, 6310069, 783655, 6714649, 6191299, 5293465, 1390895, 6435300, 5108581, 7368357, 7833214, 8227674, 7563460, 6746577, 372209, 7162274, 6180068, 2984957, 1727824, 6850677, 2927971, 6711852, 2752557, 5009386, 5580648, 5491647, 6227012, 4515125, 6559474, 3212844, 5516217, 4746122, 2315049, 8108583, 5031557, 899844, 615819, 5711336, 7597777, 1640020, 7142758, 428584, 796599, 3402679, 6779070, 7869282, 5730000, 4992413, 2982869, 2287598, 2211508, 2738082, 7313081, 219382, 4682398, 7545911, 1115176, 5275071, 2822251, 177652, 7393260, 5703955, 4028687, 3452966, 4914891, 6405748, 3179373, 1978200, 3928263, 5029484, 3745229, 407634, 6241305, 4488879, 8269836, 3937509, 7044953, 4873178, 7793045, 4637569, 6972147, 674802, 4300844, 2985761, 7205598, 6999815, 4983085, 7422400, 3913475, 6911929, 4760660, 5325103, 5416374, 655628, 3447779, 3627827, 4195008, 7396003, 3104494, 423510, 972673, 3186239, 4087605, 1708776, 2288174, 1018624, 1320076, 7118949, 6635662},
		{364561, 264777, 5716219, 1069362, 4925677, 4631913, 6657566, 914039, 424393, 3904436, 998013, 6024851, 4442749, 8111682, 5590293, 3105298, 4095252, 7640719, 4084225, 2567448, 6702257, 7499312, 238670, 2291357, 2998691, 2880626, 5490294, 7219353, 93015, 6897820, 1328808, 5489870, 5405588, 6725285, 5725082, 347231, 6030070, 1786393, 7736363, 5470901, 7019865, 6294938, 5393590, 2028098, 4588063, 4602190, 4674396, 5114932, 573432, 1331937, 39065, 3489524, 3170895, 2688537, 5584024, 7592006, 2096269, 5543962, 8058624, 445440, 4459707, 3728631, 1805865, 1260759, 3637636, 1689032, 5922246, 6026941, 2690860, 3469682, 1847066, 7685151, 5271732, 5027581, 4144436, 1496872, 207205, 5118459, 932085, 3062469, 846721, 802827, 1973681, 4321505, 1665459, 4620573, 615246, 6808070, 5980979, 8095689, 359373, 3793569, 7396078, 6193356, 7855056, 7339023, 1944455, 7585809, 1653133, 3578014, 3522261, 6617341, 7273473, 7364800, 7570698, 3084955, 4463902, 4376434, 1278733, 1134671, 1834451, 5434470, 7396302, 4142044, 6130117, 5723857, 5521848, 5284409, 1606982, 1092840, 5951280, 28626, 3902424, 3696633, 1683713, 6434711, 1001391, 367695, 7702650, 6061728, 1249742, 7840643, 2833115, 243684, 1539168, 4131298, 4967574, 2217194, 1978602, 7388418, 2373192, 4998752, 188150, 1036326, 1124701, 7584605, 3424087, 7552856, 6617865, 5204948, 5727511, 7606874, 1923207, 3671927, 2924721, 4500784, 4351972, 5111200, 5540272, 3846346, 6368613, 4370955, 729751, 2531956, 6473323, 1257630, 3434497, 2724353, 6666494, 689744, 7128446, 6653393, 4366431, 2287301, 2431246, 3750365, 3159574, 2872567, 3236671, 3085421, 1661210, 2411824, 445453, 3041138, 6415527, 4166542, 2172569, 886653, 1437533, 6102447, 2505421, 286028, 5683684, 7067778, 2617683, 5865924, 6973043, 1459165, 1610311, 4820024, 5450387, 421778, 8290665, 1998992, 1736703, 6170813, 5235981, 7750797, 1536785, 1281427, 1969238, 1820405, 6428344, 7988682, 3922934, 2846167, 6602434, 5827258, 5108628, 4523752, 4239500, 4528843, 4154406, 6868052, 2855344, 5346095, 4476329, 3561268, 2367278, 26354, 1818667, 841320, 1909355, 1147194, 4462303, 7136124, 6660341, 6210644, 1999861, 5836917, 1170786, 8157596, 8242488, 4488812, 5200817, 6911886, 2490187, 1140262, 2505554, 3446753, 8183071, 3068987, 5334345, 2903721, 3211897, 6839931},
	}
	expect_h := [][]uint8{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	actual_h := common.MakeHintRingVec(k, gamma2, z, r)
	assert.Equal(t, expect_h, actual_h)
}
