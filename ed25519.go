package main

import (
	"math/big"

	"filippo.io/edwards25519"
	"golang.org/x/crypto/blake2b"
)

func sign(privateKey *big.Int, msg []byte) ([]byte, error) {
	// This implementation based on the one from github.com/iotaledger/iota.go.

	publicKey := derivePublicKey(privateKey)
	signature := make([]byte, 64, 64)

	h, err := blake2b.New512(nil)
	if err != nil {
		return signature, err
	}
	h.Write(bigIntToBytes(privateKey, 32))

	var digest1, messageDigest, hramDigest [64]byte
	h.Sum(digest1[:0])

	s := new(edwards25519.Scalar).SetBytesWithClamping(digest1[:32])

	h.Reset()
	h.Write(digest1[32:])
	h.Write(msg)
	h.Sum(messageDigest[:0])

	rReduced := new(edwards25519.Scalar).SetUniformBytes(messageDigest[:])
	R := new(edwards25519.Point).ScalarBaseMult(rReduced)

	encodedR := R.Bytes()

	h.Reset()
	h.Write(encodedR[:])
	h.Write(bigIntToBytes(publicKey, 32))
	h.Write(msg)
	h.Sum(hramDigest[:0])

	kReduced := new(edwards25519.Scalar).SetUniformBytes(hramDigest[:])
	S := new(edwards25519.Scalar).MultiplyAdd(kReduced, s, rReduced)

	copy(signature[:], encodedR[:])
	copy(signature[32:], S.Bytes())

	return signature, nil
}
