package main

import (
	"fmt"
	"math/big"
)

func printAddress() error {
	account, err := ownAccount()
	if err == nil {
		fmt.Println(account.address)
	}
	return err
}

func getPublicKeyFromAddress(address string) (*big.Int, error) {
	if len(address) == 64 {
		return base32Decode(address[4:56])
	} else if len(address) == 65 {
		return base32Decode(address[5:57])
	}
	return nil, fmt.Errorf("could not parse address %s", address)
}
