package internal

import (
	"crypto"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"io"

	"trailofbits.com/ml-dsa/internal/field"
	"trailofbits.com/ml-dsa/internal/params"
	"trailofbits.com/ml-dsa/internal/ring"
	"trailofbits.com/ml-dsa/internal/util"
	options "trailofbits.com/ml-dsa/options"
)

// Algorithm 7
//
// The message representative being signed is:
// Mprime
//
// Additional randomness is passed as:
// rnd
//
// We do not currently support the use fo an "external mu"
//
// Returns a signature as a []byte
func (sk *SigningKey) SignInternal(Mprime, rnd []byte) []byte {
	cfg := sk.cfg
	s1hat := util.NttVec(sk.s1) // TODO - consider caching s1hat, s2hat, t0hat, Ahat
	s2hat := util.NttVec(sk.s2)
	t0hat := util.NttVec(sk.t0)
	Ahat := util.ExpandA(sk.cfg, sk.rho[:])

	// mu <- H(BytesToBits(tr) || M', 64)
	mu := make([]byte, 64)
	util.H(mu, append(sk.tr[:], Mprime...))

	// rhopp <- H(K || rnd || mu, 64)
	rhopp := make([]byte, 64)
	tmp := append(sk.K[:], rnd...)
	tmp = append(tmp, mu...)
	util.H(rhopp, tmp)

	// Rejection sampling loop
	// We do not use loop bounds:
	// "Implementations *should* not bound the number of iterations in these loops..." (FIPS 204, Appendix C)
	for kappa := uint16(0); ; kappa += uint16(cfg.L) {
		y := ring.FromSymmetricVec(util.ExpandMask(cfg, rhopp, kappa))
		w := util.InvNttVec(util.MatrixVectorNTT(Ahat, util.NttVec(y)))
		w1 := ring.HighBitsVec(w, cfg.Gamma2) // TODO - more consistent API for cfg
		w1_encoded := util.W1Encode(cfg, w1)
		c_tilde := make([]byte, cfg.Lambda>>2)
		util.H(c_tilde, append(mu, w1_encoded...))
		c := util.SampleInBall(cfg, c_tilde)
		c_hat := util.NTT(ring.FromSymmetric(c))

		cs1 := util.InvNttVec(util.ScalarVectorNTT(c_hat, s1hat))
		cs2 := util.InvNttVec(util.ScalarVectorNTT(c_hat, s2hat))
		z := util.AddVector(y, cs1)
		r0 := ring.LowBitsVec(util.SubVector(w, cs2), cfg.Gamma2)
		z_inf := ring.InfinityNormVec(z)
		r0_inf := ring.InfinityNormVec(ring.FromSymmetricVec(r0))

		// Rejection sampling
		gamma1_beta := (1 << cfg.LogGamma1) - uint32(cfg.Beta)
		gamma2_beta := cfg.Gamma2 - uint32(cfg.Beta)

		if z_inf >= gamma1_beta || r0_inf >= gamma2_beta {
			continue
		}

		// <<ct0>> <- NTT^-1(c_hat o t0_hat)
		ct0 := util.InvNttVec(util.ScalarVectorNTT(c_hat, t0hat))
		minus_ct0 := util.NegateVector(ct0)
		// w - cs2
		// w_cs2 := RingVectorSub(k, w, cs2)
		// w - cs2 + ct0
		// w_cs2_ct0 := RingVectorAdd(k, w_cs2, ct0)
		w_cs2_ct0 := util.AddVector(util.SubVector(w, cs2), ct0)
		ct0_inf := ring.InfinityNormVec(ct0)
		// Returns `nil` if hint hamming weight is too large
		h := util.MakeHint(cfg, minus_ct0, w_cs2_ct0)
		if ct0_inf >= cfg.Gamma2 || h == nil {
			continue
		}

		return util.SigEncode(cfg, c_tilde, z, h)
	}
}

