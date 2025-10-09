# Module-Lattice Digital Signature Algorithm

This repository implements [FIPS 204](https://nvlpubs.nist.gov/nistpubs/fips/nist.fips.204.pdf) in Go.

[![Build Status](https://github.com/trailofbits/ml-dsa/actions/workflows/ci.yml/badge.svg)](https://github.com/trailofbits/ml-dsa/actions/workflows/ci.yml)
[![Fuzzing](https://github.com/trailofbits/ml-dsa/actions/workflows/fuzzing.yml/badge.svg)](https://github.com/trailofbits/ml-dsa/actions/workflows/fuzzing.yml)
[![Mutation Testing](https://github.com/trailofbits/ml-dsa/actions/workflows/mutation.yml/badge.svg)](https://github.com/trailofbits/ml-dsa/actions/workflows/mutation.yml)

## Installation

```terminal
go get https://github.com/trailofbits/ml-dsa
```

## Usage

```go
import(
	"log"
    mldsa65 "github.com/trailofbits/ml-dsa/mldsa65"
)

pub, priv, err := mldsa65.GenerateKeyPair(nil)
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

## Testing

This project includes fuzzing and mutation testing to ensure the quality and
robustness of the implementation.

### Fuzzing

To run the fuzz tests, use the following commands:

```terminal
go test -fuzz=FuzzSignAndVerify -fuzztime 60s ./mldsa44
# Alternatively:
go test -fuzz=FuzzSignAndVerify -fuzztime 60s ./mldsa65
go test -fuzz=FuzzSignAndVerify -fuzztime 60s ./mldsa87
```

This will run the fuzz tests for 60 seconds.

### Mutation Testing

To run the mutation tests, you'll first need to install `go-gremlins`:

```terminal
go install github.com/go-gremlins/gremlins@latest
```

Then, run the following command:

```terminal
gremlins -v ./...
```
