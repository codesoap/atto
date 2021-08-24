package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"filippo.io/edwards25519"
	"golang.org/x/crypto/blake2b"
)

var errAccountNotFound = fmt.Errorf("account has not yet been opened")

type account struct {
	privateKey *big.Int
	publicKey  *big.Int
	address    string
}

type accountInfo struct {
	Error          string `json:"error"`
	Frontier       string `json:"frontier"`
	Representative string `json:"representative"`
	Balance        string `json:"balance"`
}

type blockInfo struct {
	Error    string `json:"error"`
	Contents block  `json:"contents"`
}

// ownAccount initializes the own account using the seed provided via
// standard input and accountIndexFlag.
func ownAccount() (a account, err error) {
	seed, err := getSeed()
	if err != nil {
		return
	}
	a.privateKey = getPrivateKey(seed, uint32(accountIndexFlag))
	a.publicKey = derivePublicKey(a.privateKey)
	a.address, err = getAddress(a.publicKey)
	return
}

// getSeed takes the first line of the standard input and interprets it
// as a hexadecimal representation of a 32byte seed.
func getSeed() (*big.Int, error) {
	in := bufio.NewReader(os.Stdin)
	firstLine, err := in.ReadString('\n')
	if err != nil {
		return nil, err
	}
	seed, ok := big.NewInt(0).SetString(strings.TrimSpace(firstLine), 16)
	if !ok {
		return nil, fmt.Errorf("could not parse seed")
	}
	return seed, nil
}

func getPrivateKey(seed *big.Int, index uint32) *big.Int {
	seedBytes := bigIntToBytes(seed, 32)
	indexBytes := bigIntToBytes(big.NewInt(int64(index)), 4)
	in := append(seedBytes, indexBytes...)
	privateKeyBytes := blake2b.Sum256(in)
	return big.NewInt(0).SetBytes(privateKeyBytes[:])
}

func derivePublicKey(privateKey *big.Int) *big.Int {
	hashBytes := blake2b.Sum512(bigIntToBytes(privateKey, 32))
	scalar := edwards25519.NewScalar().SetBytesWithClamping(hashBytes[:32])
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

	address := "nano_" +
		strings.Repeat("1", 52-len(base32PublicKey)) + base32PublicKey +
		strings.Repeat("1", 8-len(base32Hash)) + base32Hash
	return address, nil
}

func (a account) getInfo() (info accountInfo, err error) {
	requestBody := fmt.Sprintf(`{`+
		`"action": "account_info",`+
		`"account": "%s",`+
		`"representative": "true"`+
		`}`, a.address)
	responseBytes, err := doRPC(requestBody)
	if err != nil {
		return
	}
	if err = json.Unmarshal(responseBytes, &info); err != nil {
		return
	}
	// Need to check info.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if info.Error == "Account not found" {
		err = errAccountNotFound
	} else if info.Error != "" {
		err = fmt.Errorf("could not fetch account info: %s", info.Error)
	} else {
		err = a.verifyInfo(info)
	}
	return
}

// verifyInfo gets the frontier block of info, ensures that Hash,
// Representative and Balance match and verifies it's signature.
func (a account) verifyInfo(info accountInfo) error {
	requestBody := fmt.Sprintf(`{`+
		`"action": "block_info",`+
		`"json_block": "true",`+
		`"hash": "%s"`+
		`}`, info.Frontier)
	responseBytes, err := doRPC(requestBody)
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
	if err = block.Contents.verifySignature(a); err == errInvalidSignature ||
		info.Frontier != block.Contents.Hash ||
		info.Representative != block.Contents.Representative ||
		info.Balance != block.Contents.Balance {
		return fmt.Errorf("the received account info has been manipulated; " +
			"change your node immediately!")
	}
	return err
}
