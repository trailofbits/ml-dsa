package common

import (
	"errors"
	"math/bits"
	"slices"

	"golang.org/x/crypto/sha3"
)

// Algorithm 9
// Input:
//
//	x -- non-negative integer
//	a -- positive integer
//
// Returns a []byte with a distinct byte entry for each bit
func Uint32ToBits(x uint32, a int) []byte {
	var xprime = x
	var y []byte = make([]byte, a)
	for i := range a {
		y[i] = byte(xprime & 1)
		xprime = xprime >> 1
	}
	return y
}

// Algorithm 10
// Input:
//
//	y -- []byte with each value set to 1, of length a
//	a -- positive integer
//
// Output: non-negative integer
func BitsToInteger(y []byte, a int) uint32 {
	x := uint32(0)
	for i := range a {
		x = (x << 1) | uint32(y[a-i-1])
	}
	return x
}

func IntegerToBytes(x uint32, a int) []byte {
	xp := x
	y := make([]byte, a)
	for i := range a {
		y[i] = byte(xp & 0xff)
		xp >>= 8
	}
	return y
}
func IntegerToBits(x uint32, b int) []byte {
	return BitsToBytes(IntegerToBytes(x, b))
}

// Algorithm 12
// Here, y is a []byte with each value set to 0 or 1.
// This is because we can't get more memory-efficient than an array of bytes.
func BitsToBytes(y []byte) []byte {
	var a = len(y)
	// Add 7 then right shift by 3 to get a size that's always rounded up to a whole byte
	var z []byte = make([]byte, (a+7)>>3)
	for i := range a {
		i_8 := i >> 3
		z[i_8] |= y[i] << (i & 7)
	}
	return z
}

// Algorithm 13
// Input: array of bytes
// Output: array of bits (but also a []byte type)
func BytesToBits(z []byte) []byte {
	zprime := slices.Clone(z)
	len := len(z)
	y := make([]byte, len<<3)
	for i := range len {
		for j := range 8 {
			index := (i << 3) + j
			y[index] = byte(zprime[i] & 1)
			zprime[i] >>= 1
		}
	}
	return y
}

// Algorithm 14
// Inputs: 3 bytes
// Output: integer mod q or error
func CoeffFromThreeBytes(b0, b1, b2 byte) (FieldElement, error) {
	bp2 := uint32(b2 & 0x7f)
	z := (bp2 << 16) | uint32(b1)<<8 | uint32(b0)

	// if q >= z, return an error
	invalid := (q - (z + 1)) >> 31
	if invalid != 0 {
		return FieldElement(0), errors.New("reject this sample")
	}
	return FieldReduceOnce(z), nil
}

// Algorithm 15
// Input: eta, byte
// Output: integer between -eta and eta or an error
func CoeffFromHalfByte(eta int, b byte) (RingCoeff, error) {
	if eta == 2 && b < 15 {
		return CoeffReduceOnce(q + 2 - uint32(b%5)), nil
	}
	if eta == 4 && b < 9 {
		return CoeffReduceOnce(q + 4 - uint32(b)), nil
	}
	return RingCoeff(0), errors.New("reject this sample")
}

// Algorithm 16
func SimpleBitPack(w RingElement, b uint32) []byte {
	var bitlen int = bits.Len32(b)
	var z []byte
	q2 := uint32(q >> 1)
	for i := range 256 {
		wi := uint32(w[i])
		wi -= -((q2 - wi) >> 31) & q
		suffix := Uint32ToBits(wi, bitlen)
		z = append(z, suffix[0:bitlen]...)
	}
	return BitsToBytes(z)
}

// Algorithm 17
func BitPack(w RingElement, a, b uint32) []byte {
	var z []byte
	bitlen := bits.Len32(uint32(a + b))
	for i := range 256 {
		var diff uint32
		wi := uint32(w[i])
		diff = b - wi
		diff += -(diff >> 31) & q
		bits := Uint32ToBits(diff, bitlen)
		z = append(z, bits[0:bitlen]...)
	}
	// print(hex.EncodeToString(z))

	return BitsToBytes(z)
}

