// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mldsa87_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/trailofbits/ml-dsa/mldsa87"
	"github.com/trailofbits/ml-dsa/options"
)

func Example() {
	pub, priv, err := mldsa87.GenerateKeyPair(nil)
	if err != nil {
		log.Fatal(err)
	}

	msg := []byte("Hello, world!")

	sig, err := priv.Sign(nil, msg, nil)
	if err != nil {
		log.Fatal(err)
	}

	ok := pub.Verify(msg, sig)
	fmt.Println(ok)
	// Output: true
}

func ExamplePublicKey_VerifyWithOptions() {
	pub, priv, err := mldsa87.GenerateKeyPair(nil)
	if err != nil {
		log.Fatal(err)
	}

	msg := []byte("Hello, world!")

	sig, err := priv.Sign(nil, msg, &options.Options{Context: "test"})
	if err != nil {
		log.Fatal(err)
	}

	ok := pub.VerifyWithOptions(msg, sig, &options.Options{Context: "test"})
	fmt.Println(ok)
	// Output: true
}

func FuzzSignAndVerify(f *testing.F) {
	f.Fuzz(func(t *testing.T, msg []byte) {
		pub, priv, err := mldsa87.GenerateKeyPair(nil)
		if err != nil {
			t.Skip()
		}

		sig, err := priv.Sign(nil, msg, nil)
		if err != nil {
			// We expect a failure here, since the message may be too
			// large to be signed, so we just return
			return
		}

		if !pub.Verify(msg, sig) {
			t.Errorf("Signature failed to verify")
		}
	})
}
