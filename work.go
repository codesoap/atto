package atto

import (
	"context"
	"encoding/binary"

	"golang.org/x/crypto/blake2b"
)

const workerRoutines = 128

type workerResult struct {
	hashNumber uint64
	nonce      uint64
	err        error
}

func findNonce(workThreshold uint64, suffix []byte) (uint64, error) {
	// See https://docs.nano.org/integration-guides/work-generation/#work-equation
	// See https://docs.nano.org/protocol-design/spam-work-and-prioritization/#work-algorithm-details
	results := make(chan workerResult)
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < workerRoutines; i++ {
		go calculateHashes(suffix, uint64(i), results, ctx)
	}
	for {
		result := <-results
		if result.err != nil || result.hashNumber >= workThreshold {
			stopWorkers(results, cancel)
			return result.nonce, result.err
		}
	}
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

func calculateHashes(suffix []byte, nonce uint64, results chan workerResult, ctx context.Context) {
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
			results <- workerResult{
				hashNumber: binary.LittleEndian.Uint64(hashBytes),
				nonce:      nonce,
			}
			hasher.Reset()
			nonce += uint64(workerRoutines)
		}
	}
}
