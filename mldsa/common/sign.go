package common

import (
	"crypto/subtle"
	"errors"
	"fmt"
)

func SignInternal(k, l, beta, tau, omega uint8, eta int, lambda uint16, gamma1 uint32, gamma2 uint32, t0 RingVector, seed, K, tr, Mprime, rnd []byte) ([]byte, error) {
	hashed := H(append(seed[:], byte(k), byte(l)), 128)
	rho := make([]byte, 32)
	rhoprime := make([]byte, 64)
	copy(rho, hashed[0:32])
	copy(rhoprime, hashed[32:96])
	s1, s2 := ExpandS(k, l, eta, rhoprime)

	s1hat := NttVec(l, s1)
	s2hat := NttVec(k, s2)
	t0hat := NttVec(k, t0[:])
	Ahat := ExpandA(k, l, rho)

	// mu <- H(BytesToBits(tr) || M', 64)
	mu := H(append(tr[:], Mprime...), 64)

	// rhopp <- H(K || rnd || mu, 64)
	tmp := append(K[:], rnd...)
	tmp = append(tmp, mu...)
	rhopp := H(tmp, 64)

	// Rejection sampling loop
	iterations := 0
	kappa := uint16(0)
	for {
		y := ExpandMask(l, gamma1, rhopp, kappa)
		w := InvNttVec(k, MatrixVectorNTT(k, l, Ahat, NttVec(k, y)))
		fmt.Printf("w1 = {\n")
		for i := range k {
			fmt.Printf("\t[%d] => [", i)
			for j := range 256 {
				fmt.Printf("%d, ", w[i][j])
			}
			fmt.Print("]\n")
		}
		fmt.Printf("}\n")
		w1 := HighBitsVec(k, gamma2, w)
		fmt.Printf("w1 = [\n")
		for i := range k {
			fmt.Printf("%d, ", w1[i])
		}
		fmt.Printf("]\n")
		w1_encoded := W1Encode(k, gamma2, w1)
		fmt.Printf("w1_tilde = [\n")
		for j := range 256 {
			fmt.Printf("%d, ", w1_encoded[j])
		}
		fmt.Printf("]\n")
		/*
			w1 = Vector(Array([Polynomial(Array([Elem(2), Elem(36), Elem(12), Elem(14), Elem(13), Elem(19), Elem(20), Elem(10), Elem(29), Elem(43), Elem(14), Elem(24), Elem(20), Elem(7), Elem(15), Elem(41), Elem(22), Elem(6), Elem(36), Elem(22), Elem(40), Elem(24), Elem(13), Elem(37), Elem(12), Elem(12), Elem(3), Elem(3), Elem(8), Elem(3), Elem(6), Elem(13), Elem(2), Elem(17), Elem(9), Elem(25), Elem(27), Elem(6), Elem(34), Elem(41), Elem(5), Elem(28), Elem(14), Elem(33), Elem(28), Elem(23), Elem(24), Elem(14), Elem(40), Elem(38), Elem(35), Elem(43), Elem(27), Elem(11), Elem(26), Elem(23), Elem(19), Elem(5), Elem(41), Elem(21), Elem(28), Elem(40), Elem(28), Elem(16), Elem(25), Elem(39), Elem(30), Elem(38), Elem(11), Elem(19), Elem(18), Elem(36), Elem(37), Elem(2), Elem(36), Elem(6), Elem(25), Elem(15), Elem(7), Elem(17), Elem(15), Elem(32), Elem(5), Elem(24), Elem(13), Elem(14), Elem(42), Elem(6), Elem(1), Elem(13), Elem(24), Elem(20), Elem(10), Elem(17), Elem(8), Elem(22), Elem(12), Elem(42), Elem(11), Elem(27), Elem(14), Elem(39), Elem(36), Elem(23), Elem(5), Elem(9), Elem(32), Elem(39), Elem(10), Elem(13), Elem(30), Elem(30), Elem(13), Elem(15), Elem(0), Elem(35), Elem(19), Elem(3), Elem(13), Elem(13), Elem(17), Elem(13), Elem(27), Elem(3), Elem(14), Elem(33), Elem(4), Elem(28), Elem(32), Elem(31), Elem(40), Elem(35), Elem(21), Elem(36), Elem(11), Elem(25), Elem(20), Elem(11), Elem(40), Elem(31), Elem(39), Elem(19), Elem(37), Elem(3), Elem(7), Elem(15), Elem(28), Elem(16), Elem(41), Elem(16), Elem(33), Elem(16), Elem(22), Elem(4), Elem(6), Elem(25), Elem(10), Elem(10), Elem(12), Elem(28), Elem(31), Elem(5), Elem(40), Elem(1), Elem(3), Elem(36), Elem(8), Elem(25), Elem(9), Elem(32), Elem(5), Elem(3), Elem(27), Elem(6), Elem(30), Elem(3), Elem(16), Elem(26), Elem(43), Elem(32), Elem(36), Elem(11), Elem(3), Elem(38), Elem(22), Elem(0), Elem(32), Elem(40), Elem(10), Elem(33), Elem(5), Elem(33), Elem(17), Elem(19), Elem(2), Elem(41), Elem(0), Elem(39), Elem(8), Elem(22), Elem(17), Elem(40), Elem(29), Elem(22), Elem(20), Elem(9), Elem(14), Elem(34), Elem(6), Elem(2), Elem(38), Elem(39), Elem(5), Elem(26), Elem(11), Elem(27), Elem(2), Elem(27), Elem(41), Elem(33), Elem(22), Elem(27), Elem(20), Elem(5), Elem(22), Elem(37), Elem(2), Elem(37), Elem(41), Elem(13), Elem(8), Elem(2), Elem(20), Elem(1), Elem(31), Elem(24), Elem(8), Elem(39), Elem(38), Elem(13), Elem(39), Elem(7), Elem(13), Elem(10), Elem(23), Elem(38), Elem(5), Elem(15), Elem(21), Elem(19), Elem(32), Elem(13), Elem(27), Elem(35), Elem(2), Elem(35)])), Polynomial(Array([Elem(0), Elem(40), Elem(8), Elem(11), Elem(34), Elem(8), Elem(38), Elem(6), Elem(43), Elem(28), Elem(7), Elem(21), Elem(36), Elem(34), Elem(1), Elem(26), Elem(33), Elem(22), Elem(11), Elem(0), Elem(39), Elem(1), Elem(31), Elem(28), Elem(13), Elem(5), Elem(38), Elem(43), Elem(31), Elem(42), Elem(38), Elem(36), Elem(22), Elem(12), Elem(5), Elem(15), Elem(42), Elem(32), Elem(14), Elem(4), Elem(28), Elem(20), Elem(27), Elem(11), Elem(38), Elem(37), Elem(25), Elem(10), Elem(28), Elem(8), Elem(6), Elem(21), Elem(42), Elem(30), Elem(25), Elem(1), Elem(41), Elem(12), Elem(1), Elem(13), Elem(43), Elem(39), Elem(34), Elem(15), Elem(37), Elem(11), Elem(5), Elem(2), Elem(33), Elem(1), Elem(24), Elem(6), Elem(17), Elem(38), Elem(35), Elem(2), Elem(4), Elem(5), Elem(5), Elem(9), Elem(17), Elem(22), Elem(43), Elem(32), Elem(5), Elem(38), Elem(24), Elem(6), Elem(41), Elem(24), Elem(42), Elem(18), Elem(3), Elem(20), Elem(15), Elem(6), Elem(3), Elem(16), Elem(17), Elem(41), Elem(13), Elem(10), Elem(8), Elem(2), Elem(32), Elem(23), Elem(34), Elem(33), Elem(37), Elem(37), Elem(31), Elem(21), Elem(26), Elem(8), Elem(16), Elem(12), Elem(14), Elem(11), Elem(40), Elem(42), Elem(2), Elem(5), Elem(15), Elem(1), Elem(26), Elem(37), Elem(30), Elem(15), Elem(41), Elem(22), Elem(37), Elem(26), Elem(15), Elem(25), Elem(35), Elem(13), Elem(0), Elem(15), Elem(3), Elem(43), Elem(2), Elem(18), Elem(7), Elem(33), Elem(3), Elem(31), Elem(36), Elem(10), Elem(32), Elem(17), Elem(11), Elem(30), Elem(15), Elem(8), Elem(11), Elem(14), Elem(36), Elem(31), Elem(15), Elem(9), Elem(18), Elem(42), Elem(1), Elem(38), Elem(37), Elem(26), Elem(17), Elem(39), Elem(12), Elem(39), Elem(12), Elem(26), Elem(8), Elem(37), Elem(14), Elem(37), Elem(26), Elem(37), Elem(16), Elem(21), Elem(24), Elem(24), Elem(41), Elem(6), Elem(32), Elem(15), Elem(21), Elem(15), Elem(3), Elem(26), Elem(37), Elem(7), Elem(11), Elem(5), Elem(8), Elem(30), Elem(6), Elem(37), Elem(30), Elem(20), Elem(40), Elem(7), Elem(22), Elem(34), Elem(42), Elem(38), Elem(35), Elem(20), Elem(33), Elem(9), Elem(36), Elem(3), Elem(7), Elem(15), Elem(43), Elem(27), Elem(16), Elem(1), Elem(13), Elem(14), Elem(30), Elem(27), Elem(14), Elem(36), Elem(21), Elem(36), Elem(39), Elem(31), Elem(21), Elem(41), Elem(1), Elem(3), Elem(20), Elem(8), Elem(31), Elem(36), Elem(29), Elem(36), Elem(37), Elem(26), Elem(20), Elem(15), Elem(40), Elem(31), Elem(41), Elem(1), Elem(35), Elem(2), Elem(22), Elem(8), Elem(5), Elem(20), Elem(20), Elem(25), Elem(26), Elem(22)])), Polynomial(Array([Elem(35), Elem(1), Elem(3), Elem(17), Elem(36), Elem(31), Elem(16), Elem(32), Elem(13), Elem(40), Elem(1), Elem(20), Elem(24), Elem(4), Elem(21), Elem(43), Elem(39), Elem(28), Elem(22), Elem(20), Elem(36), Elem(39), Elem(2), Elem(21), Elem(26), Elem(25), Elem(7), Elem(14), Elem(42), Elem(17), Elem(31), Elem(24), Elem(41), Elem(11), Elem(42), Elem(17), Elem(33), Elem(15), Elem(41), Elem(15), Elem(23), Elem(9), Elem(22), Elem(2), Elem(28), Elem(4), Elem(34), Elem(31), Elem(1), Elem(31), Elem(40), Elem(26), Elem(1), Elem(21), Elem(33), Elem(30), Elem(20), Elem(33), Elem(27), Elem(18), Elem(34), Elem(4), Elem(8), Elem(2), Elem(38), Elem(37), Elem(16), Elem(23), Elem(0), Elem(30), Elem(28), Elem(8), Elem(34), Elem(31), Elem(8), Elem(1), Elem(19), Elem(4), Elem(23), Elem(28), Elem(8), Elem(3), Elem(13), Elem(7), Elem(21), Elem(26), Elem(26), Elem(41), Elem(43), Elem(12), Elem(31), Elem(11), Elem(30), Elem(20), Elem(15), Elem(40), Elem(8), Elem(43), Elem(3), Elem(13), Elem(36), Elem(11), Elem(0), Elem(4), Elem(41), Elem(19), Elem(7), Elem(10), Elem(12), Elem(8), Elem(5), Elem(14), Elem(30), Elem(29), Elem(7), Elem(33), Elem(34), Elem(31), Elem(28), Elem(22), Elem(21), Elem(20), Elem(3), Elem(1), Elem(18), Elem(24), Elem(3), Elem(21), Elem(3), Elem(25), Elem(10), Elem(13), Elem(8), Elem(11), Elem(21), Elem(26), Elem(13), Elem(21), Elem(35), Elem(41), Elem(20), Elem(38), Elem(7), Elem(4), Elem(28), Elem(21), Elem(21), Elem(35), Elem(13), Elem(8), Elem(34), Elem(22), Elem(32), Elem(20), Elem(23), Elem(25), Elem(40), Elem(34), Elem(38), Elem(36), Elem(22), Elem(7), Elem(2), Elem(16), Elem(8), Elem(36), Elem(23), Elem(30), Elem(19), Elem(25), Elem(15), Elem(15), Elem(3), Elem(10), Elem(42), Elem(17), Elem(26), Elem(26), Elem(34), Elem(33), Elem(28), Elem(12), Elem(10), Elem(9), Elem(8), Elem(31), Elem(18), Elem(21), Elem(24), Elem(7), Elem(43), Elem(41), Elem(30), Elem(1), Elem(4), Elem(26), Elem(22), Elem(5), Elem(37), Elem(41), Elem(31), Elem(12), Elem(24), Elem(31), Elem(24), Elem(4), Elem(14), Elem(33), Elem(27), Elem(33), Elem(2), Elem(8), Elem(3), Elem(23), Elem(4), Elem(41), Elem(12), Elem(25), Elem(32), Elem(12), Elem(10), Elem(34), Elem(2), Elem(2), Elem(15), Elem(10), Elem(3), Elem(31), Elem(32), Elem(41), Elem(19), Elem(36), Elem(9), Elem(13), Elem(6), Elem(18), Elem(21), Elem(31), Elem(28), Elem(40), Elem(24), Elem(21), Elem(4), Elem(12), Elem(33), Elem(28), Elem(10), Elem(35), Elem(35), Elem(36), Elem(11), Elem(5), Elem(24), Elem(5), Elem(29), Elem(22)])), Polynomial(Array([Elem(19), Elem(39), Elem(19), Elem(0), Elem(24), Elem(6), Elem(18), Elem(31), Elem(12), Elem(21), Elem(41), Elem(15), Elem(18), Elem(13), Elem(1), Elem(17), Elem(9), Elem(41), Elem(31), Elem(25), Elem(26), Elem(3), Elem(18), Elem(30), Elem(43), Elem(20), Elem(33), Elem(4), Elem(5), Elem(40), Elem(9), Elem(24), Elem(37), Elem(17), Elem(29), Elem(26), Elem(24), Elem(15), Elem(5), Elem(35), Elem(41), Elem(28), Elem(40), Elem(2), Elem(15), Elem(18), Elem(31), Elem(37), Elem(20), Elem(31), Elem(31), Elem(37), Elem(25), Elem(17), Elem(3), Elem(19), Elem(34), Elem(33), Elem(22), Elem(28), Elem(43), Elem(43), Elem(28), Elem(7), Elem(24), Elem(24), Elem(29), Elem(23), Elem(21), Elem(9), Elem(37), Elem(37), Elem(0), Elem(33), Elem(13), Elem(3), Elem(1), Elem(35), Elem(43), Elem(6), Elem(20), Elem(32), Elem(20), Elem(27), Elem(26), Elem(7), Elem(40), Elem(2), Elem(29), Elem(4), Elem(3), Elem(16), Elem(8), Elem(3), Elem(37), Elem(39), Elem(29), Elem(2), Elem(41), Elem(12), Elem(24), Elem(41), Elem(25), Elem(30), Elem(28), Elem(1), Elem(17), Elem(14), Elem(14), Elem(26), Elem(28), Elem(13), Elem(39), Elem(34), Elem(20), Elem(5), Elem(10), Elem(28), Elem(11), Elem(32), Elem(8), Elem(24), Elem(23), Elem(4), Elem(23), Elem(6), Elem(17), Elem(37), Elem(4), Elem(9), Elem(0), Elem(23), Elem(11), Elem(40), Elem(36), Elem(36), Elem(11), Elem(35), Elem(29), Elem(20), Elem(0), Elem(37), Elem(34), Elem(25), Elem(24), Elem(34), Elem(32), Elem(5), Elem(38), Elem(31), Elem(43), Elem(18), Elem(18), Elem(6), Elem(39), Elem(9), Elem(23), Elem(6), Elem(28), Elem(33), Elem(2), Elem(6), Elem(13), Elem(24), Elem(14), Elem(6), Elem(11), Elem(40), Elem(40), Elem(40), Elem(4), Elem(37), Elem(18), Elem(13), Elem(22), Elem(9), Elem(19), Elem(19), Elem(31), Elem(36), Elem(33), Elem(19), Elem(40), Elem(40), Elem(26), Elem(26), Elem(33), Elem(32), Elem(22), Elem(27), Elem(28), Elem(36), Elem(28), Elem(1), Elem(26), Elem(17), Elem(36), Elem(40), Elem(0), Elem(13), Elem(29), Elem(11), Elem(37), Elem(25), Elem(21), Elem(31), Elem(1), Elem(26), Elem(34), Elem(37), Elem(39), Elem(39), Elem(42), Elem(8), Elem(39), Elem(37), Elem(24), Elem(36), Elem(27), Elem(10), Elem(11), Elem(0), Elem(8), Elem(11), Elem(16), Elem(31), Elem(0), Elem(34), Elem(20), Elem(21), Elem(4), Elem(15), Elem(11), Elem(17), Elem(12), Elem(24), Elem(41), Elem(4), Elem(35), Elem(12), Elem(29), Elem(19), Elem(25), Elem(43), Elem(14), Elem(36), Elem(14), Elem(38), Elem(2), Elem(28), Elem(25), Elem(16), Elem(23), Elem(37), Elem(35), Elem(8)]))]))
			w1_tilde = Array([2, 201, 56, 205, 68, 41, 221, 234, 96, 212, 241, 164, 150, 65, 90, 40, 214, 148, 12, 51, 12, 200, 96, 52, 66, 148, 100, 155, 33, 166, 5, 231, 132, 220, 133, 57, 168, 57, 174, 219, 162, 93, 83, 145, 86, 28, 202, 65, 217, 233, 153, 203, 36, 145, 165, 64, 26, 217, 115, 68, 15, 88, 96, 141, 163, 26, 65, 131, 81, 74, 132, 88, 140, 186, 108, 206, 73, 94, 69, 2, 158, 74, 227, 121, 205, 3, 140, 211, 208, 52, 81, 179, 13, 78, 72, 112, 224, 135, 142, 21, 185, 100, 212, 130, 126, 231, 84, 14, 199, 195, 65, 41, 20, 66, 22, 97, 100, 138, 194, 112, 95, 129, 6, 3, 137, 100, 9, 88, 12, 155, 225, 13, 144, 182, 130, 228, 50, 152, 22, 0, 162, 74, 88, 132, 209, 36, 164, 192, 137, 88, 17, 218, 89, 84, 226, 136, 134, 96, 158, 133, 182, 108, 194, 150, 134, 214, 70, 21, 86, 41, 148, 105, 131, 8, 84, 240, 97, 200, 105, 54, 231, 209, 40, 151, 89, 60, 213, 4, 54, 219, 40, 140, 0, 138, 44, 34, 98, 26, 43, 119, 84, 164, 24, 104, 161, 181, 0, 103, 240, 113, 77, 97, 174, 159, 106, 146, 22, 83, 60, 42, 232, 16, 28, 181, 45, 102, 153, 41, 28, 98, 84, 170, 151, 5, 41, 19, 52, 235, 41, 62, 229, 82, 8, 97, 128, 25, 145, 57, 10, 68, 81, 36, 145, 181, 130, 133, 137, 25, 41, 166, 74, 3, 245, 24, 3, 20, 165, 141, 130, 8, 224, 37, 134, 101, 249, 85, 26, 2, 49, 206, 130, 170, 66, 241, 4, 90, 233, 61, 169, 85, 106, 79, 54, 54, 192, 51, 172, 130, 116, 132, 195, 71, 42, 96, 180, 120, 15, 178, 56, 228, 247, 36, 146, 26, 152, 165, 22, 157, 204, 201, 104, 72, 233, 148, 90, 9, 85, 24, 150, 26, 224, 83, 61, 131, 86, 30, 75, 129, 120, 70, 233, 81, 232, 97, 137, 170, 57, 82, 97, 66, 14, 199, 179, 110, 80, 208, 56, 222, 230, 144, 21, 121, 126, 85, 26, 12, 20, 242, 145, 29, 89, 106, 212, 131, 126, 105, 48, 10, 22, 82, 80, 84, 166, 89, 99, 48, 68, 228, 7, 129, 13, 26, 80, 24, 81, 173, 39, 103, 81, 228, 41, 84, 90, 118, 56, 106, 244, 97, 233, 162, 70, 225, 147, 62, 87, 98, 9, 28, 33, 126, 193, 135, 106, 65, 21, 122, 84, 184, 73, 34, 129, 8, 102, 9, 93, 128, 199, 33, 226, 135, 4, 19, 113, 113, 200, 208, 28, 149, 166, 165, 43, 243, 45, 30, 245, 160, 200, 58, 52, 228, 2, 16, 233, 116, 40, 12, 82, 56, 94, 119, 132, 226, 199, 89, 21, 53, 4, 18, 54, 84, 67, 166, 52, 200, 82, 105, 77, 53, 166, 148, 121, 16, 92, 85, 141, 13, 34, 90, 32, 117, 101, 168, 104, 146, 214, 33, 64, 8, 121, 121, 83, 246, 60, 131, 162, 70, 154, 38, 134, 28, 163, 36, 200, 39, 85, 216, 177, 166, 94, 64, 104, 86, 81, 166, 31, 131, 125, 24, 225, 132, 91, 40, 32, 195, 69, 164, 76, 6, 50, 138, 40, 8, 143, 50, 124, 96, 58, 145, 73, 99, 72, 213, 199, 161, 88, 69, 48, 33, 167, 140, 35, 185, 20, 88, 209, 89, 211, 57, 1, 152, 33, 125, 76, 149, 62, 82, 19, 68, 73, 250, 101, 218, 32, 121, 43, 21, 18, 5, 154, 96, 101, 212, 105, 216, 83, 140, 41, 135, 10, 143, 244, 149, 212, 247, 149, 89, 52, 76, 98, 104, 113, 235, 202, 29, 24, 214, 93, 85, 82, 150, 64, 216, 12, 193, 184, 26, 20, 72, 109, 218, 129, 10, 29, 49, 64, 200, 80, 158, 157, 144, 50, 88, 154, 121, 92, 16, 57, 142, 198, 53, 167, 72, 21, 10, 183, 128, 8, 118, 17, 151, 17, 149, 68, 2, 92, 11, 74, 146, 203, 216, 81, 64, 41, 102, 152, 8, 22, 230, 183, 74, 146, 113, 38, 151, 193, 133, 130, 209, 96, 142, 177, 160, 40, 74, 148, 82, 99, 37, 211, 244, 145, 225, 132, 162, 154, 22, 130, 214, 198, 145, 92, 160, 69, 36, 10, 52, 221, 82, 102, 213, 23, 104, 98, 121, 158, 42, 114, 150, 24, 185, 41, 11, 128, 44, 208, 7, 136, 84, 69, 60, 75, 196, 96, 41, 49, 50, 221, 148, 173, 14, 233, 152, 2, 151, 65, 87, 57, 34])
		*/

		c_tilde := H(append(mu, w1_encoded...), uint32(lambda>>2))
		c := SampleInBall(tau, c_tilde)
		fmt.Printf("c = [\n")
		for j := range 256 {
			fmt.Printf("%d, ", c[j])
		}
		fmt.Printf("]\n\n")
		c_hat := NTT(c)
		/*
			c = Polynomial(Array([Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(0), Elem(0), Elem(1), Elem(0), Elem(8380416), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(1), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(8380416), Elem(1), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(8380416), Elem(0), Elem(0), Elem(0), Elem(0), Elem(1), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0), Elem(0)]))
		*/

		cs1 := InvNttVec(l, ScalarVectorNTT(l, c_hat, s1hat))
		cs2 := InvNttVec(k, ScalarVectorNTT(k, c_hat, s2hat))
		z := RingVectorAdd(l, y, cs1)
		for i := range l {
			/*
				fmt.Printf("s1[%d] = [", i)
				for j := range 256 {
					fmt.Printf("%d, ", s1hat[i][j])
				}
				fmt.Print("]\n")
			*/
			fmt.Printf("cs1[%d] = [", i)
			for j := range 256 {
				fmt.Printf("%d, ", cs1[i][j])
			}
			fmt.Print("]\n")
		}
		for i := range k {
			fmt.Printf("s2[%d] = [", i)
			for j := range 256 {
				fmt.Printf("%d, ", s2hat[i][j])
			}
			fmt.Print("]\n")
			fmt.Printf("cs2[%d] = [", i)
			for j := range 256 {
				fmt.Printf("%d, ", cs2[i][j])
			}
			fmt.Print("]\n")
		}
		for i := range k {
			fmt.Printf("z[%d] = [", i)
			for j := range 256 {
				fmt.Printf("%d, ", z[i][j])
			}
			fmt.Print("]\n")
		}
		r0 := LowBitsVec(k, gamma2, RingVectorSub(k, w, cs2))
		z_inf := InfinityNormRingVector(k, z)
		r0_inf := InfinityNormRingVector(k, r0)
		return nil, errors.New("Unfinished")

		// Validity checks
		gamma1_beta := gamma1 - uint32(beta)
		gamma2_beta := gamma2 - uint32(beta)
		if z_inf >= gamma1_beta || r0_inf >= gamma2_beta {
			fmt.Printf("z_inf (%d) >= gamma1 - beta (%d)", z_inf, gamma1_beta)
			fmt.Printf(" or r0_inf (%d) >= gamma2 - beta (%d)\n", r0_inf, gamma2_beta)
			z = nil
		}

		// <<ct0>> <- NTT^-1(c_hat o t0_hat)
		ct0 := InvNttVec(k, ScalarVectorNTT(l, c_hat, t0hat))
		minus_ct0 := NegateRingVector(k, ct0)
		w_cs2_ct0 := RingVectorSub(k, w, cs2)
		h := MakeHintRingVec(k, gamma2, minus_ct0, w_cs2_ct0)
		ct0_inf := InfinityNormRingVector(k, ct0)
		if ct0_inf >= gamma2 || CountOnesHint(k, h) > uint32(omega) {
			z = nil
			h = nil
		}
		kappa = kappa + uint16(l)

		// Loop termination via return
		if z != nil && h != nil {
			// Convert to the expected data type
			h_vec := NewRingVector(k)
			for i := range k {
				for j := range 256 {
					h_vec[i][j] = CoeffReduceOnce(uint32(h[i][j]))
				}
			}
			return SigEncode(k, l, omega, gamma1, c_tilde, z, h_vec), nil
		}

		// Appendix C - Loop Bounds
		iterations++
		if iterations > 814 {
			return nil, errors.New("too many rejections in common.SignInternal()")
		}
	}
}

