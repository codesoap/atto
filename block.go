package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"golang.org/x/crypto/blake2b"
)

var errInvalidSignature = fmt.Errorf("invalid block signature")

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
	HashBytes      []byte `json:"-"`
}

type workGenerateResponse struct {
	Error string `json:"error"`
	Work  string `json:"work"`
}

func (b *block) sign(a account) error {
	if err := b.addHashIfUnhashed(a); err != nil {
		return err
	}
	signature, err := sign(a, b.HashBytes)
	if err != nil {
		return err
	}
	b.Signature = fmt.Sprintf("%0128X", signature)
	return nil
}

func (b *block) verifySignature(a account) (err error) {
	if err = b.addHashIfUnhashed(a); err != nil {
		return
	}
	sig, ok := big.NewInt(0).SetString(b.Signature, 16)
	if !ok {
		return fmt.Errorf("cannot parse '%s' as an integer", b.Signature)
	}
	if !isValidSignature(a, b.HashBytes, bigIntToBytes(sig, 64)) {
		err = errInvalidSignature
	}
	return
}

func (b *block) addHashIfUnhashed(a account) error {
	if b.Hash == "" || len(b.HashBytes) == 0 {
		hashBytes, err := b.hash(a)
		if err != nil {
			return err
		}
		b.HashBytes = hashBytes
		b.Hash = fmt.Sprintf("%064X", b.HashBytes)
	}
	return nil
}

func (b *block) hash(a account) ([]byte, error) {
	// See https://nanoo.tools/block for a reference.

	msg := make([]byte, 176, 176)

	msg[31] = 0x6 // block preamble

	copy(msg[32:64], bigIntToBytes(a.publicKey, 32))

	previous, err := hex.DecodeString(b.Previous)
	if err != nil {
		return []byte{}, err
	}
	copy(msg[64:96], previous)

	representative, err := getPublicKeyFromAddress(b.Representative)
	if err != nil {
		return []byte{}, err
	}
	copy(msg[96:128], bigIntToBytes(representative, 32))

	balance, ok := big.NewInt(0).SetString(b.Balance, 10)
	if !ok {
		return []byte{}, fmt.Errorf("cannot parse '%s' as an integer", b.Balance)
	}
	copy(msg[128:144], bigIntToBytes(balance, 16))

	link, err := hex.DecodeString(b.Link)
	if err != nil {
		return []byte{}, err
	}
	copy(msg[144:176], link)

	hash := blake2b.Sum256(msg)
	return hash[:], nil
}

func (b *block) addWork(workThreshold uint64, a account) error {
	var hash string
	if b.Previous == "0000000000000000000000000000000000000000000000000000000000000000" {
		hash = fmt.Sprintf("%064X", bigIntToBytes(a.publicKey, 32))
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
	if err = json.Unmarshal(responseBytes, &response); err != nil {
		return err
	}
	// Need to check response.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if response.Error != "" {
		return fmt.Errorf("could not get work for block: %s", response.Error)
	}
	b.Work = response.Work
	return nil
}
