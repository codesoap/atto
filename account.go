package main

import (
	"encoding/json"
	"fmt"
)

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

func getAccountInfo(address string) (info accountInfo, err error) {
	requestBody := fmt.Sprintf(`{`+
		`"action": "account_info",`+
		`"account": "%s",`+
		`"representative": "true"`+
		`}`, address)
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
		info.Frontier = "0000000000000000000000000000000000000000000000000000000000000000"
		info.Representative = defaultRepresentative
		info.Balance = "0"
	} else if info.Error != "" {
		err = fmt.Errorf("could not fetch account info: %s", info.Error)
		return
	} else {
		err = verifyInfo(info, address)
	}
	return
}

// verifyInfo gets the frontier block of info, ensures that Hash,
// Representative and Balance match and verifies it's signature.
func verifyInfo(info accountInfo, address string) error {
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
	publicKey, err := getPublicKeyFromAddress(address)
	if err != nil {
		return err
	}
	if err = block.Contents.verifySignature(publicKey); err == errInvalidSignature ||
		info.Frontier != block.Contents.Hash ||
		info.Representative != block.Contents.Representative ||
		info.Balance != block.Contents.Balance {
		return fmt.Errorf("the received account info has been manipulated; " +
			"change your node immediately!")
	}
	return err
}