// Algorithm 18
// As of 2025-02-26, I suspect this has bugs
func SimpleBitUnpack(v []byte, b uint32) (w RingElement) {
	c := bits.Len32(b)
	z := BytesToBits(v)
	for i := range 256 {
		start := i * c
		end := start + c
		bits := z[start:end]
		w[i] = CoeffReduceOnce(BitsToInteger(bits, c))
	}
	return w
}

// Algorithm 19
func BitUnpack(v []byte, a, b uint32) (w RingElement) {
	c := bits.Len32(a + b)
	z := BytesToBits(v)
	for i := range 256 {
		start := i * c
		stop := start + c
		bits := BitsToInteger(z[start:stop], c)
		diff := b - bits
		diff += (diff >> 31) * q
		w[i] = CoeffReduceOnce(diff)
	}
	return w
}

// Algorithm 20
func HintBitPack(k, omega uint8, h RingVector) []byte {
	y := make([]byte, k+omega)
	index := uint8(0)
	for i := range k {
		for j := range 256 {
			// if h[i][j] != 0 {
			//    y[index] = j
			//    index++
			// }

			// Implement the conditional without a branch
			swap := uint8(((h[i][j] - 1) >> 15) & 1)
			mask := byte(-swap)
			jb := byte(j)
			y[index] = (jb & ^mask) ^ (y[index] & mask)
			index += 1 - swap
		}
		y[omega+i] = byte(index)
	}
	return y
}

// Algorithm 21
func HintBitUnpack(k, omega uint8, y []byte) (RingVector, error) {
	h := NewRingVector(k)
	index := byte(0)
	for i := range k {
		if y[omega+i] < index || y[omega+i] > omega {
			return nil, errors.New("malformed input: 201")
		}
		first := index
		for index < y[omega+i] {
			if index > first {
				if y[index-1] >= y[index] {
					return nil, errors.New("malformed input: 207")
				}
			}
			yidx := y[index]
			h[i][yidx] = 1
			index++
		}
	}
	for i := index; i < omega; i++ {
		if y[i] != 0 {
			return nil, errors.New("malformed input: trailing nonzero values")
		}
	}
	return h, nil
}

// Algorithm 22, parametrized by `k` from each ML-DSA parameter set
func PKEncode(k uint8, rho []byte, t1 RingVector) []byte {
	var pk []byte
	pk = append(pk, rho...)
	for i := range k {
		// 2^(bitlen(q - 1) - d) - 1 = 1023
		packed := SimpleBitPack(t1[i], 1023)
		pk = append(pk, packed...)
	}
	return pk
}

// Algorithm 23, parametrized by `k` from each ML-DSA parameter set
func PKDecode(k uint8, pk []byte) ([]byte, RingVector) {
	rho := pk[0:32]
	z := pk[32:]
	t1 := RingVector(make([]RingElement, k))
	j := 0
	for i := range k {
		// 2^(bitlen(q - 1) - d) - 1 = 1023
		t1[i] = SimpleBitUnpack(z[j:], 1023)
		j += 40
	}
	return rho, t1
}

// Algorithm 24
func SKEncode(k, l, eta uint8, rho, K, tr []byte, s1, s2, t0 RingVector) []byte {
	sk := rho[:]
	sk = append(sk, K...)
	sk = append(sk, tr...)
	for i := range l {
		packed := BitPack(s1[i], uint32(eta), uint32(eta))
		sk = append(sk, packed[:]...)
	}
	for i := range k {
		packed := BitPack(s2[i], uint32(eta), uint32(eta))
		sk = append(sk, packed[:]...)
	}

	max := uint32(1 << (d - 1))
	min := max - 1
	for i := range k {
		packed := BitPack(t0[i], min, max)
		sk = append(sk, packed[:]...)
	}
	return sk
}

