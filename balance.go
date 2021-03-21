package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"math/big"
)

type accountInfo struct {
	Frontier                   string `json:"frontier"`
	OpenBlock                  string `json:"open_block"`
	RepresentativeBlock        string `json:"representative_block"`
	Balance                    string `json:"balance"`
	ModifiedTimestamp          string `json:"modified_timestamp"`
	BlockCount                 string `json:"block_count"`
	ConfirmationHeight         string `json:"confirmation_height"`
	ConfirmationHeightFrontier string `json:"confirmation_height_frontier"`
	AccountVersion             string `json:"account_version"`
}

func printBalance() error {
	seed, err := getSeed()
	if err != nil {
		return err
	}
	address, err := getAddress(seed, uint32(accountIndexFlag))
	if err != nil {
		return err
	}

	// TODO: Receive pending funds.

	info, err := getAccountInfo(address)
	if err != nil {
		return err
	}
	// FIXME: Don't use float!
	balanceRaw, ok := big.NewFloat(0).SetString(info.Balance)
	if !ok {
		return fmt.Errorf("invalid balance '%s' received", info.Balance)
	}
	rawPerNano, _ := big.NewFloat(0).SetString("1000000000000000000000000000000")
	balance, _ := balanceRaw.Quo(balanceRaw, rawPerNano).Float64()
	fmt.Printf("%.4f NANO\n", balance)
	return nil
}

func getAccountInfo(address string) (info accountInfo, err error) {
	url := getNodeUrl()
	requestBody := fmt.Sprintf(`{"action": "account_info", "account": "%s"}`, address)
	resp, err := http.Post(url, "application/json", strings.NewReader(requestBody))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(responseBytes, &info)
	return
}
