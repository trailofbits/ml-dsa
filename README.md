# Module-Lattice Digital Signature Algorithm (ML-DSA, FIPS 204)

This repository implements [FIPS 204](https://nvlpubs.nist.gov/nistpubs/fips/nist.fips.204.pdf) in Go.

## Installation

```terminal
go get https://github.com/trailfobits/go-ml-dsa
```

## Usage

```go
pk, sk, err := mldsa44.GenerateKeyPair(rng)
signature, err := sk.Sign(msg, ctx, rng)
if pk.Verify(msg, ctx, pk) {
    // valid
}
```
