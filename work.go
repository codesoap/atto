package atto

import (
	"context"
	"encoding/binary"

	"github.com/klauspost/cpuid/v2"
	"golang.org/x/crypto/blake2b"
)

// Using cpuid.CPU.LogicalCores seems to yield the best performance.
var workerRoutines = cpuid.CPU.LogicalCores

type workerResult struct {
	nonce uint64
	err   error
}

func findNonce(workThreshold uint64, suffix []byte) (uint64, error) {
	// See https://docs.nano.org/integration-guides/work-generation/#work-equation
	// See https://docs.nano.org/protocol-design/spam-work-and-prioritization/#work-algorithm-details
	results := make(chan workerResult)
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < workerRoutines; i++ {
		go calculateHashes(workThreshold, suffix, uint64(i), results, ctx)
	}
	result := <-results
	stopWorkers(results, cancel)
	return result.nonce, result.err
}

func stopWorkers(results chan workerResult, cancel context.CancelFunc) {
	cancel()

	// Empty channel, so that workers get ready to quit:
	for {
		select {
		case <-results:
		default:
			return
		}
	}
}

func calculateHashes(workThreshold uint64, suffix []byte, nonce uint64, results chan workerResult, ctx context.Context) {
	nonceBytes := make([]byte, 8)
	hasher, err := blake2b.New(8, nil)
	if err != nil {
		results <- workerResult{err: err}
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			binary.LittleEndian.PutUint64(nonceBytes, nonce)
			_, err := hasher.Write(append(nonceBytes, suffix...))
			if err != nil {
				results <- workerResult{err: err}
				return
			}
			hashBytes := hasher.Sum(nil)
			hashNumber := binary.LittleEndian.Uint64(hashBytes)
			if hashNumber >= workThreshold {
				results <- workerResult{nonce: nonce}
			}
			hasher.Reset()
			nonce += uint64(workerRoutines)
		}
	}
}
