package internal

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"slices"

	"golang.org/x/crypto/sha3"
	"github.com/trailofbits/ml-dsa/internal/params"
	"github.com/trailofbits/ml-dsa/internal/ring"
	"github.com/trailofbits/ml-dsa/internal/util"
)

type VerifyingKey struct {
	cfg *params.Cfg
	rho [32]byte  // Rho is the public seed
	t1  []ring.Rq // Length cfg.K
}

type SigningKey struct {
	cfg  *params.Cfg
	seed []byte   // ξ from the specification - nil if no seed is known
	rho  [32]byte // Rho is the public seed
	K    [32]byte
	tr   [64]byte
	s1   []ring.Rq // Length cfg.L  // TODO - move to Rq
	s2   []ring.Rq // Length cfg.K
	t0   []ring.Rq // Length cfg.K
	t1   []ring.Rq // Component of verifying key - cached for efficiency
}

// Serialize a public verifying key to bytes.
// Algorithm 22
func (vk *VerifyingKey) Bytes() []byte {
	pk := make([]byte, 0, vk.cfg.PkSize)
	pk = append(pk, vk.rho[:]...)
	for i := range vk.cfg.K {
		pk = append(pk, util.SimpleBitPack(vk.t1[i].Symmetric(), 10)...)
	}
	return pk
}

func PkDecode(cfg *params.Cfg, pk []byte) (*VerifyingKey, error) {
	if len(pk) != int(cfg.PkSize) {
		return nil, errors.New("invalid public key size")
	}
	rho := pk[0:32]
	z := pk[32:]

	t1 := make([]ring.Rq, cfg.K)
	elemLen := int(10*params.N) / 8
	for i := range cfg.K {
		t1[i] = ring.FromSymmetric(util.SimpleBitUnpack(z[:elemLen], 10))
		z = z[elemLen:]
	}

	res := new(VerifyingKey)
	res.cfg = cfg
	copy(res.rho[:], rho)
	res.t1 = t1

	return res, nil
}

// Algorithm 25
// We do not recommend using SkDecode. Users should prefer FromSeed instead.
// This implementation adds some extra validity checks beyond the FIPS-204 spec, however
// SkDecode should still only be run on inputs that come from trusted sources.
// This implementation is also much less efficient than the FIP-204, as it re-computes the
// full public key from the secret key material.
func SkDecode(cfg *params.Cfg, sk []byte) (*SigningKey, error) {
	var err error
	if len(sk) != int(cfg.SkSize) {
		return nil, errors.New("invalid secret key size")
	}
	expected := sk[:] // re-slice sk to allow round-trip verification

	l, k := cfg.L, cfg.K
	rho, sk := sk[0:32], sk[32:]
	K, sk := sk[0:32], sk[32:]
	_, sk = sk[0:64], sk[64:] // skip tr, we will re-compute it later

	s1 := make([]ring.Rz, l)
	s2 := make([]ring.Rz, k)

	// We don't need to parse t0, since we can compute it from s1 and s2.
	// t0 := make([]ring.Rz, k)

	var y []byte
	elemLen := int(32 * (cfg.LogEta + 2))
	for i := range l {
		y, sk = sk[0:elemLen], sk[elemLen:]
		s1[i], err = util.BitUnpackClosed(y, cfg.LogEta)
		if err != nil {
			return nil, err
		}
	}

	for i := range k {
		y, sk = sk[0:elemLen], sk[elemLen:]
		s2[i], err = util.BitUnpackClosed(y, cfg.LogEta)
		if err != nil {
			return nil, err
		}
	}

	/*
		elemLen = 32 * params.D
		for i := range k {
			y, sk = sk[0:elemLen], sk[elemLen:]
			t0[i] = util.BitUnpack(y, params.D-1)
		}
	*/

	res := new(SigningKey)
	res.cfg = cfg
	copy(res.rho[:], rho)
	copy(res.K[:], K)
	res.s1 = ring.FromSymmetricVec(s1)
	res.s2 = ring.FromSymmetricVec(s2)

	// This computes `t0` and `t1` from `s1` and `s2`
	// which means that we don't actually need to parse `t0` from the serialized key.
	err = res.computeT()
	if err != nil {
		return nil, err
	}

	// We do guarantee that `t0` and `tr` in the serialized key is correct, by
	// checking the round-trip serialization.
	enc := res.EncodeExpanded()
	if subtle.ConstantTimeCompare(enc, expected) != 1 {
		return nil, errors.New("invalid secret key")
	}

	return res, nil
}

