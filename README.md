# Module-Lattice Digital Signature Algorithm

This repository implements [FIPS 204](https://nvlpubs.nist.gov/nistpubs/fips/nist.fips.204.pdf) in Go.

[![Build Status](https://github.com/trailofbits/go-ml-dsa/actions/workflows/ci.yml/badge.svg)](https://github.com/trailofbits/go-ml-dsa/actions/workflows/ci.yml)

## Installation

```terminal
go get https://github.com/trailfobits/go-ml-dsa
```

## Usage

```go
import(
	"log"
    mldsa44 "trailofbits.com/ml-dsa/mldsa44"
)

pub, priv, err := mldsa44.GenerateKeyPair(nil)
if err != nil {
    log.Fatal(err)
}

msg := []byte("Hello, world!")

sig, err := priv.Sign(nil, msg, nil)
if err != nil {
    log.Fatal(err)
}

ok := pub.Verify(msg, sig)
```
