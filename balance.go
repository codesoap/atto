package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
)

// TODO: Remove unused fields.
type accountInfo struct {
	Error                      string `json:"error"`
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

	if !noReceiveFlag {
		if err = receivePendingSends(privateKey); err != nil {
			return err
		}
	}

	info, err := getAccountInfo(address)
	if err != nil {
		return err
	}
	balance, err := rawToNano(info.Balance)
	if err != nil {
		return err
	}
	fmt.Println(balance)
	return nil
}

func receivePendingSends(privateKey *big.Int) error {
	address, err := getAddress(privateKey)
	if err != nil {
		return err
	}
	sends, err := getPendingSends(address)
	if err != nil {
		return err
	}
	for blockHash, source := range sends {
		amount, err := rawToNano(source.Amount)
		if err != nil {
			return err
		}
		txt := "Initiating receival of %s from %s... "
		fmt.Fprintf(os.Stderr, txt, amount, source.Source)
		err = receiveSend(blockHash, source, privateKey)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "done")
	}
	return nil
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
