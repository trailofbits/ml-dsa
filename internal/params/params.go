// Package params specifies parameter sets for the ML-DSA signature scheme.

package params

const Q = 8380417
const N = 256
const Zeta = 1753
const D = 13

type Cfg struct {
	Name      string
	Tau       uint16
	Lambda    uint16
	LogGamma1 uint8
	Gamma2    uint32
	K         uint8
	L         uint8
	LogEta    uint8
	Beta      uint8
	Omega     uint8
	W1Bits    uint8  // Bit length of entries in W1 = bitlen((q-1)/(2*Gamma2) - 1)
	SkSize    uint16 // Byte size of expanded secret key - only used for known answer testing
	PkSize    uint16 // Byte size of encoded public key
}

var MLDSA44Cfg = &Cfg{
	Name:      "MLDSA-44",
	Tau:       39,
	Lambda:    128,
	LogGamma1: 17,
	Gamma2:    (Q - 1) / 88,
	K:         4,
	L:         4,
	LogEta:    1,
	Beta:      78,
	Omega:     80,
	W1Bits:    6, // bitlen(43)
	SkSize:    2560,
	PkSize:    1312,
}

var MLDSA65Cfg = &Cfg{
	Name:      "MLDSA-65",
	Tau:       49,
	Lambda:    192,
	LogGamma1: 19,
	Gamma2:    (Q - 1) / 32,
	K:         6,
	L:         5,
	LogEta:    2,
	Beta:      196,
	Omega:     55,
	W1Bits:    4, // bitlen(15)
	SkSize:    4032,
	PkSize:    1952,
}

var MLDSA87Cfg = &Cfg{
	Name:      "MLDSA-87",
	Tau:       60,
	Lambda:    256,
	LogGamma1: 19,
	Gamma2:    (Q - 1) / 32,
	K:         8,
	L:         7,
	LogEta:    1,
	Beta:      120,
	Omega:     75,
	W1Bits:    4, // bitlen(15)
	SkSize:    4896,
	PkSize:    2592,
}