// Algorithm 25
// Only run SKDecode() on values that come from trusted input
func SKDecode(k, l, eta uint8, sk []byte) ([]byte, []byte, []byte, RingVector, RingVector, RingVector) {
	rho, sk := sk[0:32], sk[32:]
	K, sk := sk[0:32], sk[32:]
	tr, sk := sk[0:64], sk[64:]
	y, sk := sk[0:l], sk[l:]
	z, w := sk[0:k], sk[k:]

	s1 := RingVector(make([]RingElement, l))
	s2 := RingVector(make([]RingElement, k))
	t0 := RingVector(make([]RingElement, k))
	for i := range l {
		s1[i] = BitUnpack(y[i:i], uint32(eta), uint32(eta))
	}
	for i := range k {
		s2[i] = BitUnpack(z[i:i], uint32(eta), uint32(eta))
	}
	for i := range k {
		t0[i] = BitUnpack(w[i:i], uint32((1<<12)-1), uint32(1<<12))
	}
	return rho, K, tr, s1, s2, t0
}

// Algorithm 26
func SigEncode(k, l, omega uint8, gamma1 uint32, c []byte, z, h RingVector) []byte {
	sigma := c[:]
	for i := range l {
		packed := BitPack(z[i], gamma1-1, gamma1)
		sigma = append(sigma, packed...)
	}
	packed := HintBitPack(k, omega, h)
	sigma = append(sigma, packed...)
	return sigma[:]
}

// Algorithm 27
func SigDecode(k, l, omega uint8, lambda uint16, gamma1 uint32, sig []byte) ([]byte, RingVector, RingVector, error) {
	sigma := slices.Clone(sig)
	z := NewRingVector(k)
	bitlen := bits.Len32(gamma1 - 1)

	length := uint32(lambda >> 2)
	c, sigma := sigma[0:length], sigma[length:]

	length = uint32((1 + bitlen) << 5)
	start := uint32(0)
	for i := range l {
		x := sigma[start:]
		start += length
		z[i] = BitUnpack(x[:], uint32(gamma1-1), uint32(gamma1))
	}
	y := sigma[start:]
	h, err := HintBitUnpack(k, omega, y)
	if err != nil {
		return nil, nil, nil, err
	}
	return c, z, h, nil
}

// Algorithm 28
func W1Encode(k uint8, gamma2 uint32, w1 RingVector) []byte {
	var w []byte
	div := (q-1)/(gamma2<<1) - 1
	// div: gamma2 == (q-1/88) -> 43
	// div: gamma2 == (q-1/32) -> 15
	for i := range k {
		packed := SimpleBitPack(w1[i], div)
		w = append(w, packed...)
	}
	return w
}

// Algorithm 29
func SampleInBall(tau uint8, seed []byte) (c RingElement) {
	ctx := sha3.NewShake256()
	ctx.Write(seed)
	s := make([]byte, 8)
	ctx.Read(s)
	h := BytesToBits(s) // 64 bits
	tau16 := int16(tau)

	// The NIST specification says "to 255" which means "< 256"
	// for i from (256 - tau) to 255 do
	for i := int16(256 - tau16); i < 256; i++ {
		j := make([]byte, 1)
		ctx.Read(j)
		for int16(j[0]) > i {
			ctx.Read(j)
		}
		j0 := int16(j[0])
		c[i] = c[j0]
		// Swap between 1 and -1 without side-channels
		diff := uint32(1) - uint32(h[i+tau16-256])<<1
		diff += -(diff >> 31) & q
		c[j0] = CoeffReduceOnce(diff) // c_j <- (-1)^h[i+tau-256]
	}
	return c
}

// Algorithm 30
func RejNTTPoly(seed []byte) (ah NttElement) {
	j := 0
	ctx := sha3.NewShake128()
	ctx.Write(seed)
	for j < 256 {
		s := make([]byte, 3)
		ctx.Read(s)
		tmp, err := CoeffFromThreeBytes(s[0], s[1], s[2])
		if err == nil {
			ah[j] = tmp
			j++
		}
	}
	return ah
}

// Algorithm 31
func RejBoundedPoly(eta int, seed []byte) RingElement {
	var a RingElement
	ctx := sha3.NewShake256()
	ctx.Write(seed)
	j := 0
	for j < 256 {
		z := make([]byte, 1)
		ctx.Read(z)
		z0, err := CoeffFromHalfByte(eta, z[0]&15)
		if err == nil {
			a[j] = z0
			j++
		}
		z1, err := CoeffFromHalfByte(eta, z[0]>>4)
		if err == nil {
			if j < 256 {
				a[j] = z1
				j++
			}
		}
	}
	return a
}

