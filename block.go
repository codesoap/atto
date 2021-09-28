package atto

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"golang.org/x/crypto/blake2b"
)

var errInvalidSignature = fmt.Errorf("invalid block signature")

// ErrSignatureMissing is used when the Signature of a Block is missing
// but required for the attempted operation.
var ErrSignatureMissing = fmt.Errorf("signature is missing")

// ErrWorkMissing is used when the Work of a Block is missing but
// required for the attempted operation.
var ErrWorkMissing = fmt.Errorf("work is missing")

// Block represents a block in the block chain of an account.
type Block struct {
	Type           string `json:"type"`
	SubType        string `json:"-"`
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

// Sign computes and sets the Signature of b.
func (b *Block) Sign(publicKey, privateKey *big.Int) error {
	signature, err := sign(publicKey, privateKey, b.HashBytes)
	if err != nil {
		return err
	}
	b.Signature = fmt.Sprintf("%0128X", signature)
	return nil
}

func (b *Block) verifySignature(a Account) (err error) {
	sig, ok := big.NewInt(0).SetString(b.Signature, 16)
	if !ok {
		return fmt.Errorf("cannot parse '%s' as an integer", b.Signature)
	}
	if !isValidSignature(a.PublicKey, b.HashBytes, bigIntToBytes(sig, 64)) {
		err = errInvalidSignature
	}
	return
}

// FetchWork uses the generate_work RPC on node to fetch and then set
// the Work of b.
func (b *Block) FetchWork(workThreshold uint64, publicKey *big.Int, node string) error {
	var hash string
	if b.Previous == "0000000000000000000000000000000000000000000000000000000000000000" {
		hash = fmt.Sprintf("%064X", bigIntToBytes(publicKey, 32))
	} else {
		hash = b.Previous
	}
	requestBody := fmt.Sprintf(`{`+
		`"action": "work_generate",`+
		`"hash": "%s",`+
		`"difficulty": "%016x"`+
		`}`, string(hash), workThreshold)
	responseBytes, err := doRPC(requestBody, node)
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

func (b *Block) hash(publicKey *big.Int) error {
	// See https://nanoo.tools/block for a reference.

	msg := make([]byte, 176, 176)

	msg[31] = 0x6 // block preamble

	copy(msg[32:64], bigIntToBytes(publicKey, 32))

	previous, err := hex.DecodeString(b.Previous)
	if err != nil {
		return err
	}
	copy(msg[64:96], previous)

	representative, err := getPublicKeyFromAddress(b.Representative)
	if err != nil {
		return err
	}
	copy(msg[96:128], bigIntToBytes(representative, 32))

	balance, ok := big.NewInt(0).SetString(b.Balance, 10)
	if !ok {
		return fmt.Errorf("cannot parse '%s' as an integer", b.Balance)
	}
	copy(msg[128:144], bigIntToBytes(balance, 16))

	link, err := hex.DecodeString(b.Link)
	if err != nil {
		return err
	}
	copy(msg[144:176], link)

	hash := blake2b.Sum256(msg)
	b.HashBytes = hash[:]
	b.Hash = fmt.Sprintf("%064X", b.HashBytes)
	return nil
}

// Submit submits the Block to the given node. Work and Signature of b
// must be populated beforehand.
func (b Block) Submit(node string) error {
	if b.Work == "" {
		return ErrWorkMissing
	}
	if b.Signature == "" {
		return ErrSignatureMissing
	}
	process := process{
		Action:    "process",
		JsonBlock: "true",
		SubType:   b.SubType,
		Block:     b,
	}
	return doProcessRPC(process, node)
}
