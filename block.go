package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"filippo.io/edwards25519"
	"golang.org/x/crypto/blake2b"
)

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
	Hash           string `json:"-"`
}

type workerResult struct {
	hashNumber uint64
	nonce      uint64
}

func (b *block) sign(privateKey *big.Int) error {
	// Look at https://nanoo.tools/block for a reference. This
	// implementation based on the one from github.com/iotaledger/iota.go.

	publicKey := derivePublicKey(privateKey)
	hash, err := b.hash(publicKey)
	if err != nil {
		return err
	}
	b.Hash = fmt.Sprintf("%064X", hash)
	signature := make([]byte, 64, 64)

	privateKeyBytes := make([]byte, 32, 32)
	privateKey.FillBytes(privateKeyBytes)
	h, err := blake2b.New512(nil)
	if err != nil {
		return err
	}
	h.Write(privateKeyBytes)

	var digest1, messageDigest, hramDigest [64]byte
	h.Sum(digest1[:0])

	s := new(edwards25519.Scalar).SetBytesWithClamping(digest1[:32])

	h.Reset()
	h.Write(digest1[32:])
	h.Write(hash)
	h.Sum(messageDigest[:0])

	rReduced := new(edwards25519.Scalar).SetUniformBytes(messageDigest[:])
	R := new(edwards25519.Point).ScalarBaseMult(rReduced)

	encodedR := R.Bytes()

	h.Reset()
	h.Write(encodedR[:])
	publicKeyBytes := make([]byte, 32, 32)
	publicKey.FillBytes(publicKeyBytes)
	h.Write(publicKeyBytes)
	h.Write(hash)
	h.Sum(hramDigest[:0])

	kReduced := new(edwards25519.Scalar).SetUniformBytes(hramDigest[:])
	S := new(edwards25519.Scalar).MultiplyAdd(kReduced, s, rReduced)

	copy(signature[:], encodedR[:])
	copy(signature[32:], S.Bytes())

	b.Signature = fmt.Sprintf("%0128X", signature)
	return nil
}

func (b *block) hash(publicKey *big.Int) ([]byte, error) {
	msg := make([]byte, 176, 176)

	msg[31] = 0x6 // block preamble

	publicKeyBytes := make([]byte, 32, 32)
	publicKey.FillBytes(publicKeyBytes)
	copy(msg[32:64], publicKeyBytes)

	previous, err := hex.DecodeString(b.Previous)
	if err != nil {
		return []byte{}, err
	}
	copy(msg[64:96], previous)

	representative, err := getPublicKeyFromAddress(b.Representative)
	representativeBytes := make([]byte, 32, 32)
	representative.FillBytes(representativeBytes)
	if err != nil {
		return []byte{}, err
	}
	copy(msg[96:128], representativeBytes)

	balance, ok := big.NewInt(0).SetString(b.Balance, 10)
	if !ok {
		return []byte{}, fmt.Errorf("cannot parse '%s' as an integer", b.Balance)
	}
	balanceBytes := make([]byte, 16, 16)
	balance.FillBytes(balanceBytes)
	copy(msg[128:144], balanceBytes)

	link, err := hex.DecodeString(b.Link)
	if err != nil {
		return []byte{}, err
	}
	copy(msg[144:176], link)

	hash := blake2b.Sum256(msg)
	return hash[:], nil
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
	errs := make(chan error, workerRoutines)
	for i := 0; i < workerRoutines; i++ {
		go calculateHashes(suffix, uint64(i), results, quit, errs)
	}
	for {
		select {
		case result := <-results:
			if result.hashNumber >= workThreshold {
				for i := 0; i < workerRoutines; i++ {
					quit <- true
				}
				return result.nonce, nil
			}
		case err := <-errs:
			for i := 0; i < workerRoutines; i++ {
				quit <- true
			}
			return 0, err
		}
	}
}

func calculateHashes(suffix []byte, nonce uint64, results chan workerResult, quit chan bool, errs chan error) {
	nonceBytes := make([]byte, 8)
	hasher, err := blake2b.New(8, nil)
	if err != nil {
		errs <- err
		return
	}
	for {
		select {
		case <-quit:
			return
		default:
			binary.LittleEndian.PutUint64(nonceBytes, nonce)
			_, err := hasher.Write(append(nonceBytes, suffix...))
			if err != nil {
				errs <- err
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
