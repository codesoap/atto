package atto

import (
	"math/big"

	"filippo.io/edwards25519"
	"golang.org/x/crypto/blake2b"
)

func sign(publicKey, privateKey *big.Int, msg []byte) ([]byte, error) {
	// This implementation based on the one from github.com/iotaledger/iota.go.

	signature := make([]byte, 64, 64)

	h, err := blake2b.New512(nil)
	if err != nil {
		return signature, err
	}
	h.Write(bigIntToBytes(privateKey, 32))

	var digest1, messageDigest, hramDigest [64]byte
	h.Sum(digest1[:0])

	s, err := new(edwards25519.Scalar).SetBytesWithClamping(digest1[:32])
	if err != nil {
		return signature, err
	}

	h.Reset()
	h.Write(digest1[32:])
	h.Write(msg)
	h.Sum(messageDigest[:0])

	rReduced, err := new(edwards25519.Scalar).SetUniformBytes(messageDigest[:])
	if err != nil {
		return signature, err
	}
	R := new(edwards25519.Point).ScalarBaseMult(rReduced)

	encodedR := R.Bytes()

	h.Reset()
	h.Write(encodedR[:])
	h.Write(bigIntToBytes(publicKey, 32))
	h.Write(msg)
	h.Sum(hramDigest[:0])

	kReduced, err := new(edwards25519.Scalar).SetUniformBytes(hramDigest[:])
	if err != nil {
		return signature, err
	}
	S := new(edwards25519.Scalar).MultiplyAdd(kReduced, s, rReduced)

	copy(signature[:], encodedR[:])
	copy(signature[32:], S.Bytes())

	return signature, nil
}

func isValidSignature(publicKey *big.Int, msg, sig []byte) bool {
	// This implementation based on the one from github.com/iotaledger/iota.go.

	publicKeyBytes := bigIntToBytes(publicKey, 32)

	// ZIP215: this works because SetBytes does not check that encodings are canonical
	A, err := new(edwards25519.Point).SetBytes(publicKeyBytes)
	if err != nil {
		return false
	}
	A.Negate(A)

	h, err := blake2b.New512(nil)
	if err != nil {
		return false
	}
	h.Write(sig[:32])
	h.Write(publicKeyBytes)
	h.Write(msg)
	var digest [64]byte
	h.Sum(digest[:0])
	hReduced, err := new(edwards25519.Scalar).SetUniformBytes(digest[:])
	if err != nil {
		return false
	}

	// ZIP215: this works because SetBytes does not check that encodings are canonical
	checkR, err := new(edwards25519.Point).SetBytes(sig[:32])
	if err != nil {
		return false
	}

	// https://tools.ietf.org/html/rfc8032#section-5.1.7 requires that s be in
	// the range [0, order) in order to prevent signature malleability
	s, err := new(edwards25519.Scalar).SetCanonicalBytes(sig[32:])
	if err != nil {
		return false
	}

	R := new(edwards25519.Point).VarTimeDoubleScalarBaseMult(hReduced, A, s)

	// ZIP215: We want to check [8](R - checkR) == 0
	p := new(edwards25519.Point).Subtract(R, checkR)     // p = R - checkR
	p.Add(p, p)                                          // p = [2]p
	p.Add(p, p)                                          // p = [4]p
	p.Add(p, p)                                          // p = [8]p
	return p.Equal(edwards25519.NewIdentityPoint()) == 1 // p == 0
}