// Algorithm 32
func ExpandA(k, l uint8, rho []byte) NttMatrix {
	Ahat := NewNttMatrix(k, l)
	for r := range k {
		for s := range l {
			rhoprime := append(rho, byte(s), byte(r))
			Ahat[r][s] = RejNTTPoly(rhoprime)
		}
	}
	return Ahat
}

func PackUint16(x uint16) []byte {
	b := make([]byte, 2)
	b[0] = byte(x & 0xff)
	b[1] = byte(x >> 8)
	return b
}

// Algorithm 33
func ExpandS(k, l uint8, eta int, rho []byte) (RingVector, RingVector) {
	s1 := NewRingVector(l)
	s2 := NewRingVector(k)
	for r := range l {
		r_le := PackUint16(uint16(r))
		packed := append(rho, r_le...)
		s1[r] = RejBoundedPoly(eta, packed[:])
	}
	for r := range k {
		r_le := PackUint16(uint16(r) + uint16(l))
		packed := append(rho, r_le...)
		s2[r] = RejBoundedPoly(eta, packed[:])
	}
	return s1, s2
}

// H(str, l) -> SHAKE256(str, 8l)
func H(data []byte, length uint32) []byte {
	ctx := sha3.NewShake256()
	ctx.Write(data)
	out := make([]byte, length)
	ctx.Read(out)
	return out
}

// Algorithm 34
func ExpandMask(l uint8, gamma1 uint32, rho []byte, mu uint16) RingVector {
	y := NewRingVector(l)
	c := uint32(1 + bits.Len32(gamma1-1))
	for r := range l {
		// rho' <- rho || IntegerToBytes(mu + r, 2)
		as16 := PackUint16(uint16(r) + mu)
		packed := append(rho, as16...)
		// v <- H(rho', 32c)
		v := H(packed, c<<5)
		// y[r] = BitUnpack(v, gamma1 - 1, gamma1)
		y[r] = BitUnpack(v, uint32(gamma1-1), uint32(gamma1))
	}
	return y
}

// Algorithm 35
func Power2Round(r uint32) (uint32, uint32) {
	shift := uint32(1 << (d - 1))
	a1 := (r + shift) >> d
	a0 := (r - (a1 << d))
	// If we underflowed, let's mask out the relevant bits
	a0 += -(a0 >> 31) & q
	// a0 &= (1 << d) - 1
	return a1, a0
}

// Algorithm 36
func Decompose(gamma2 uint32, r uint32) (uint32, uint32) {
	return DecomposeVarTime(gamma2, r)
}

func DecomposeVarTime(gamma2 uint32, r uint32) (uint32, uint32) {
	m := gamma2 << 1
	r_plus := r % q
	r0 := ModPlusMinus(r, m)
	diff := r_plus - r0
	diff += q & -(diff >> 31)
	if diff == q-1 {
		return uint32(0), r0 - 1
	} else {
		r1 := diff / m
		return r1 % q, r0 % q
	}
}

// This is only really used for Decompose().
// x in [0, m/2) -> x
// x in [m/2, m) -> y - m/2
// TODO, make constant-time
func ModPlusMinus(x uint32, m uint32) uint32 {
	halfm := m >> 1
	y := x % m // Reduce x mod m
	if y > halfm {
		return q - m + y
	}
	return y
	/*
		// Check if y >= halfm
		// diff := ^((halfm - y) >> 31) & 1 // diff = 1 if y < halfm, 0 otherwise
		diff := (halfm - y) >> 31 // diff = 1 if y < halfm, 0 otherwise
		mask := -diff             // mask = 0 if y >= halfm, -1 otherwise

		// If y >= halfm, it needs to wrap around to negative values (q-m)
		y = (mask & (q - y)) ^ (^mask & y)
		return y
	*/
}

