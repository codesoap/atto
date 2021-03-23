package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
)

type accountInfo struct {
	Error                      string `json:"error"`
	Frontier                   string `json:"frontier"`
	RepresentativeBlock        string `json:"representative_block"`
	Balance                    string `json:"balance"`
}

type pending struct {
	Error  string        `json:"error"`
	Blocks pendingBlocks `json:"blocks"`
}

type pendingBlocks map[string]pendingBlockSource

// UnmarshalJSON just unmarshals a list of strings, but
// interprets an empty string as an empty list. This is
// neccessary due to a bug in the Nano node implementation. See
// https://github.com/nanocurrency/nano-node/issues/3161.
func (b *pendingBlocks) UnmarshalJSON(in []byte) error {
	if string(in) == `""` {
		return nil
	}
	var raw map[string]pendingBlockSource
	err := json.Unmarshal(in, &raw)
	*b = pendingBlocks(raw)
	return err
}

type pendingBlockSource struct {
	Amount string `json:"amount"`
	Source string `json:"source"`
}

func printBalance() error {
	seed, err := getSeed()
	if err != nil {
		return err
	}
	privateKey := getPrivateKey(seed, uint32(accountIndexFlag))
	address, err := getAddress(privateKey)
	if err != nil {
		return err
	}

	info, err := getAccountInfo(address)
	if err != nil {
		return err
	}
	receivedAmount, err := receivePendingSends(info.Frontier, privateKey)
	if err != nil {
		return err
	}
	balance, ok := big.NewInt(0).SetString(info.Balance, 10)
	if !ok {
		return fmt.Errorf("cannot parse '%s' as an integer", info.Balance)
	}
	balance = big.NewInt(0).Add(balance, receivedAmount)
	fmt.Println(rawToNanoString(balance))
	return nil
}

func receivePendingSends(frontier string, privateKey *big.Int) (receivedAmount *big.Int, err error) {
	receivedAmount = big.NewInt(0)
	address, err := getAddress(privateKey)
	if err != nil {
		return
	}
	sends, err := getPendingSends(address)
	if err != nil {
		return
	}
	for blockHash, source := range sends {
		amount, ok := big.NewInt(0).SetString(source.Amount, 10)
		if !ok {
			err = fmt.Errorf("cannot parse '%s' as an integer", source.Amount)
			return
		}
		receivedAmount = big.NewInt(0).Add(receivedAmount, amount)
		txt := "Initiating receival of %s from %s... "
		fmt.Fprintf(os.Stderr, txt, rawToNanoString(amount), source.Source)
		err = receiveSend(blockHash, source, privateKey)
		if err != nil {
			return
		}
		fmt.Fprintln(os.Stderr, "done")
	}
	return
}

func getPendingSends(address string) (sends pendingBlocks, err error) {
	requestBody := fmt.Sprintf(`{`+
		`"action": "pending", `+
		`"account": "%s", `+
		`"include_only_confirmed": "true", `+
		`"source": "true"`+
		`}`, address)
	responseBytes, err := doRPC(requestBody)
	if err != nil {
		return
	}
	var pending pending
	err = json.Unmarshal(responseBytes, &pending)
	// Need to check pending.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if err == nil && pending.Error != "" {
		err = fmt.Errorf("could not fetch unreceived sends: %s", pending.Error)
	}
	return pending.Blocks, err
}

func receiveSend(blockHash string, source pendingBlockSource, privateKey *big.Int) error {
	return nil
}

func getAccountInfo(address string) (info accountInfo, err error) {
	requestBody := fmt.Sprintf(`{"action": "account_info", "account": "%s"}`, address)
	responseBytes, err := doRPC(requestBody)
	if err != nil {
		return
	}
	err = json.Unmarshal(responseBytes, &info)
	// Need to check info.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if err == nil && info.Error != "" {
		err = fmt.Errorf("could not fetch balance: %s", info.Error)
	}
	return
}
