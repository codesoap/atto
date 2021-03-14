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
		err = fmt.Errorf("could not fetch balance: %s", info.Error)
	}
	return
}