/*
func barretReduce(a uint64, x uint64, mod uint64, shift uint8) uint64 {
	quotient := (a * x) >> uint64(shift)
	return a - (quotient * mod)
}

// Scott's attempt to write this function in constant-time:
func DecomposeCT(gamma2 uint32, r int32) (int32, int32) {
	var invmod uint64
	var bmod uint64
	// Divide r1 by 2*gamma2 without division
	// r1 /= (gamma2 << 1)
	if gamma2 == 95232 {
		// 95232^(q-2) % q
		invmod = uint64(8380329)
		// 2**36 / (95232)
		bmod = uint64(360800)
	} else {
		// 261888^(q-2) % q
		invmod = uint64(8380385)
		// 2**36 / (2 * 261888)
		bmod = uint64(131200)
	}
	m := gamma2 << 1
	// rpos := r % q
	rpos := uint32(r - q)
	rpos += (rpos >> 31) * q

	// rpos is in range [0, q)
	// m is either (q-1)/88 or (q-1)/32
	// We need r0 := r % m
	// r0 := rpos % (gamma2 << 1)
	// tmp := (uint64(rpos) * bmod) >> 36
	// r0 := ModPlusMinus(rpos, m)
	r0 := uint32(barretReduce(uint64(rpos), bmod, uint64(m), uint8(36)))
	// r0 := uint32(uint64(rpos) - tmp*uint64(m))
	// r0 is between 0 and 2m, we need to conditionally subtract it
	sub := (m - (r0 + 1)) >> 31
	r0 -= sub * m

	fmt.Printf("r0 %% m == (%d %% %d) == %d\n", r, m, r0)

	// if rpos - r0 == q - 1 {
	diff := uint32((rpos - r0) ^ (q - 1))
	cmp := (diff - 1) >> 31
	mask := uint64(-cmp)
	// fmt.Printf("diff == rpos - r0 == (%d - %d) == %d, mask = %d, cmp = %d\n", rpos, r0, diff, mask, cmp)

	//	tmp = uint64(rpos) - uint64(r0)
	//	red := (tmp * bmod) >> 36
	//	r1 := ^mask & uint32(tmp-red*uint64(m))
	//	fmt.Printf("rpos - r0 === (%d - %d) === r1 (%d)\n", rpos, r0, r1)
	//
	// r1 should be 0 if cmp == 1
	// otherwise, r1 = (rpos - r0)/m
	r1tmp := uint64(rpos) - uint64(r0)
	r1 := ^mask & r1tmp
	fmt.Printf("cmp (%d) -> mask = %d, rpos - r0 = %d, r1 = %d\n", cmp, mask, r1tmp, r1)

	// Calculate (r1 / m) % q, using a multiplicative inverse mod q
	// invmod = 1 / m (mod q)
	// unreduced = r1 * invmod
	unreduced := uint64(r1) * invmod
	// reduced = unreduced % q
	// reduced := FieldReduce(unreduced)

	// TODO Replace mod with barrett reduction
	reduced := unreduced % q

	// If cmp was set earlier, we subtract 1 from r0
	r0 = r0 - cmp
	fmt.Printf("r1 = %d, unreduced = %d, reduced = %d, r0 = %d\n\n", r1, unreduced, reduced, r0)
	return int32(reduced), int32(r0)
}
*/

// Algorithm 37
func HighBits(gamma2 uint32, r uint32) uint32 {
	r1, _ := Decompose(gamma2, r)
	return r1
}

func HighBitsElement(gamma2 uint32, e RingElement) RingElement {
	x := NewRingElement()
	for i := range 256 {
		r1, _ := Decompose(gamma2, uint32(e[i]))
		x[i] = RingCoeff(r1)
	}
	return x
}

func HighBitsVec(k uint8, gamma2 uint32, r RingVector) RingVector {
	v := NewRingVector(k)
	for i := range k {
		v[i] = HighBitsElement(gamma2, r[i])
	}
	return v
}

// Algorithm 38
func LowBits(gamma2 uint32, r uint32) uint32 {
	_, r0 := Decompose(gamma2, r)
	return r0
}

