// Common options for ML-DSA signatures.
package options

import "crypto"

type Options struct {
	// Hash must be currently be zero, for pure ML-DSA.
	Hash crypto.Hash

	// Optional application-specific context string. At most 255 bytes.
	Context string
}

// Implements crypto.SignerOpts
func (o *Options) HashFunc() crypto.Hash {
	if o == nil {
		return 0
	}
	return o.Hash
}
