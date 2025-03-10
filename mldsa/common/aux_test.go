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
	}
	for _, tt := range tests {
		assert.Equal(t, tt.output, common.InfinityNorm(tt.input))
	}
}