func LowBitsElement(gamma2 uint32, e RingElement) RingElement {
	x := NewRingElement()
	for i := range 256 {
		_, r0 := Decompose(gamma2, uint32(e[i]))
		x[i] = CoeffReduceOnce(r0)
	}
	return x
}

func LowBitsVec(k uint8, gamma2 uint32, r RingVector) RingVector {
	v := NewRingVector(k)
	for i := range k {
		v[i] = LowBitsElement(gamma2, r[i])
	}
	return v
}

// Algorithm 39
func MakeHint(gamma2 uint32, z, r FieldElement) uint8 {
	r1 := HighBits(gamma2, uint32(r))
	v1 := HighBits(gamma2, uint32(r+z)%q)
	// return (r1 ^ v1) != 0
	// r1 == v1 -> return 0
	// r1 != v1 -> return 1
	return uint8(^((r1^v1)-1)>>31) & 1
}

func MakeHintRingElement(gamma2 uint32, z, r RingElement) []uint8 {
	hints := make([]uint8, 256)
	for j := range 256 {
		zj := FieldReduceOnce(uint32(z[j]))
		rj := FieldReduceOnce(uint32(r[j]))
		hints[j] = MakeHint(gamma2, zj, rj)
	}
	return hints
}

func MakeHintRingVec(k uint8, gamma2 uint32, z, r RingVector) [][]uint8 {
	hints := make([][]uint8, k)
	for i := range k {
		hints[i] = MakeHintRingElement(gamma2, z[i], r[i])
	}
	return hints
}

// This is used to sum up the number of 1's in a Hint
func CountOnesHint(k uint8, w [][]uint8) uint32 {
	ones := uint32(0)
	for i := range k {
		for j := range 256 {
			wij := uint32(w[i][j])
			ones += wij & 1
		}
	}
	return ones
}

// Algorithm 40
func UseHint(gamma2 uint32, h uint8, r FieldElement) FieldElement {
	// This is a constant value. We can make this a look-up table if we want to avoid the division.
	m := (q - 1) / (gamma2 << 1)
	q2 := uint32(q >> 1)
	r1, r0 := Decompose(gamma2, uint32(r))

	// We rewrote some conditional logic here to be constant-time
	// The variable time algorithm looks like this:
	//
	// if h == 1 {
	//   if r0 > 0 {
	//     return (r1 + 1) % m
	//   } else {
	//     return (r1 - 1) % m
	//   }
	// }
	// return r1

	// Given h (a bit, stored as a uint8), r0, and r1, we can compute an unreduced
	// field element by doing some bitwise operators. Consult the table below:
	//
	// | h | r0sign | adjust | unreduced |
	// |---|--------|--------|-----------|
	// | 1 |    0   |   +1   |   r1 + 1  |
	// | 1 |    1   |   -1   |   r1 - 1  |
	// | 0 |    0   |    0   |       r1  |
	// | 0 |    1   |    0   |       r1  |

	// r0sign is the sign bit of r0-1
	r0sign := uint32(q2-r0) >> 31

	// -h becomes -1 or 0, which is then used as a mask for bitwise AND
	mask := -uint32(h)
	// (1 - (r0sign << 1)) becomes 1 if r0sign == 0. It becomes -1 if r0sign == 1.
	// This works because 1 - 0 == 1, but 1 - 2 == -1, and (r0sign << 1) is either 0 or 2.
	adjust := mask & uint32(1-(r0sign<<1))

	// Now that we have an adjustment value in the range (-1, 0, +1), add it to r1.
	unreduced := r1 + adjust

	// The final step is to reduce the result mod m:
	x := uint32(unreduced)
	x += -(x >> 31) & m

	// We return the result as a fieldElement:
	return FieldElement(x)
}

func UseHintRingElement(gamma2 uint32, h []uint8, r RingElement) RingElement {
	r1 := NewRingElement()
	for j := range 256 {
		rj := FieldReduceOnce(uint32(r[j]))
		r1[j] = RingCoeff(UseHint(gamma2, h[j], rj))
	}
	return r1
}

