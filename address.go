package main

import (
	"fmt"
	"math/big"
	"strings"

	"filippo.io/edwards25519"
	"golang.org/x/crypto/blake2b"
)

func printAddress() error {
	seed, err := getSeed()
	if err != nil {
		return err
	}
	privateKey := getPrivateKey(seed, uint32(accountIndexFlag))
	address, err := getAddress(privateKey)
	if err != nil {
		return err
	}
	fmt.Println(address)
	return nil
}

func getAddress(privateKey *big.Int) (string, error) {
	publicKey := derivePublicKey(privateKey)
	base32PublicKey := base32Encode(publicKey)

	publicKeyBytes := make([]byte, 32, 32)
	publicKey.FillBytes(publicKeyBytes)
	hasher, err := blake2b.New(5, nil)
	if err != nil {
		return "", err
	}
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

func getPrivateKey(seed *big.Int, index uint32) *big.Int {
	seedBytes := make([]byte, 32, 32)
	seed.FillBytes(seedBytes)
	indexBytes := make([]byte, 4, 4)
	big.NewInt(int64(index)).FillBytes(indexBytes)
	in := append(seedBytes, indexBytes...)
	privateKeyBytes := blake2b.Sum256(in)
	return big.NewInt(0).SetBytes(privateKeyBytes[:])
}

func derivePublicKey(privateKey *big.Int) *big.Int {
	privateKeyBytes := make([]byte, 32, 32)
	privateKey.FillBytes(privateKeyBytes)
	hashBytes := blake2b.Sum512(privateKeyBytes)
	scalar := edwards25519.NewScalar().SetBytesWithClamping(hashBytes[:32])
	publicKeyBytes := edwards25519.NewIdentityPoint().ScalarBaseMult(scalar).Bytes()
	return big.NewInt(0).SetBytes(publicKeyBytes)
}

func getPublicKeyFromAddress(address string) (*big.Int, error) {
	if len(address) == 64 {
		return base32Decode(address[4:56])
	} else if len(address) == 65 {
		return base32Decode(address[5:57])
	}
	return nil, fmt.Errorf("could not parse address %s", address)
}
