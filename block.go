package atto

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/blake2b"
)

var errInvalidSignature = fmt.Errorf("invalid block signature")

// ErrSignatureMissing is used when the Signature of a Block is missing
// but required for the attempted operation.
var ErrSignatureMissing = fmt.Errorf("signature is missing")

// ErrWorkMissing is used when the Work of a Block is missing but
// required for the attempted operation.
var ErrWorkMissing = fmt.Errorf("work is missing")

var (
	// See https://docs.nano.org/integration-guides/work-generation/#difficulty-thresholds
	defaultWorkThreshold uint64 = 0xfffffff800000000
	receiveWorkThreshold uint64 = 0xfffffe0000000000
)

// BlockSubType represents the sub-type of a block.
type BlockSubType int64

const (
	// SubTypeReceive denotes blocks which raise the balance.
	SubTypeReceive BlockSubType = iota

	// SubTypeChange denotes blocks which change the representative.
	SubTypeChange

	// SubTypeSend denotes blocks which lower the balance.
	SubTypeSend
)

// Block represents a block in the block chain of an account.
type Block struct {
	Type           string `json:"type"`
	Account        string `json:"account"`
	Previous       string `json:"previous"`
	Representative string `json:"representative"`
	Balance        string `json:"balance"`
	Link           string `json:"link"`
	Signature      string `json:"signature"`
	Work           string `json:"work"`

	// This field is not part of the JSON but needed to improve the
	// performance of FetchWork and the security of Submit.
	SubType BlockSubType `json:"-"`
}

type workGenerateResponse struct {
	Error string `json:"error"`
	Work  string `json:"work"`
}

// Sign computes and sets the Signature of b.
func (b *Block) Sign(privateKey *big.Int) error {
	publicKey, err := getPublicKeyFromAddress(b.Account)
	if err != nil {
		return err
	}
	hash, err := b.hashBytes()
	if err != nil {
		return err
	}
	signature, err := sign(publicKey, privateKey, hash)
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
	hash, err := b.hashBytes()
	if err != nil {
		return err
	}
	if !isValidSignature(a.PublicKey, hash, bigIntToBytes(sig, 64)) {
		err = errInvalidSignature
	}
	return
}

// FetchWork uses the generate_work RPC on node to fetch and then set
// the Work of b.
func (b *Block) FetchWork(node string) error {
	hash, err := b.workHash()
	if err != nil {
		return err
	}

	requestBody := fmt.Sprintf(`{"action":"work_generate", "hash":"%s"`, hash)
	if b.SubType == SubTypeReceive {
		// Receive blocks need less work, so lower the difficulty.
		requestBody += fmt.Sprintf(`, "difficulty":"%016x"`, receiveWorkThreshold)
	}
	requestBody += `}`

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

// GenerateWork uses the CPU of the local computer to generate work and
// then sets it as b.Work.
func (b *Block) GenerateWork() error {
	hashString, err := b.workHash()
	if err != nil {
		return err
	}
	hash, err := hex.DecodeString(hashString)
	if err != nil {
		return err
	}
	workThreshold := defaultWorkThreshold
	if b.SubType == SubTypeReceive {
		// Receive blocks need less work, so lower the difficulty.
		workThreshold = receiveWorkThreshold
	}
	nonce, err := findNonce(workThreshold, hash)
	if err != nil {
		return err
	}
	b.Work = fmt.Sprintf("%016x", nonce)
	return nil
}

func (b Block) workHash() (string, error) {
	if b.Previous == strings.Repeat("0", 64) {
		publicKey, err := getPublicKeyFromAddress(b.Account)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%064X", bigIntToBytes(publicKey, 32)), nil
	}
	return b.Previous, nil
}

// Hash calculates the block's hash and returns it's string
// representation.
func (b Block) Hash() (string, error) {
	hashBytes, err := b.hashBytes()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%064X", hashBytes), nil
}

func (b Block) hashBytes() ([]byte, error) {
	// See https://nanoo.tools/block for a reference.

	msg := make([]byte, 176, 176)

	msg[31] = 0x6 // block preamble

	publicKey, err := getPublicKeyFromAddress(b.Account)
	if err != nil {
		return nil, err
	}
	copy(msg[32:64], bigIntToBytes(publicKey, 32))

	previous, err := hex.DecodeString(b.Previous)
	if err != nil {
		return nil, err
	}
	copy(msg[64:96], previous)

	representative, err := getPublicKeyFromAddress(b.Representative)
	if err != nil {
		return nil, err
	}
	copy(msg[96:128], bigIntToBytes(representative, 32))

	balance, ok := big.NewInt(0).SetString(b.Balance, 10)
	if !ok {
		return nil, fmt.Errorf("cannot parse '%s' as an integer", b.Balance)
	}
	copy(msg[128:144], bigIntToBytes(balance, 16))

	link, err := hex.DecodeString(b.Link)
	if err != nil {
		return nil, err
	}
	copy(msg[144:176], link)

	hash := blake2b.Sum256(msg)
	return hash[:], nil
}

// Submit submits the Block to the given node. Work and Signature of b
// must be populated beforehand.
//
// May return ErrWorkMissing or ErrSignatureMissing.
func (b Block) Submit(node string) error {
	if b.Work == "" {
		return ErrWorkMissing
	}
	if b.Signature == "" {
		return ErrSignatureMissing
	}
	var subType string
	switch b.SubType {
	case SubTypeReceive:
		subType = "receive"
	case SubTypeChange:
		subType = "change"
	case SubTypeSend:
		subType = "send"
	}
	process := process{
		Action:    "process",
		JsonBlock: "true",
		SubType:   subType,
		Block:     b,
	}
	return doProcessRPC(process, node)
}