func UseHintRingVector(k uint8, gamma2 uint32, h [][]uint8, rv RingVector) RingVector {
	rv1 := NewRingVector(k)
	for i := range k {
		rv1[i] = UseHintRingElement(gamma2, h[i], rv[i])
	}
	return rv1
}

// Precomputed; only [1..255] are used:
var zetas = [n]FieldElement{0, 4808194, 3765607, 3761513, 5178923, 5496691, 5234739, 5178987, 7778734, 3542485, 2682288, 2129892, 3764867, 7375178, 557458, 7159240, 5010068, 4317364, 2663378, 6705802, 4855975, 7946292, 676590, 7044481, 5152541, 1714295, 2453983, 1460718, 7737789, 4795319, 2815639, 2283733, 3602218, 3182878, 2740543, 4793971, 5269599, 2101410, 3704823, 1159875, 394148, 928749, 1095468, 4874037, 2071829, 4361428, 3241972, 2156050, 3415069, 1759347, 7562881, 4805951, 3756790, 6444618, 6663429, 4430364, 5483103, 3192354, 556856, 3870317, 2917338, 1853806, 3345963, 1858416, 3073009, 1277625, 5744944, 3852015, 4183372, 5157610, 5258977, 8106357, 2508980, 2028118, 1937570, 4564692, 2811291, 5396636, 7270901, 4158088, 1528066, 482649, 1148858, 5418153, 7814814, 169688, 2462444, 5046034, 4213992, 4892034, 1987814, 5183169, 1736313, 235407, 5130263, 3258457, 5801164, 1787943, 5989328, 6125690, 3482206, 4197502, 7080401, 6018354, 7062739, 2461387, 3035980, 621164, 3901472, 7153756, 2925816, 3374250, 1356448, 5604662, 2683270, 5601629, 4912752, 2312838, 7727142, 7921254, 348812, 8052569, 1011223, 6026202, 4561790, 6458164, 6143691, 1744507, 1753, 6444997, 5720892, 6924527, 2660408, 6600190, 8321269, 2772600, 1182243, 87208, 636927, 4415111, 4423672, 6084020, 5095502, 4663471, 8352605, 822541, 1009365, 5926272, 6400920, 1596822, 4423473, 4620952, 6695264, 4969849, 2678278, 4611469, 4829411, 635956, 8129971, 5925040, 4234153, 6607829, 2192938, 6653329, 2387513, 4768667, 8111961, 5199961, 3747250, 2296099, 1239911, 4541938, 3195676, 2642980, 1254190, 8368000, 2998219, 141835, 8291116, 2513018, 7025525, 613238, 7070156, 6161950, 7921677, 6458423, 4040196, 4908348, 2039144, 6500539, 7561656, 6201452, 6757063, 2105286, 6006015, 6346610, 586241, 7200804, 527981, 5637006, 6903432, 1994046, 2491325, 6987258, 507927, 7192532, 7655613, 6545891, 5346675, 8041997, 2647994, 3009748, 5767564, 4148469, 749577, 4357667, 3980599, 2569011, 6764887, 1723229, 1665318, 2028038, 1163598, 5011144, 3994671, 8368538, 7009900, 3020393, 3363542, 214880, 545376, 7609976, 3105558, 7277073, 508145, 7826699, 860144, 3430436, 140244, 6866265, 6195333, 3123762, 2358373, 6187330, 5365997, 6663603, 2926054, 7987710, 8077412, 3531229, 4405932, 4606686, 1900052, 7598542, 1054478, 7648983}

// Algorithm 41
func NTT(w RingElement) (wh NttElement) {
	// what[j] <- wj
	for j := range 256 {
		wh[j] = FieldElement(uint32(w[j]))
	}
	// m <- 0, len <- 128
	m := 0
	for len := 128; len >= 1; len >>= 1 {
		// start <- 0, while start < 256 do
		for start := 0; start < 256; start += len << 1 {
			// m <- m + 1
			m++
			// z <- zetas[m]
			z := zetas[m]
			// for j from start to start + len - 1 do
			for j := start; j < start+len; j++ {
				// t <- (z * what[j + len]) mod q
				t := FieldMul(z, wh[j+len])
				// what[j + len] <- (what[j] - t) mod q
				wh[j+len] = FieldSub(wh[j], t)
				// what[j] <- (what[j] + t) mod q
				wh[j] = FieldAdd(wh[j], t)
			}
		}
		// start <- start + 2 * len in the for loop above
	}
	// len <- floor(len / 2) in the for loop above
	return wh
}