func VerifyInternal(k, l, beta, tau, omega uint8, lambda uint16, gamma1, gamma2 uint32, rho []byte, t1 RingVector, Mprime, sigma []byte) bool {
	c_tilde, z, h, err := SigDecode(k, l, omega, lambda, gamma1, sigma)
	if err != nil {
		return false
	}
	if h == nil {
		return false
	}
	Ahat := ExpandA(k, l, rho)
	tr := H(PKEncode(k, rho, t1), 64)
	tr_bits := BytesToBits(tr)
	mu := H(append(tr_bits, Mprime...), 64)
	c := SampleInBall(tau, c_tilde)

	z_hat := NttVec(k, z)
	c_hat := NTT(c)
	t1_2d := NewNttVector(k)
	for i := range k {
		for j := range 256 {
			t1_2d[i][j] = FieldReduceOnce(uint32(t1[i][j] << d))
		}
	}
	ct1_2d_hat := ScalarVectorNTT(k, c_hat, t1_2d)
	Azhat := MatrixVectorNTT(k, l, Ahat, z_hat)
	// w_approx := InvNttVec(k, Azhat - ct1_2d_hat)
	w_approx := InvNttVec(k, SubVectorNTT(k, Azhat, ct1_2d_hat))

	// Convert back
	h8 := make([][]uint8, k)
	for i := range k {
		h8[i] = make([]uint8, 256)
		for j := range 256 {
			h8[i][j] = uint8(h[i][j])
		}
	}

	w1 := UseHintRingVector(k, gamma2, h8, w_approx)
	w1_encoded := W1Encode(k, gamma2, w1)
	c_tilde_prime := H(append(mu, w1_encoded...), uint32(lambda>>2))
	z_inf := InfinityNormRingVector(k, z)
	b32 := uint32(beta)
	return z_inf <= (gamma1-b32) && subtle.ConstantTimeCompare(c_tilde, c_tilde_prime) == 1
}