// Sign takes a message and a context and returns a signature.
// Only pure ML-DSA is supported.
// Context must be less than 256 bytes long, or else this function will return an error.
func (sk *SigningKey) Sign(rng io.Reader, message []byte, opts crypto.SignerOpts) ([]byte, error) {
	var h crypto.Hash
	ctx := []byte{}

	if opts != nil {
		h = opts.HashFunc()
		ops, ok := opts.(*options.Options)
		if ok {
			ctx = []byte(ops.Context)
		}
	}

	if h != 0 {
		return nil, errors.New("opts.HashFunc() must be zero for pure ML-DSA")
	}

	if len(ctx) > 255 {
		return nil, errors.New("context must be less than 256 bytes long")
	}

	rnd := make([]byte, 32)
	if rng == nil {
		rng = rand.Reader
	}
	n, err := rng.Read(rnd)
	if err != nil {
		return nil, err
	}
	if n != len(rnd) {
		return nil, errors.New("rng.Read() returned too few bytes")
	}

	Mprime := make([]byte, 0, len(ctx)+len(message)+2)
	Mprime = append(Mprime, byte(0), byte(len(ctx)))
	Mprime = append(Mprime, ctx...)
	Mprime = append(Mprime, message...)

	sigma := sk.SignInternal(Mprime, rnd)
	return sigma, nil
}

// Algorithm 8
//
// The message representitve for the signature is:
// Mprime
//
// The signature being validated is:
// sigma
//
// Returns true if the signature is valid.
// Returns false otherwise (even if an error occurs).
func (vk *VerifyingKey) VerifyInternal(Mprime, sigma []byte) bool {
	cfg := vk.cfg
	c_tilde, z, h, err := util.SigDecode(cfg, sigma)
	if err != nil {
		return false
	}

	Ahat := util.ExpandA(cfg, vk.rho[:])
	tr := make([]byte, 64)
	util.H(tr, vk.Bytes())
	mu := make([]byte, 64)
	util.H(mu, append(tr, Mprime...))

	c := util.SampleInBall(cfg, c_tilde)

	z_hat := util.NttVec(ring.FromSymmetricVec(z))
	c_hat := util.NTT(ring.FromSymmetric(c))

	t1_2d := util.ScalarVector(field.NewFromReduced(1<<params.D), vk.t1)

	t1_2d_hat := util.NttVec(t1_2d)
	ct1_2d_hat := util.ScalarVectorNTT(c_hat, t1_2d_hat)
	Azhat := util.MatrixVectorNTT(Ahat, z_hat)
	// w_approx := InvNttVec(k, Azhat - ct1_2d_hat)
	w_approx := util.InvNttVec(util.SubVectorNTT(Azhat, ct1_2d_hat))

	w1 := util.UseHint(cfg, h, w_approx)
	w1_encoded := util.W1Encode(cfg, w1)
	// TODO - change this API back, IDK why I did this
	c_tilde_prime := make([]byte, cfg.Lambda>>2)
	util.H(c_tilde_prime, append(mu, w1_encoded...))

	// Seems better to just do this directly on the Rz vec..
	z_inf := ring.InfinityNormVec(ring.FromSymmetricVec(z))

	bound := (1 << cfg.LogGamma1) - uint32(cfg.Beta)
	return z_inf <= bound && subtle.ConstantTimeCompare(c_tilde, c_tilde_prime) == 1
}

// Verify verifies a signature.
//
// Only pure ML-DSA is supported. opts.HashFunc() must return 0.
//
// opts may be nil, in which case empty context is used.
func (vk *VerifyingKey) Verify(msg, sig []byte, opts *options.Options) bool {
	ctx := []byte{}
	if opts != nil {
		if opts.HashFunc() != 0 {
			return false
		}
		ctx = []byte(opts.Context)
	}

	if len(ctx) > 255 {
		return false
	}

	Mprime := make([]byte, 0, len(ctx)+len(msg)+2)
	Mprime = append(Mprime, byte(0), byte(len(ctx)))
	Mprime = append(Mprime, ctx...)
	Mprime = append(Mprime, msg...)

	return vk.VerifyInternal(Mprime, sig)
}
