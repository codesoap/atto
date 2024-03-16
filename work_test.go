package atto

import (
	"testing"
)

func BenchmarkNonceSearch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		findNonce(0xffffff0000000000, nil)
	}
}