// We do not recommend actually ever using this. Store the seed instead.
func (sk SigningKey) EncodeExpanded() []byte {
	encoded := append(sk.rho[:], sk.K[:]...)
	encoded = append(encoded, sk.tr[:]...)

	for i := range sk.cfg.L {
		packed := util.BitPackClosed(sk.s1[i].Symmetric(), sk.cfg.LogEta)
		encoded = append(encoded, packed[:]...)
	}
	for i := range sk.cfg.K {
		packed := util.BitPackClosed(sk.s2[i].Symmetric(), sk.cfg.LogEta)
		encoded = append(encoded, packed[:]...)
	}
	for i := range sk.cfg.K {
		packed := util.BitPack(sk.t0[i].Symmetric(), params.D-1)
		encoded = append(encoded, packed[:]...)
	}
	return encoded
}

// FromSeed creates a SigningKey from a 32-byte seed using
// Algorithm 6 of FIPS 204 (ML-DSA.KeyGen_internal)
func FromSeed(cfg *params.Cfg, seed []byte) (*SigningKey, error) {
	if len(seed) != 32 {
		return nil, errors.New("invalid seed length")
	}

	sk := new(SigningKey)
	sk.cfg = cfg
	sk.seed = make([]byte, 32)
	copy(sk.seed, seed)

	rhoPrime := make([]byte, 64)

	h := sha3.NewShake256()
	h.Write(seed[:])
	h.Write([]byte{cfg.K, cfg.L})

	_, err := h.Read(sk.rho[:])
	if err != nil {
		return sk, err
	}
	_, err = h.Read(rhoPrime[:])
	if err != nil {
		return sk, err
	}
	_, err = h.Read(sk.K[:])
	if err != nil {
		return sk, err
	}

	sk.s1, sk.s2 = util.ExpandS(cfg, rhoPrime[:])

	err = sk.computeT()
	if err != nil {
		return sk, err
	}

	return sk, nil
}

// GenerateKeyPair creates a new SigningKey and VerifyingKey pair using randomness from
// the provided io.Reader. The reader must be cryptographically secure.
// The keys are generated using a random seed of 32 bytes.
func GenerateKeyPair(cfg *params.Cfg, rng io.Reader) (*SigningKey, *VerifyingKey, error) {
	seed := make([]byte, 32)
	if rng == nil {
		rng = rand.Reader
	}
	n, err := io.ReadFull(rng, seed)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read random bytes: %w", err)
	}
	if n != 32 {
		// This should never happen, using ReadFull
		return nil, nil, fmt.Errorf("expected 32 bytes, got %d", n)
	}

	sk, err := FromSeed(cfg, seed)

	// This should never happen; only error case is if seed is not 32 bytes
	if err != nil {
		panic(err)
	}

	return sk, sk.Public(), nil
}

func (sk *SigningKey) Bytes() ([]byte, error) {
	if sk.seed == nil {
		return nil, errors.New("key was not generated from a seed; use EncodeExpanded instead")
	}
	return append([]byte(nil), sk.seed...), nil
}

func (sk *SigningKey) Public() *VerifyingKey {
	pk := new(VerifyingKey)
	pk.cfg = sk.cfg
	copy(pk.rho[:], sk.rho[:])
	pk.t1 = slices.Clone(sk.t1)
	return pk
}

// Fills in `t0, t1, tr` based on already-computed `rho, s0, s1`
func (sk *SigningKey) computeT() error {
	ahat := util.ExpandA(sk.cfg, sk.rho[:])
	s1hat := util.NttVec(sk.s1)
	tmp := util.InvNttVec(util.MatrixVectorNTT(ahat, s1hat))
	t1, t0 := util.Power2RoundVec(util.AddVector(tmp, sk.s2))

	sk.t0 = ring.FromSymmetricVec(t0)
	sk.t1 = ring.FromSymmetricVec(t1)

	h := sha3.NewShake256()
	h.Write(sk.Public().Bytes())
	_, err := h.Read(sk.tr[:])
	if err != nil {
		return err
	}

	return nil
}
