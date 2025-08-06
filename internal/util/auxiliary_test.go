package util

import (
	"testing"

	"crypto/rand"

	"github.com/stretchr/testify/assert"
	"trailofbits.com/ml-dsa/internal/params"
)

func TestSampleInBall(t *testing.T) {
	// Sample 32-byte slice
	buf := make([]byte, params.MLDSA44Cfg.Lambda/4)
	_, err := rand.Read(buf)
	if err != nil {
		t.Fatalf("Could not read from RNG")
	}

	sample := SampleInBall(params.MLDSA44Cfg, buf)

	cnt := 0
	for _, v := range sample {
		assert.True(t, v >= -1 && v <= 1)
		if v != 0 {
			cnt++
		}
	}

	// Sample should have low hamming weight
	assert.LessOrEqual(t, cnt, 64)
}
