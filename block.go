package main

import (
	"encoding/hex"
	"encoding/json"
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
	Signature      string `json:"signature"`
	Work           string `json:"work"`
	Hash           string `json:"-"`
}

type workGenerateResponse struct {
	Error string `json:"error"`
	Work  string `json:"work"`
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

func (b *block) addWork(workThreshold uint64, privateKey *big.Int) error {
	var hash string
	if b.Previous == "0000000000000000000000000000000000000000000000000000000000000000" {
		publicKey := derivePublicKey(privateKey)
		publicKeyBytes := make([]byte, 32, 32)
		publicKey.FillBytes(publicKeyBytes)
		hash = fmt.Sprintf("%064X", publicKeyBytes)
	} else {
		hash = b.Previous
	}
	requestBody := fmt.Sprintf(`{`+
		`"action": "work_generate",`+
		`"hash": "%s",`+
		`"difficulty": "%016x"`+
		`}`, string(hash), workThreshold)
	responseBytes, err := doRPC(requestBody)
	if err != nil {
		return err
	}
	var response workGenerateResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return err
	}
	// Need to check pending.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if response.Error != "" {
		return fmt.Errorf("could not get work for block: %s", response.Error)
	}
	b.Work = response.Work
	return nil
}
