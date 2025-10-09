// Copyright 2025 Trail of Bits. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build valgrind

package main

import (
	"bytes"
	"crypto/rand"
	"log"
	_ "runtime/valgrind" // Import for side-effects

	"github.com/trailofbits/ml-dsa/mldsa44"
)

// The following functions are not actually defined in the valgrind package,
// but are recognized as intrinsics by the compiler.

//go:noescape
func valgrind_make_mem_undefined(data []byte)

//go:noescape
func valgrind_make_mem_defined(data []byte)

func main() {
	// Generate a new key pair
	pub, priv, err := mldsa44.GenerateKeyPair(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	// The message to be signed
	message := []byte("The quick brown fox jumps over the lazy dog")

	// Get the private key seed
	seed, err := priv.Seed()
	if err != nil {
		log.Fatalf("Failed to get private key seed: %v", err)
	}

	// Poison the private key seed to mark it as secret
	valgrind_make_mem_undefined(seed)

	// Sign the message. If the Sign function is not constant-time,
	// Valgrind will detect a conditional jump or move that depends on
	// the poisoned (uninitialized) private key.
	sig, err := priv.Sign(rand.Reader, message, nil)
	if err != nil {
		log.Fatalf("Failed to sign message: %v", err)
	}

	// Un-poison the private key seed
	valgrind_make_mem_defined(seed)

	// Verify the signature to ensure the test is valid
	if !pub.Verify(message, sig) {
		log.Fatalf("Failed to verify signature")
	}

	// Test that an incorrect signature fails to verify.
	// This is to ensure that the Verify function is working correctly.
	if pub.Verify(message, bytes.ToUpper(sig)) {
		log.Fatalf("Verified incorrect signature")
	}

	log.Println("Valgrind test completed successfully")
}
