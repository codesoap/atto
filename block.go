package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/blake2b"
)

type process struct {
	Action    string `json:"action"`
	JsonBlock string `json:"json_block"`
	Subtype   string `json:"subtype"`
	Block     block  `json:"block"`
}

type block struct {
	Type           string `json:"type"`
	Account        string `json:"account"`
	Previous       string `json:"previous"`
	Representative string `json:"representative"`
	Balance        string `json:"balance"`
	Link           string `json:"link"`
	LinkAsAccount  string `json:"link_as_account"`
	Signature      string `json:"signature"`
	Work           string `json:"work"`
}

type workerResult struct {
	hashNumber uint64
	nonce      uint64
}

func (b *block) sign(privateKey *big.Int) {
	// TODO
}

func (b *block) addWork(workThreshold uint64, privateKey *big.Int) (err error) {
	var suffix []byte
	if b.Previous == "0000000000000000000000000000000000000000000000000000000000000000" {
		publicKey := derivePublicKey(privateKey)
		suffix = make([]byte, 32, 32)
		publicKey.FillBytes(suffix)
	} else {
		suffix, err = hex.DecodeString(b.Previous)
		if err != nil {
			return
		}
	}
	nonce, err := findNonce(workThreshold, suffix)
	if err != nil {
		return
	}
	b.Work = fmt.Sprintf("%016x", nonce)
	return
}

func findNonce(workThreshold uint64, suffix []byte) (uint64, error) {
	results := make(chan workerResult)
	quit := make(chan bool, workerRoutines)
	for i := 0; i < workerRoutines; i++ {
		go calculateHashes(suffix, uint64(i), results, quit)
	}
	for {
		result := <-results
		if result.hashNumber >= workThreshold {
			for i := 0; i < workerRoutines; i++ {
				quit <- true
			}
			return result.nonce, nil
		}
	}
}

func calculateHashes(suffix []byte, nonce uint64, results chan workerResult, quit chan bool) {
	nonceBytes := make([]byte, 8)
	hasher, err := blake2b.New(8, nil)
	if err != nil {
		panic(err) // Kinda cheap, but will likely never happen.
	}
	for {
		select {
		case <-quit:
			return
		default:
			binary.BigEndian.PutUint64(nonceBytes, nonce)
			_, err := hasher.Write(append(nonceBytes, suffix...))
			if err != nil {
				panic(err) // Kinda cheap, but will likely never happen.
			}
			hashBytes := hasher.Sum(nil)
			results <- workerResult{
				hashNumber: binary.BigEndian.Uint64(hashBytes),
				nonce:      nonce,
			}
			hasher.Reset()
			nonce += uint64(workerRoutines)
		}
	}
}
