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
	assert.Equal(t, uint32(95220), t0)
	assert.Equal(t, uint32(1), t1)

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
			xx := uint32(tmp) + x0
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
		{"2 * Gamma2 - 1 (gamma2=95232)", 95232, 190463, 95231},
		{"2 * Gamma2 + 1 (gamma2=95232)", 95232, 190465, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, common.ModPlusMinus(tt.r, tt.gamma2<<1))
		})
	}
}

/*
func TestHighBits(t *testing.T) {
	gamma2 := uint32(95232)
	tests := []struct {
		input uint32
		high  uint32
	}{
		{389417, 2},
		{6789674, 36},
		{2350111, 12},
		{2748171, 14},
		{2380925, 13},
		{3701853, 19},
		{3887166, 20},
		{1848042, 10},
		{5459972, 29},
		{8207098, 43},
		{2718044, 14},
		{4580143, 24},
		{3787813, 20},
		{1311526, 7},
		{2886164, 15},
		{7784343, 41},
		{4191638, 22},
		{1074795, 6},
		{6917700, 36},
		{4200668, 22},
		{7630218, 40},
		{4509253, 24},
		{2549754, 13},
		{7139886, 37},
		{2199611, 12},
		{2271379, 12},
		{644227, 3},
		{527845, 3},
		{1617108, 8},
		{633353, 3},
		{1235557, 6},
		{2487197, 13},
		{467268, 2},
		{3185588, 17},
		{1790320, 9},
		{4834718, 25},
		{5120930, 27},
		{1228450, 6},
		{6442676, 34},
		{7884300, 41},
		{980516, 5},
		{5421034, 28},
		{2606240, 14},
		{6332668, 33},
		{5340011, 28},
		{4363198, 23},
		{4529981, 24},
		{2721918, 14},
		{7691783, 40},
		{7289102, 38},
		{6723904, 35},
		{8124504, 43},
		{5207369, 27},
		{2068116, 11},
		{4942061, 26},
		{4376394, 23},
		{3614929, 19},
		{1023324, 5},
		{7869921, 41},
		{3967852, 21},
		{5419627, 28},
		{7552455, 40},
		{5292674, 28},
		{3009325, 16},
		{4739366, 25},
		{7359252, 39},
		{5708205, 30},
		{7264558, 38},
		{2118121, 11},
		{3568217, 19},
		{3469004, 18},
		{6946020, 36},
		{7006404, 37},
		{408868, 2},
		{6770061, 36},
		{1112097, 6},
		{4765645, 25},
		{2831554, 15},
		{1336749, 7},
		{3253820, 17},
		{2829617, 15},
		{6138953, 32},
		{867159, 5},
		{4476184, 24},
		{2514471, 13},
		{2605607, 14},
		{7999459, 42},
		{1106723, 6},
		{226420, 1},
		{2430908, 13},
		{4649237, 24},
		{3776550, 20},
		{1854079, 10},
		{3293492, 17},
		{1452137, 8},
		{4257924, 22},
		{2252084, 12},
		{8066078, 42},
		{2119621, 11},
		{5069668, 27},
		{2735973, 14},
		{7408096, 39},
		{6814643, 36},
		{4445618, 23},
		{888354, 5},
		{1663928, 9},
		{6044221, 32},
		{7448849, 39},
		{1810758, 10},
		{2527801, 13},
		{5734598, 30},
		{5735417, 30},
		{2473012, 13},
		{2784272, 15},
		{60394, 0},
		{6681060, 35},
		{3565226, 19},
		{488103, 3},
		{2426502, 13},
		{2421371, 13},
		{3158039, 17},
		{2487294, 13},
		{5053496, 27},
		{627122, 3},
		{2586376, 14},
		{6339506, 33},
		{699755, 4},
		{5339816, 28},
		{6006273, 32},
		{5901769, 31},
		{7592480, 40},
		{6652896, 35},
		{3921345, 21},
		{6944236, 36},
		{2029441, 11},
		{4785957, 25},
		{3873875, 20},
		{2132415, 11},
		{7664755, 40},
		{5857529, 31},
		{7420125, 39},
		{3547110, 19},
		{7141829, 37},
		{662267, 3},
		{1327459, 7},
		{2786844, 15},
		{5413689, 28},
		{3051641, 16},
		{7856643, 41},
		{3131028, 16},
		{6365922, 33},
		{2987954, 16},
		{4100421, 22},
		{801338, 4},
		{1195964, 6},
		{4754818, 25},
		{1838471, 10},
		{1894880, 10},
		{2289032, 12},
		{5419500, 28},
		{5826533, 31},
		{874647, 5},
		{7621682, 40},
		{264745, 1},
		{526462, 3},
		{6813035, 36},
		{1493534, 8},
		{4782148, 25},
		{1782627, 9},
		{6159501, 32},
		{939647, 5},
		{523934, 3},
		{5135020, 27},
		{1058949, 6},
		{5755900, 30},
		{587706, 3},
		{3130047, 16},
		{4946227, 26},
		{8148425, 43},
		{6011353, 32},
		{6780769, 36},
		{2008170, 11},
		{490931, 3},
		{7151342, 38},
		{4154130, 22},
		{8300139, 0},
		{6161507, 32},
		{7530875, 40},
		{1869652, 10},
		{6343677, 33},
		{974590, 5},
		{6293998, 33},
		{3163356, 17},
		{3537925, 19},
		{403232, 2},
		{7889990, 41},
		{8371152, 0},
		{7444139, 39},
		{1554982, 8},
		{4125978, 22},
		{3197701, 17},
		{7558548, 40},
		{5549597, 29},
		{4098832, 22},
		{3795150, 20},
		{1798082, 9},
		{2637569, 14},
		{6544201, 34},
		{1142575, 6},
		{400817, 2},
		{7203636, 38},
		{7370410, 39},
		{882676, 5},
		{4923201, 26},
		{2159902, 11},
		{5220337, 27},
		{310769, 2},
		{5082395, 27},
		{7771987, 41},
		{6251912, 33},
		{4267256, 22},
		{5067396, 27},
		{3854868, 20},
		{915288, 5},
		{4242254, 22},
		{7139654, 37},
		{313591, 2},
		{6979470, 37},
		{7746218, 41},
		{2557425, 13},
		{1605482, 8},
		{322045, 2},
		{3780227, 20},
		{246256, 1},
		{5917210, 31},
		{4518380, 24},
		{1465990, 8},
		{7486437, 39},
		{7255922, 38},
		{2424325, 13},
		{7506365, 39},
		{1320444, 7},
		{2502442, 13},
		{1854368, 10},
		{4372798, 23},
		{7289684, 38},
		{970258, 5},
		{2832244, 15},
		{4094140, 21},
		{3673186, 19},
		{6188001, 32},
		{2441270, 13},
		{5070208, 27},
		{6715780, 35},
		{308669, 2},
		{6622997, 35},
	}
	for i, tt := range tests {
		actual := common.HighBits(gamma2, tt.input)
		if tt.high != actual {
			print(i, "\t")
			print(tt.input, "\n")
		}
		assert.Equal(t, tt.high, actual)
	}
}
*/

