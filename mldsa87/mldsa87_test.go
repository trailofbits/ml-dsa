// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mldsa87_test

import (
	"fmt"
	"log"

	"trailofbits.com/ml-dsa/mldsa87"
	"trailofbits.com/ml-dsa/options"
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
