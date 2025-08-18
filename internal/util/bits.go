package util

import (
	"crypto/subtle"
	"errors"

	"github.com/trailofbits/ml-dsa/internal/params"
	"github.com/trailofbits/ml-dsa/internal/ring"
)

// Algorithm 9
// Returns a length-`a` []byte with a distinct byte entry for each bit, in lsb order
func IntegerToBits[T ~uint32](x T, a int) []byte {
	var y = make([]byte, a)
	for i := range a {
		y[i] = byte(x & 1)
		x = x >> 1
	}
	return y
}

// bitPack takes a slice of k-bit unsigned integers and packs them into a byte slice in lsb order.
// If entries are signed, they are assumed to be nonnegative
func bitPack[T int32 | uint32](w []T, k uint8) []byte {
	n := len(w)
	numBytes := (n*int(k) + 7) / 8
	z := make([]byte, numBytes)

	// l is the number of unoccupied bits remaining in the current byte
	l := uint8(8)
	// m is the index of the current byte in z
	m := 0

	for i := 0; i < n; i++ {
		v := uint32(w[i])
		j := k // number of bits left to store in the current value

		for l <= j {
			z[m] |= byte(v << (8 - l)) // bytes fill from the lsb to the msb
			v >>= l
			j -= l
			m += 1
			l = 8
		}

		if j == 0 {
			continue
		}

		// Store the remaining bits in the current byte
		z[m] |= byte(v << (8 - l))
		l -= j
	}
	return z
}

// Algorithm 16
// Assumes that all coefficients are in the range 0 <= x < 2^k
func SimpleBitPack(w ring.Rz, k uint8) []byte {
	return bitPack(w[:], k)
}

// Algorithm 17, specialized to values in the closed interval -2^k <= x <= 2^k
// 2^k is called eta in the context of FIPS 204
func BitPackClosed(w ring.Rz, k uint8) []byte {
	z := make([]uint32, len(w))
	for i := range w {
		z[i] = uint32((1 << k) - w[i])
	}
	return bitPack(z, k+2) // max value is 2^(k+1), which requires k+2 bits
}

// Algorithm 17, for open intervals -2^k < x <= 2^k
func BitPack(w ring.Rz, k uint8) []byte {
	z := make([]uint32, len(w))
	for i := range w {
		z[i] = uint32((1 << k) - w[i])
	}
	return bitPack(z, uint8(k+1)) // max value is 2^(k+1) - 1, which requires k+1 bits
}

// bitUnpack takes a byte slice and unpacks it into a slice of k-bit unsigned integers.
// The byte slice is assumed to be in lsb order.
func bitUnpack(z []byte, k uint8) []int32 {
	// Every use case packs or unpacks full ring elements
	w := make([]int32, params.N)

	// l is the number of bits available to be read from the current byte
	l := uint8(8)
	// m is the index of the current byte in z
	m := 0

	for i := 0; i < params.N; i++ {
		v := uint32(0)
		j := k // number of bits left to store in the current value

		for l <= j {
			v |= (uint32(z[m]) >> (8 - l)) << (k - j) // no need to mask since we are using all remaining bits in the byte
			j -= l
			m += 1
			l = 8
		}

		// Read the remaining bits from the current byte. Must mask as we may not be using all high bits in the byte.
		if j != 0 {
			v |= uint32(((z[m] >> (8 - l)) & ((1 << j) - 1))) << (k - j)
		}

		w[i] = int32(v)
		l -= j
	}
	return w
}

// Algorithm 18
func SimpleBitUnpack(b []byte, k uint8) (z ring.Rz) {
	w := bitUnpack(b, k)
	copy(z[:], w)
	return z
}

// Algorithm 19, for open intervals -2^k < x <= 2^k
func BitUnpack(b []byte, k uint8) (z ring.Rz) {
	w := bitUnpack(b, k+1)
	for i := range w {
		z[i] = (1 << k) - w[i]
	}
	return z
}

// Algorithm 19, specialized to values in -eta <= x <= eta
// Returns an error if any value is out of range
// 2^k is called eta in the context of FIPS 204
// k is always either 1 or 2.
// This is only used during sk decoding
func BitUnpackClosed(b []byte, k uint8) (z ring.Rz, err error) {
	w := bitUnpack(b, k+2)
	ok := 1
	for i := range w {
		// Malformed values can fall in the range 2^(k+1) < x < 2^(k+2)
		ok &= subtle.ConstantTimeLessOrEq(int(w[i]), int(2<<k))
		z[i] = (1 << k) - w[i]
	}

	// ok to be non-constant-time here, since it won't leak non-negligible information about the secret key
	if ok == 0 {
		return z, errors.New("malformed input") // TODO - real error type
	}
	return z, nil
}

// Algorithm 20
// Does not need to be constant-time, as hints are public
// This is used during signature encoding
func HintBitPack(k, omega uint8, h []ring.R2) []byte {
	y := make([]byte, k+omega)
	index := uint8(0)
	for i := range k {
		for j := range 256 {
			if h[i][j] == 1 {
				y[index] = byte(j)
				index = index + 1
			}
		}
		y[omega+i] = byte(index)
	}
	return y
}

// Algorithm 21
// This is used by signature verification, which does not need to be constant-time
func HintBitUnpack(k, omega uint8, y []byte) ([]ring.R2, error) {
	h := make([]ring.R2, k)
	index := byte(0)
	for i := range k {
		if y[omega+i] < index || y[omega+i] > omega {
			return nil, errors.New("malformed input")
		}
		first := index
		for index < y[omega+i] {
			if index > first {
				if y[index-1] >= y[index] {
					return nil, errors.New("malformed input")
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