// Algorithm 42
func InverseNTT(wh NttElement) RingElement {
	w := NewRingElement()
	wt := NewNttElement()
	for j := range 256 {
		wt[j] = wh[j]
	}
	m := 256
	for len := 1; len < 256; len <<= 1 {
		for start := 0; start < 256; start += len << 1 {
			m--
			z := q - zetas[m] // z <- -zetas[m]
			for j := start; j < start+len; j++ {
				t := wt[j]
				wt[j] = FieldAdd(t, wt[j+len])
				wt[j+len] = FieldSub(t, wt[j+len])
				wt[j+len] = FieldMul(z, wt[j+len])
			}
		}
	}
	f := FieldElement(8347681) // 256⁻¹ mod q
	for j := range 256 {
		w[j] = RingCoeff(FieldMul(wt[j], f))
	}
	return RingElement(w)
}

// Helper function for iterating over a vector of k RingElements
func NttVec(k uint8, r RingVector) NttVector {
	v := NewNttVector(k)
	for i := range k {
		v[i] = NTT(r[i])
	}
	return v
}

// Helper function for iterating over a vector of k NttElements
func InvNttVec(k uint8, w NttVector) RingVector {
	r := NewRingVector(k)
	for i := range k {
		r[i] = InverseNTT(w[i])
	}
	return r
}

// Algorithm 43
func BitRev8(m byte) byte {
	b := Uint32ToBits(uint32(m), 8)
	b_rev := make([]byte, 8)
	for i := range 8 {
		b_rev[i] = b[7-i]
	}
	return byte(BitsToInteger(b_rev, 8))
}

// Algorithm 44
func NTTAdd(a, b NttElement) (c NttElement) {
	for i := range c {
		c[i] = FieldAdd(a[i], b[i])
	}
	return c
}
func NTTSub(a, b NttElement) (c NttElement) {
	for i := range c {
		c[i] = FieldSub(a[i], b[i])
	}
	return c
}

// Algorithm 45
func NTTMul(a, b NttElement) (c NttElement) {
	for i := range c {
		c[i] = FieldMul(a[i], b[i])
	}
	return c
}

// Algorithm 46
func AddVectorNTT(l uint8, v, w NttVector) NttVector {
	u := NewNttVector(l)
	for i := range l {
		u[i] = NTTAdd(v[i], w[i])
	}
	return u
}

func SubVectorNTT(k uint8, a, b NttVector) NttVector {
	c := NewNttVector(k)
	for i := range k {
		c[i] = NTTSub(a[i], b[i])
	}
	return c
}

// Algorithm 47
func ScalarVectorNTT(l uint8, c_hat NttElement, v_hat NttVector) NttVector {
	w := NewNttVector(l)
	for i := range l {
		w[i] = NTTMul(c_hat, v_hat[i])
	}
	return w
}

// Algorithm 48
func MatrixVectorNTT(k, l uint8, M_hat NttMatrix, v_hat NttVector) NttVector {
	w := NewNttVector(k)
	for i := range k {
		for j := range l {
			w[i] = NTTAdd(w[i], NTTMul(M_hat[i][j], v_hat[j]))
		}
	}
	return w
}

// Calculate the infinity norm.
func InfinityNorm(x uint32) uint32 {
	q2 := uint32(q >> 1)
	// Get the sign bit of x - q/2, to see if we're dealing with a "negative" number.
	// We encode (-x) as (q - x) since they are congruent mod q. The only signed ints
	// we deal with in ML-DSA are in the range [-2^10, 2^10], which is less than q/2.
	x -= -((q - x) >> 31) & q
	x += -(x >> 31) & q
	// put x in [0, q)
	if x >= q2 {
		return q - x
	}
	return x
}
