package atto

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"filippo.io/edwards25519"
	"golang.org/x/crypto/blake2b"
)

// ErrAccountNotFound is used when an account could not be found by the
// queried node.
var ErrAccountNotFound = fmt.Errorf("account has not yet been opened")

// ErrAccountManipulated is used when it seems like an account has been
// manipulated. This probably means someone is trying to steal funds.
var ErrAccountManipulated = fmt.Errorf("the received account info has been manipulated")

// Account holds the public key and address of a Nano account.
type Account struct {
	PublicKey *big.Int
	Address   string
}

type blockInfo struct {
	Error    string `json:"error"`
	Contents Block  `json:"contents"`
}

// NewAccount creates a new Account and populates both its fields.
func NewAccount(privateKey *big.Int) (a Account, err error) {
	a.PublicKey = derivePublicKey(privateKey)
	a.Address, err = getAddress(a.PublicKey)
	return
}

// NewAccountFromAddress creates a new Account and populates both its
// fields.
func NewAccountFromAddress(address string) (a Account, err error) {
	a.Address = address
	a.PublicKey, err = getPublicKeyFromAddress(address)
	return
}

func derivePublicKey(privateKey *big.Int) *big.Int {
	hashBytes := blake2b.Sum512(bigIntToBytes(privateKey, 32))
	scalar, err := edwards25519.NewScalar().SetBytesWithClamping(hashBytes[:32])
	if err != nil {
		panic(err)
	}
	publicKeyBytes := edwards25519.NewIdentityPoint().ScalarBaseMult(scalar).Bytes()
	return big.NewInt(0).SetBytes(publicKeyBytes)
}

func getAddress(publicKey *big.Int) (string, error) {
	base32PublicKey := base32Encode(publicKey)

	hasher, err := blake2b.New(5, nil)
	if err != nil {
		return "", err
	}
	publicKeyBytes := bigIntToBytes(publicKey, 32)
	if _, err := hasher.Write(publicKeyBytes); err != nil {
		return "", err
	}
	hashBytes := hasher.Sum(nil)
	base32Hash := base32Encode(big.NewInt(0).SetBytes(revertBytes(hashBytes)))

	address := "ban_" +
		strings.Repeat("1", 52-len(base32PublicKey)) + base32PublicKey +
		strings.Repeat("1", 8-len(base32Hash)) + base32Hash
	return address, nil
}

// FetchAccountInfo fetches the AccountInfo of Account from the given
// node.
//
// It is also verified, that the retreived AccountInfo is valid by
// doing a block_info RPC for the frontier, verifying the signature
// and ensuring that no fields have been changed in the account_info
// response.
//
// May return ErrAccountNotFound or ErrAccountManipulated.
//
// If ErrAccountNotFound is returned, FirstReceive can be used to
// create a first Block and AccountInfo and create the account by then
// submitting this Block.
func (a Account) FetchAccountInfo(node string) (i AccountInfo, err error) {
	requestBody := fmt.Sprintf(`{`+
		`"action": "account_info",`+
		`"account": "%s",`+
		`"representative": "true"`+
		`}`, a.Address)
	responseBytes, err := doRPC(requestBody, node)
	if err != nil {
		return
	}
	if err = json.Unmarshal(responseBytes, &i); err != nil {
		return
	}
	// Need to check i.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if i.Error == "Account not found" {
		err = ErrAccountNotFound
	} else if i.Error != "" {
		err = fmt.Errorf("could not fetch account info: %s", i.Error)
	} else {
		i.PublicKey = a.PublicKey
		i.Address = a.Address
		err = a.verifyInfo(i, node)
	}
	return
}

// verifyInfo gets the frontier block of info, ensures that Hash,
// Representative and Balance match and verifies it's signature.
func (a Account) verifyInfo(info AccountInfo, node string) error {
	requestBody := fmt.Sprintf(`{`+
		`"action": "block_info",`+
		`"json_block": "true",`+
		`"hash": "%s"`+
		`}`, info.Frontier)
	responseBytes, err := doRPC(requestBody, node)
	if err != nil {
		return err
	}
	var block blockInfo
	if err = json.Unmarshal(responseBytes, &block); err != nil {
		return err
	}
	if info.Error != "" {
		return fmt.Errorf("could not get block info: %s", info.Error)
	}
	hash, err := block.Contents.Hash()
	if err != nil {
		return err
	}
	if err = block.Contents.verifySignature(a); err == errInvalidSignature ||
		info.Frontier != hash ||
		info.Representative != block.Contents.Representative ||
		info.Balance != block.Contents.Balance {
		return ErrAccountManipulated
	}
	return err
}

// FetchPending fetches all unreceived blocks of Account from node.
func (a Account) FetchPending(node string) ([]Pending, error) {
	requestBody := fmt.Sprintf(`{`+
		`"action": "pending", `+
		`"account": "%s", `+
		`"include_only_confirmed": "true", `+
		`"source": "true"`+
		`}`, a.Address)
	responseBytes, err := doRPC(requestBody, node)
	if err != nil {
		return nil, err
	}
	var pending internalPending
	err = json.Unmarshal(responseBytes, &pending)
	// Need to check pending.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if err == nil && pending.Error != "" {
		err = fmt.Errorf("could not fetch unreceived sends: %s", pending.Error)
	}
	return internalPendingToPending(pending), err
}

// FirstReceive creates the first receive block of an account. The block
// will still be missing its signature and work. FirstReceive will also
// return AccountInfo, which can be used to create further blocks.
func (a Account) FirstReceive(pending Pending, representative string) (AccountInfo, Block, error) {
	block := Block{
		Type:           "state",
		SubType:        SubTypeReceive,
		Account:        a.Address,
		Previous:       "0000000000000000000000000000000000000000000000000000000000000000",
		Representative: representative,
		Balance:        pending.Amount,
		Link:           pending.Hash,
	}
	hash, err := block.Hash()
	if err != nil {
		return AccountInfo{}, Block{}, err
	}
	info := AccountInfo{
		Frontier:       hash,
		Representative: block.Representative,
		Balance:        block.Balance,
		PublicKey:      a.PublicKey,
		Address:        a.Address,
	}
	return info, block, err
}
