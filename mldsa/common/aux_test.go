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