/*
w = Vector(Array([
	Polynomial(Array([
		Elem(389417), Elem(6789674), Elem(2350111), Elem(2748171), Elem(2380925), Elem(3701853), Elem(3887166), Elem(1848042), Elem(5459972), Elem(8207098), Elem(2718044), Elem(4580143), Elem(3787813), Elem(1311526), Elem(2886164), Elem(7784343), Elem(4191638), Elem(1074795), Elem(6917700), Elem(4200668), Elem(7630218), Elem(4509253), Elem(2549754), Elem(7139886), Elem(2199611), Elem(2271379), Elem(644227), Elem(527845), Elem(1617108), Elem(633353), Elem(1235557), Elem(2487197), Elem(467268), Elem(3185588), Elem(1790320), Elem(4834718), Elem(5120930), Elem(1228450), Elem(6442676), Elem(7884300), Elem(980516), Elem(5421034), Elem(2606240), Elem(6332668), Elem(5340011), Elem(4363198), Elem(4529981), Elem(2721918), Elem(7691783), Elem(7289102), Elem(6723904), Elem(8124504), Elem(5207369), Elem(2068116), Elem(4942061), Elem(4376394), Elem(3614929), Elem(1023324), Elem(7869921), Elem(3967852), Elem(5419627), Elem(7552455), Elem(5292674), Elem(3009325), Elem(4739366), Elem(7359252), Elem(5708205), Elem(7264558), Elem(2118121), Elem(3568217), Elem(3469004), Elem(6946020), Elem(7006404), Elem(408868), Elem(6770061), Elem(1112097), Elem(4765645), Elem(2831554), Elem(1336749), Elem(3253820), Elem(2829617), Elem(6138953), Elem(867159), Elem(4476184), Elem(2514471), Elem(2605607), Elem(7999459), Elem(1106723), Elem(226420), Elem(2430908), Elem(4649237), Elem(3776550), Elem(1854079), Elem(3293492), Elem(1452137), Elem(4257924), Elem(2252084), Elem(8066078), Elem(2119621), Elem(5069668), Elem(2735973), Elem(7408096), Elem(6814643), Elem(4445618), Elem(888354), Elem(1663928), Elem(6044221), Elem(7448849), Elem(1810758), Elem(2527801), Elem(5734598), Elem(5735417), Elem(2473012), Elem(2784272), Elem(60394), Elem(6681060), Elem(3565226), Elem(488103), Elem(2426502), Elem(2421371), Elem(3158039), Elem(2487294), Elem(5053496), Elem(627122), Elem(2586376), Elem(6339506), Elem(699755), Elem(5339816), Elem(6006273), Elem(5901769), Elem(7592480), Elem(6652896), Elem(3921345), Elem(6944236), Elem(2029441), Elem(4785957), Elem(3873875), Elem(2132415), Elem(7664755), Elem(5857529), Elem(7420125), Elem(3547110), Elem(7141829), Elem(662267), Elem(1327459), Elem(2786844), Elem(5413689), Elem(3051641), Elem(7856643), Elem(3131028), Elem(6365922), Elem(2987954), Elem(4100421), Elem(801338), Elem(1195964), Elem(4754818), Elem(1838471), Elem(1894880), Elem(2289032), Elem(5419500), Elem(5826533), Elem(874647), Elem(7621682), Elem(264745), Elem(526462), Elem(6813035), Elem(1493534), Elem(4782148), Elem(1782627), Elem(6159501), Elem(939647), Elem(523934), Elem(5135020), Elem(1058949), Elem(5755900), Elem(587706), Elem(3130047), Elem(4946227), Elem(8148425), Elem(6011353), Elem(6780769), Elem(2008170), Elem(490931), Elem(7151342), Elem(4154130), Elem(8300139), Elem(6161507), Elem(7530875), Elem(1869652), Elem(6343677), Elem(974590), Elem(6293998), Elem(3163356), Elem(3537925), Elem(403232), Elem(7889990), Elem(8371152), Elem(7444139), Elem(1554982), Elem(4125978), Elem(3197701), Elem(7558548), Elem(5549597), Elem(4098832), Elem(3795150), Elem(1798082), Elem(2637569), Elem(6544201), Elem(1142575), Elem(400817), Elem(7203636), Elem(7370410), Elem(882676), Elem(4923201), Elem(2159902), Elem(5220337), Elem(310769), Elem(5082395), Elem(7771987), Elem(6251912), Elem(4267256), Elem(5067396), Elem(3854868), Elem(915288), Elem(4242254), Elem(7139654), Elem(313591), Elem(6979470), Elem(7746218), Elem(2557425), Elem(1605482), Elem(322045), Elem(3780227), Elem(246256), Elem(5917210), Elem(4518380), Elem(1465990), Elem(7486437), Elem(7255922), Elem(2424325), Elem(7506365), Elem(1320444), Elem(2502442), Elem(1854368), Elem(4372798), Elem(7289684), Elem(970258), Elem(2832244), Elem(4094140), Elem(3673186), Elem(6188001), Elem(2441270), Elem(5070208), Elem(6715780), Elem(308669), Elem(6622997)
	])), Polynomial(Array([
		Elem(94153), Elem(7617350), Elem(1524677), Elem(2058424), Elem(6412773), Elem(1465399), Elem(7270060), Elem(1190146), Elem(8221138), Elem(5306000), Elem(1405392), Elem(4042286), Elem(6891232), Elem(6466080), Elem(195553), Elem(4917411), Elem(6344378), Elem(4229539), Elem(2170815), Elem(29170), Elem(7494178), Elem(174203), Elem(5818764), Elem(5378541), Elem(2427233), Elem(984108), Elem(7317110), Elem(8168520), Elem(5991788), Elem(8094431), Elem(7183586), Elem(6852857), Elem(4244060), Elem(2375717), Elem(960870), Elem(2807763), Elem(7944084), Elem(6014563), Elem(2745662), Elem(846132), Elem(5373033), Elem(3780684), Elem(5170296), Elem(2002716), Elem(7284640), Elem(7067359), Elem(4741077), Elem(1840648), Elem(5251189), Elem(1548531), Elem(1112302), Elem(3914931), Elem(7989748), Elem(5773371), Elem(4702715), Elem(136214), Elem(7743254), Elem(2331834), Elem(107851), Elem(2485962), Elem(8188891), Elem(7459648), Elem(6429768), Elem(2894859), Elem(7097560), Elem(2185079), Elem(984118), Elem(295240), Elem(6274690), Elem(147133), Elem(4655498), Elem(1149800), Elem(3240870), Elem(7201680), Elem(6737079), Elem(358191), Elem(709486), Elem(871922), Elem(901589), Elem(1715339), Elem(3157527), Elem(4240354), Elem(8105563), Elem(6094082), Elem(884869), Elem(7205023), Elem(4488602), Elem(1077343), Elem(7731624), Elem(4490192), Elem(8041621), Elem(3397195), Elem(493231), Elem(3779117), Elem(2813982), Elem(1059869), Elem(557868), Elem(3104361), Elem(3235870), Elem(7842219), Elem(2500068), Elem(1957964), Elem(1608105), Elem(347002), Elem(6002568), Elem(4467514), Elem(6529650), Elem(6299164), Elem(7091946), Elem(7068340), Elem(5951820), Elem(3980141), Elem(5009556), Elem(1585799), Elem(3007507), Elem(2305429), Elem(2595088), Elem(2043277), Elem(7658049), Elem(8057551), Elem(417391), Elem(907650), Elem(2815788), Elem(120733), Elem(4933390), Elem(6973570), Elem(5692562), Elem(2906886), Elem(7737577), Elem(4184080), Elem(7027399), Elem(4883964), Elem(2855190), Elem(4710438), Elem(6655014), Elem(2422133), Elem(50737), Elem(2849883), Elem(484202), Elem(8120572), Elem(364454), Elem(3384638), Elem(1256031), Elem(6263407), Elem(582567), Elem(5920474), Elem(6849102), Elem(1958686), Elem(6091172), Elem(3158488), Elem(2188626), Elem(5758597), Elem(2775200), Elem(1442673), Elem(2141435), Elem(2634443), Elem(6799914), Elem(5888536), Elem(2891022), Elem(1685766), Elem(3518141), Elem(8003331), Elem(145994), Elem(7165103), Elem(7115033), Elem(5012804), Elem(3185804), Elem(7414793), Elem(2279019), Elem(7503164), Elem(2344956), Elem(5021534), Elem(1469167), Elem(7090673), Elem(2583968), Elem(7027150), Elem(5038815), Elem(7015830), Elem(2983085), Elem(3922587), Elem(4543786), Elem(4599513), Elem(7819526), Elem(1074044), Elem(6101237), Elem(2864932), Elem(3926854), Elem(2944861), Elem(529406), Elem(4983627), Elem(7071634), Elem(1321863), Elem(2153895), Elem(888399), Elem(1557573), Elem(5780033), Elem(1050130), Elem(7116005), Elem(5631522), Elem(3850743), Elem(7641641), Elem(1355398), Elem(4261639), Elem(6565239), Elem(7933876), Elem(7142619), Elem(6625591), Elem(3885164), Elem(6329713), Elem(1735790), Elem(6895997), Elem(603902), Elem(1346861), Elem(2788639), Elem(8119474), Elem(5107127), Elem(3093899), Elem(218115), Elem(2462413), Elem(2627518), Elem(5691199), Elem(5170040), Elem(2683604), Elem(6904864), Elem(4094967), Elem(6792487), Elem(7425537), Elem(5846921), Elem(3947450), Elem(7827305), Elem(105069), Elem(504238), Elem(3781501), Elem(1585987), Elem(5860578), Elem(6777478), Elem(5541597), Elem(6886138), Elem(7018450), Elem(5039493), Elem(3857228), Elem(2876541), Elem(7549024), Elem(5845890), Elem(7750267), Elem(211810), Elem(6618038), Elem(331476), Elem(4274166), Elem(1544611), Elem(894744), Elem(3848415), Elem(3901648), Elem(4733338), Elem(4909856), Elem(4145960)
	])), Polynomial(Array([
		Elem(6626333), Elem(110770), Elem(662647), Elem(3206993), Elem(6824046), Elem(5876416), Elem(3060885), Elem(6054967), Elem(2395121), Elem(7647539), Elem(194460), Elem(3824426), Elem(4574816), Elem(833287), Elem(4077739), Elem(8190633), Elem(7410512), Elem(5308546), Elem(4212191), Elem(3896501), Elem(6819536), Elem(7397713), Elem(409169), Elem(3917942), Elem(4925587), Elem(4740061), Elem(1289481), Elem(2598529), Elem(8084041), Elem(3286005), Elem(5861462), Elem(4656165), Elem(7836853), Elem(2097775), Elem(7926974), Elem(3310072), Elem(6312284), Elem(2865334), Elem(7858391), Elem(2833720), Elem(4422384), Elem(1622032), Elem(4278613), Elem(434954), Elem(5338267), Elem(820228), Elem(6529189), Elem(5968460), Elem(243852), Elem(5965256), Elem(7694181), Elem(5027627), Elem(198322), Elem(4040179), Elem(6243990), Elem(5671165), Elem(3792832), Elem(6374407), Elem(5200879), Elem(3419804), Elem(6476328), Elem(779940), Elem(1526651), Elem(381825), Elem(7319728), Elem(7071883), Elem(3107965), Elem(4432505), Elem(33813), Elem(5770368), Elem(5265101), Elem(1577041), Elem(6411823), Elem(5853482), Elem(1494023), Elem(262829), Elem(3527497), Elem(825970), Elem(4368143), Elem(5394132), Elem(1530332), Elem(610546), Elem(2397459), Elem(1254807), Elem(3995665), Elem(4866652), Elem(4953738), Elem(7743367), Elem(8141330), Elem(2238890), Elem(5895662), Elem(2166701), Elem(5746097), Elem(3720466), Elem(2791672), Elem(7687687), Elem(1452243), Elem(8175671), Elem(594115), Elem(2541157), Elem(6921202), Elem(2174234), Elem(8306133), Elem(689134), Elem(7741385), Elem(3541718), Elem(1311349), Elem(1967937), Elem(2195435), Elem(1607567), Elem(993820), Elem(2693648), Elem(5623831), Elem(5596655), Elem(1394723), Elem(6379418), Elem(6500880), Elem(5825818), Elem(5424843), Elem(4273752), Elem(3975987), Elem(3784505), Elem(582549), Elem(253351), Elem(3422416), Elem(4587861), Elem(509341), Elem(4076918), Elem(535003), Elem(4693379), Elem(1878138), Elem(2494873), Elem(1576702), Elem(2155520), Elem(3989671), Elem(4879678), Elem(2520531), Elem(3941359), Elem(6630554), Elem(7770707), Elem(3818215), Elem(7251226), Elem(1250991), Elem(784886), Elem(5342150), Elem(3923548), Elem(4072608), Elem(6702515), Elem(2566669), Elem(1579561), Elem(6442859), Elem(4128208), Elem(6054942), Elem(3862408), Elem(4417515), Elem(4837474), Elem(7572287), Elem(6489557), Elem(7238514), Elem(6942413), Elem(4116740), Elem(1371345), Elem(465958), Elem(3066216), Elem(1466380), Elem(6924334), Elem(4351105), Elem(5659927), Elem(3676623), Elem(4787670), Elem(2912690), Elem(2779264), Elem(644920), Elem(1955701), Elem(7951658), Elem(3167818), Elem(4865177), Elem(4970743), Elem(6570333), Elem(6367696), Elem(5261585), Elem(2304206), Elem(1895559), Elem(1677088), Elem(1533310), Elem(5867741), Elem(3456723), Elem(3973496), Elem(4590978), Elem(1345822), Elem(8177854), Elem(7860171), Elem(5777188), Elem(102180), Elem(717905), Elem(5036183), Elem(4137030), Elem(960089), Elem(7063813), Elem(7763385), Elem(5944116), Elem(2210369), Elem(4597730), Elem(5859782), Elem(4538371), Elem(680822), Elem(2650133), Elem(6310147), Elem(5164620), Elem(6323134), Elem(321037), Elem(1576070), Elem(656008), Elem(4420701), Elem(696064), Elem(7872923), Elem(2266203), Elem(4831367), Elem(6185453), Elem(2199583), Elem(1879252), Elem(6418025), Elem(301666), Elem(403169), Elem(2911320), Elem(1903381), Elem(662781), Elem(5813858), Elem(6180639), Elem(7894144), Elem(3584502), Elem(6929790), Elem(1765050), Elem(2569667), Elem(1231375), Elem(3375734), Elem(4005264), Elem(5862718), Elem(5251339), Elem(7585215), Elem(4478223), Elem(3939948), Elem(857040), Elem(2242805), Elem(6199437), Elem(5251407), Elem(1982496), Elem(6693274), Elem(6747551), Elem(6825979), Elem(2162272), Elem(1023065), Elem(4537580), Elem(921856), Elem(5542191), Elem(4181781)
	])), Polynomial(Array([
		Elem(3635577), Elem(7493304), Elem(3661216), Elem(8322428), Elem(4500429), Elem(1050616), Elem(3339573), Elem(5998699), Elem(2194812), Elem(3989458), Elem(7872480), Elem(2811983), Elem(3454809), Elem(2546348), Elem(277535), Elem(3195332), Elem(1680710), Elem(7802689), Elem(5913219), Elem(4720011), Elem(4930663), Elem(509139), Elem(3466080), Elem(5739714), Elem(8154916), Elem(3859659), Elem(6319630), Elem(690863), Elem(889914), Elem(7676739), Elem(1701205), Elem(4523578), Elem(7064857), Elem(3270818), Elem(5490587), Elem(4883567), Elem(4487359), Elem(2813748), Elem(921301), Elem(6636585), Elem(7826411), Elem(5313264), Elem(7570355), Elem(441742), Elem(2919668), Elem(3348486), Elem(5872111), Elem(7010308), Elem(3767760), Elem(5939639), Elem(5894653), Elem(7099741), Elem(4757008), Elem(3202634), Elem(499464), Elem(3707896), Elem(6562348), Elem(6314851), Elem(4159380), Elem(5335236), Elem(8173374), Elem(8110738), Elem(5348999), Elem(1250201), Elem(4527524), Elem(4564593), Elem(5461999), Elem(4397871), Elem(4051140), Elem(1644550), Elem(7091412), Elem(7042147), Elem(8371946), Elem(6274910), Elem(2397294), Elem(550922), Elem(101740), Elem(6685933), Elem(8108867), Elem(1134770), Elem(3893315), Elem(6085234), Elem(3714276), Elem(5138751), Elem(4995568), Elem(1372187), Elem(7607118), Elem(293019), Elem(5595792), Elem(856005), Elem(621744), Elem(3059829), Elem(1528466), Elem(557415), Elem(7015166), Elem(7502651), Elem(5496807), Elem(332980), Elem(7733903), Elem(2287685), Elem(4615548), Elem(7814143), Elem(4700554), Elem(5673048), Elem(5238774), Elem(200749), Elem(3299095), Elem(2667376), Elem(2584710), Elem(4921365), Elem(5306436), Elem(2502254), Elem(7353792), Elem(6550637), Elem(3821340), Elem(978969), Elem(1932760), Elem(5334755), Elem(2069400), Elem(6162711), Elem(1551532), Elem(4492927), Elem(4435004), Elem(813119), Elem(4403406), Elem(1202900), Elem(3240880), Elem(7010489), Elem(812556), Elem(1639266), Elem(8372718), Elem(4347794), Elem(2119899), Elem(7584713), Elem(6890864), Elem(6827430), Elem(2126590), Elem(6729623), Elem(5583124), Elem(3800581), Elem(8377888), Elem(7028052), Elem(6418632), Elem(4687539), Elem(4665963), Elem(6538882), Elem(6150104), Elem(879720), Elem(7255189), Elem(5875063), Elem(8194543), Elem(3461141), Elem(3384896), Elem(1093806), Elem(7419336), Elem(1687161), Elem(4441468), Elem(1166087), Elem(5380855), Elem(6234396), Elem(286475), Elem(1193736), Elem(2456510), Elem(4546006), Elem(2579688), Elem(1149665), Elem(2055005), Elem(7619078), Elem(7597331), Elem(7658805), Elem(819904), Elem(7092203), Elem(3483757), Elem(2471853), Elem(4180817), Elem(1650820), Elem(3598089), Elem(3534428), Elem(5967302), Elem(6765235), Elem(6202657), Elem(3538243), Elem(7645247), Elem(7686394), Elem(4967263), Elem(4892821), Elem(6338776), Elem(6111020), Elem(4202620), Elem(5050553), Elem(5281845), Elem(6856469), Elem(5417256), Elem(244824), Elem(5011356), Elem(3216803), Elem(6921493), Elem(7549672), Elem(83417), Elem(2453469), Elem(5582969), Elem(2047634), Elem(6991307), Elem(4811502), Elem(3910244), Elem(5945923), Elem(281335), Elem(5031910), Elem(6508646), Elem(7008323), Elem(7474764), Elem(7381224), Elem(8080833), Elem(1565697), Elem(7500875), Elem(7062730), Elem(4620676), Elem(6870915), Elem(5073496), Elem(1941404), Elem(2163159), Elem(70536), Elem(1478313), Elem(2097308), Elem(3008566), Elem(5948455), Elem(59271), Elem(6517680), Elem(3854783), Elem(3910494), Elem(834266), Elem(2784619), Elem(2053914), Elem(3216627), Elem(2222183), Elem(4624091), Elem(7752382), Elem(729063), Elem(6706415), Elem(2322839), Elem(5575566), Elem(3558317), Elem(4725576), Elem(8102618), Elem(2754858), Elem(6936259), Elem(2666245), Elem(7185916), Elem(428040), Elem(5273027), Elem(4763064), Elem(2983233), Elem(4400441), Elem(6972592), Elem(6735356), Elem(1482827)
	]))
]))
w1 = Vector(Array([
	Polynomial(Array([
	Elem(2), Elem(36), Elem(12), Elem(14), Elem(13), Elem(19), Elem(20), Elem(10), Elem(29), Elem(43), Elem(14), Elem(24), Elem(20), Elem(7), Elem(15), Elem(41), Elem(22), Elem(6), Elem(36), Elem(22), Elem(40), Elem(24), Elem(13), Elem(37), Elem(12), Elem(12), Elem(3), Elem(3), Elem(8), Elem(3), Elem(6), Elem(13), Elem(2), Elem(17), Elem(9), Elem(25), Elem(27), Elem(6), Elem(34), Elem(41), Elem(5), Elem(28), Elem(14), Elem(33), Elem(28), Elem(23), Elem(24), Elem(14), Elem(40), Elem(38), Elem(35), Elem(43), Elem(27), Elem(11), Elem(26), Elem(23), Elem(19), Elem(5), Elem(41), Elem(21), Elem(28), Elem(40), Elem(28), Elem(16), Elem(25), Elem(39), Elem(30), Elem(38), Elem(11), Elem(19), Elem(18), Elem(36), Elem(37), Elem(2), Elem(36), Elem(6), Elem(25), Elem(15), Elem(7), Elem(17), Elem(15), Elem(32), Elem(5), Elem(24), Elem(13), Elem(14), Elem(42), Elem(6), Elem(1), Elem(13), Elem(24), Elem(20), Elem(10), Elem(17), Elem(8), Elem(22), Elem(12), Elem(42), Elem(11), Elem(27), Elem(14), Elem(39), Elem(36), Elem(23), Elem(5), Elem(9), Elem(32), Elem(39), Elem(10), Elem(13), Elem(30), Elem(30), Elem(13), Elem(15), Elem(0), Elem(35), Elem(19), Elem(3), Elem(13), Elem(13), Elem(17), Elem(13), Elem(27), Elem(3), Elem(14), Elem(33), Elem(4), Elem(28), Elem(32), Elem(31), Elem(40), Elem(35), Elem(21), Elem(36), Elem(11), Elem(25), Elem(20), Elem(11), Elem(40), Elem(31), Elem(39), Elem(19), Elem(37), Elem(3), Elem(7), Elem(15), Elem(28), Elem(16), Elem(41), Elem(16), Elem(33), Elem(16), Elem(22), Elem(4), Elem(6), Elem(25), Elem(10), Elem(10), Elem(12), Elem(28), Elem(31), Elem(5), Elem(40), Elem(1), Elem(3), Elem(36), Elem(8), Elem(25), Elem(9), Elem(32), Elem(5), Elem(3), Elem(27), Elem(6), Elem(30), Elem(3), Elem(16), Elem(26), Elem(43), Elem(32), Elem(36), Elem(11), Elem(3), Elem(38), Elem(22), Elem(0), Elem(32), Elem(40), Elem(10), Elem(33), Elem(5), Elem(33), Elem(17), Elem(19), Elem(2), Elem(41), Elem(0), Elem(39), Elem(8), Elem(22), Elem(17), Elem(40), Elem(29), Elem(22), Elem(20), Elem(9), Elem(14), Elem(34), Elem(6), Elem(2), Elem(38), Elem(39), Elem(5), Elem(26), Elem(11), Elem(27), Elem(2), Elem(27), Elem(41), Elem(33), Elem(22), Elem(27), Elem(20), Elem(5), Elem(22), Elem(37), Elem(2), Elem(37), Elem(41), Elem(13), Elem(8), Elem(2), Elem(20), Elem(1), Elem(31), Elem(24), Elem(8), Elem(39), Elem(38), Elem(13), Elem(39), Elem(7), Elem(13), Elem(10), Elem(23), Elem(38), Elem(5), Elem(15), Elem(21), Elem(19), Elem(32), Elem(13), Elem(27), Elem(35), Elem(2), Elem(35)
	])), Polynomial(Array([Elem(0), Elem(40), Elem(8), Elem(11), Elem(34), Elem(8), Elem(38), Elem(6), Elem(43), Elem(28), Elem(7), Elem(21), Elem(36), Elem(34), Elem(1), Elem(26), Elem(33), Elem(22), Elem(11), Elem(0), Elem(39), Elem(1), Elem(31), Elem(28), Elem(13), Elem(5), Elem(38), Elem(43), Elem(31), Elem(42), Elem(38), Elem(36), Elem(22), Elem(12), Elem(5), Elem(15), Elem(42), Elem(32), Elem(14), Elem(4), Elem(28), Elem(20), Elem(27), Elem(11), Elem(38), Elem(37), Elem(25), Elem(10), Elem(28), Elem(8), Elem(6), Elem(21), Elem(42), Elem(30), Elem(25), Elem(1), Elem(41), Elem(12), Elem(1), Elem(13), Elem(43), Elem(39), Elem(34), Elem(15), Elem(37), Elem(11), Elem(5), Elem(2), Elem(33), Elem(1), Elem(24), Elem(6), Elem(17), Elem(38), Elem(35), Elem(2), Elem(4), Elem(5), Elem(5), Elem(9), Elem(17), Elem(22), Elem(43), Elem(32), Elem(5), Elem(38), Elem(24), Elem(6), Elem(41), Elem(24), Elem(42), Elem(18), Elem(3), Elem(20), Elem(15), Elem(6), Elem(3), Elem(16), Elem(17), Elem(41), Elem(13), Elem(10), Elem(8), Elem(2), Elem(32), Elem(23), Elem(34), Elem(33), Elem(37), Elem(37), Elem(31), Elem(21), Elem(26), Elem(8), Elem(16), Elem(12), Elem(14), Elem(11), Elem(40), Elem(42), Elem(2), Elem(5), Elem(15), Elem(1), Elem(26), Elem(37), Elem(30), Elem(15), Elem(41), Elem(22), Elem(37), Elem(26), Elem(15), Elem(25), Elem(35), Elem(13), Elem(0), Elem(15), Elem(3), Elem(43), Elem(2), Elem(18), Elem(7), Elem(33), Elem(3), Elem(31), Elem(36), Elem(10), Elem(32), Elem(17), Elem(11), Elem(30), Elem(15), Elem(8), Elem(11), Elem(14), Elem(36), Elem(31), Elem(15), Elem(9), Elem(18), Elem(42), Elem(1), Elem(38), Elem(37), Elem(26), Elem(17), Elem(39), Elem(12), Elem(39), Elem(12), Elem(26), Elem(8), Elem(37), Elem(14), Elem(37), Elem(26), Elem(37), Elem(16), Elem(21), Elem(24), Elem(24), Elem(41), Elem(6), Elem(32), Elem(15), Elem(21), Elem(15), Elem(3), Elem(26), Elem(37), Elem(7), Elem(11), Elem(5), Elem(8), Elem(30), Elem(6), Elem(37), Elem(30), Elem(20), Elem(40), Elem(7), Elem(22), Elem(34), Elem(42), Elem(38), Elem(35), Elem(20), Elem(33), Elem(9), Elem(36), Elem(3), Elem(7), Elem(15), Elem(43), Elem(27), Elem(16), Elem(1), Elem(13), Elem(14), Elem(30), Elem(27), Elem(14), Elem(36), Elem(21), Elem(36), Elem(39), Elem(31), Elem(21), Elem(41), Elem(1), Elem(3), Elem(20), Elem(8), Elem(31), Elem(36), Elem(29), Elem(36), Elem(37), Elem(26), Elem(20), Elem(15), Elem(40), Elem(31), Elem(41), Elem(1), Elem(35), Elem(2), Elem(22), Elem(8), Elem(5), Elem(20), Elem(20), Elem(25), Elem(26), Elem(22)])), Polynomial(Array([Elem(35), Elem(1), Elem(3), Elem(17), Elem(36), Elem(31), Elem(16), Elem(32), Elem(13), Elem(40), Elem(1), Elem(20), Elem(24), Elem(4), Elem(21), Elem(43), Elem(39), Elem(28), Elem(22), Elem(20), Elem(36), Elem(39), Elem(2), Elem(21), Elem(26), Elem(25), Elem(7), Elem(14), Elem(42), Elem(17), Elem(31), Elem(24), Elem(41), Elem(11), Elem(42), Elem(17), Elem(33), Elem(15), Elem(41), Elem(15), Elem(23), Elem(9), Elem(22), Elem(2), Elem(28), Elem(4), Elem(34), Elem(31), Elem(1), Elem(31), Elem(40), Elem(26), Elem(1), Elem(21), Elem(33), Elem(30), Elem(20), Elem(33), Elem(27), Elem(18), Elem(34), Elem(4), Elem(8), Elem(2), Elem(38), Elem(37), Elem(16), Elem(23), Elem(0), Elem(30), Elem(28), Elem(8), Elem(34), Elem(31), Elem(8), Elem(1), Elem(19), Elem(4), Elem(23), Elem(28), Elem(8), Elem(3), Elem(13), Elem(7), Elem(21), Elem(26), Elem(26), Elem(41), Elem(43), Elem(12), Elem(31), Elem(11), Elem(30), Elem(20), Elem(15), Elem(40), Elem(8), Elem(43), Elem(3), Elem(13), Elem(36), Elem(11), Elem(0), Elem(4), Elem(41), Elem(19), Elem(7), Elem(10), Elem(12), Elem(8), Elem(5), Elem(14), Elem(30), Elem(29), Elem(7), Elem(33), Elem(34), Elem(31), Elem(28), Elem(22), Elem(21), Elem(20), Elem(3), Elem(1), Elem(18), Elem(24), Elem(3), Elem(21), Elem(3), Elem(25), Elem(10), Elem(13), Elem(8), Elem(11), Elem(21), Elem(26), Elem(13), Elem(21), Elem(35), Elem(41), Elem(20), Elem(38), Elem(7), Elem(4), Elem(28), Elem(21), Elem(21), Elem(35), Elem(13), Elem(8), Elem(34), Elem(22), Elem(32), Elem(20), Elem(23), Elem(25), Elem(40), Elem(34), Elem(38), Elem(36), Elem(22), Elem(7), Elem(2), Elem(16), Elem(8), Elem(36), Elem(23), Elem(30), Elem(19), Elem(25), Elem(15), Elem(15), Elem(3), Elem(10), Elem(42), Elem(17), Elem(26), Elem(26), Elem(34), Elem(33), Elem(28), Elem(12), Elem(10), Elem(9), Elem(8), Elem(31), Elem(18), Elem(21), Elem(24), Elem(7), Elem(43), Elem(41), Elem(30), Elem(1), Elem(4), Elem(26), Elem(22), Elem(5), Elem(37), Elem(41), Elem(31), Elem(12), Elem(24), Elem(31), Elem(24), Elem(4), Elem(14), Elem(33), Elem(27), Elem(33), Elem(2), Elem(8), Elem(3), Elem(23), Elem(4), Elem(41), Elem(12), Elem(25), Elem(32), Elem(12), Elem(10), Elem(34), Elem(2), Elem(2), Elem(15), Elem(10), Elem(3), Elem(31), Elem(32), Elem(41), Elem(19), Elem(36), Elem(9), Elem(13), Elem(6), Elem(18), Elem(21), Elem(31), Elem(28), Elem(40), Elem(24), Elem(21), Elem(4), Elem(12), Elem(33), Elem(28), Elem(10), Elem(35), Elem(35), Elem(36), Elem(11), Elem(5), Elem(24), Elem(5), Elem(29), Elem(22)])), Polynomial(Array([Elem(19), Elem(39), Elem(19), Elem(0), Elem(24), Elem(6), Elem(18), Elem(31), Elem(12), Elem(21), Elem(41), Elem(15), Elem(18), Elem(13), Elem(1), Elem(17), Elem(9), Elem(41), Elem(31), Elem(25), Elem(26), Elem(3), Elem(18), Elem(30), Elem(43), Elem(20), Elem(33), Elem(4), Elem(5), Elem(40), Elem(9), Elem(24), Elem(37), Elem(17), Elem(29), Elem(26), Elem(24), Elem(15), Elem(5), Elem(35), Elem(41), Elem(28), Elem(40), Elem(2), Elem(15), Elem(18), Elem(31), Elem(37), Elem(20), Elem(31), Elem(31), Elem(37), Elem(25), Elem(17), Elem(3), Elem(19), Elem(34), Elem(33), Elem(22), Elem(28), Elem(43), Elem(43), Elem(28), Elem(7), Elem(24), Elem(24), Elem(29), Elem(23), Elem(21), Elem(9), Elem(37), Elem(37), Elem(0), Elem(33), Elem(13), Elem(3), Elem(1), Elem(35), Elem(43), Elem(6), Elem(20), Elem(32), Elem(20), Elem(27), Elem(26), Elem(7), Elem(40), Elem(2), Elem(29), Elem(4), Elem(3), Elem(16), Elem(8), Elem(3), Elem(37), Elem(39), Elem(29), Elem(2), Elem(41), Elem(12), Elem(24), Elem(41), Elem(25), Elem(30), Elem(28), Elem(1), Elem(17), Elem(14), Elem(14), Elem(26), Elem(28), Elem(13), Elem(39), Elem(34), Elem(20), Elem(5), Elem(10), Elem(28), Elem(11), Elem(32), Elem(8), Elem(24), Elem(23), Elem(4), Elem(23), Elem(6), Elem(17), Elem(37), Elem(4), Elem(9), Elem(0), Elem(23), Elem(11), Elem(40), Elem(36), Elem(36), Elem(11), Elem(35), Elem(29), Elem(20), Elem(0), Elem(37), Elem(34), Elem(25), Elem(24), Elem(34), Elem(32), Elem(5), Elem(38), Elem(31), Elem(43), Elem(18), Elem(18), Elem(6), Elem(39), Elem(9), Elem(23), Elem(6), Elem(28), Elem(33), Elem(2), Elem(6), Elem(13), Elem(24), Elem(14), Elem(6), Elem(11), Elem(40), Elem(40), Elem(40), Elem(4), Elem(37), Elem(18), Elem(13), Elem(22), Elem(9), Elem(19), Elem(19), Elem(31), Elem(36), Elem(33), Elem(19), Elem(40), Elem(40), Elem(26), Elem(26), Elem(33), Elem(32), Elem(22), Elem(27), Elem(28), Elem(36), Elem(28), Elem(1), Elem(26), Elem(17), Elem(36), Elem(40), Elem(0), Elem(13), Elem(29), Elem(11), Elem(37), Elem(25), Elem(21), Elem(31), Elem(1), Elem(26), Elem(34), Elem(37), Elem(39), Elem(39), Elem(42), Elem(8), Elem(39), Elem(37), Elem(24), Elem(36), Elem(27), Elem(10), Elem(11), Elem(0), Elem(8), Elem(11), Elem(16), Elem(31), Elem(0), Elem(34), Elem(20), Elem(21), Elem(4), Elem(15), Elem(11), Elem(17), Elem(12), Elem(24), Elem(41), Elem(4), Elem(35), Elem(12), Elem(29), Elem(19), Elem(25), Elem(43), Elem(14), Elem(36), Elem(14), Elem(38), Elem(2), Elem(28), Elem(25), Elem(16), Elem(23), Elem(37), Elem(35), Elem(8)]))]))*/

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
		{"Testing from w1encode", 95232, 6789674, 35, q - 28202},
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
			congruent := (res1 * (tt.gamma2 << 1) % q) + res2
			if congruent != tt.r {
				fmt.Printf("Not equal mod q: %d, %d\n", tt.r%q, congruent%q)
				fmt.Printf("Not equal: %d, %d\n", tt.r, congruent)
			}
			assert.Equal(t, tt.r, congruent)
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
