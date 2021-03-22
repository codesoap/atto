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
	Error  string `json:"error"`
	Blocks blocks `json:"blocks"`
}

type blocks []string

// UnmarshalJSON just unmarshals a list of strings, but
// interprets an empty string as an empty list. This is
// neccessary due to a bug in the Nano node implementation. See
// https://github.com/nanocurrency/nano-node/issues/3161.
func (b *blocks) UnmarshalJSON(in []byte) error {
	if string(in) == `""` {
		return nil
	}
	var raw []string
	err := json.Unmarshal(in, &raw)
	*b = blocks(raw)
	return err
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
	balanceRaw, ok := big.NewInt(0).SetString(info.Balance, 10)
	if !ok {
		return fmt.Errorf("invalid balance '%s' received", info.Balance)
	}
	rawPerKnano, _ := big.NewInt(0).SetString("1000000000000000000000000000", 10)
	balance := balanceRaw.Div(balanceRaw, rawPerKnano).Uint64()
	fmt.Printf("%d.%03d NANO\n", balance/1000, balance%1000)
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
	for _, send := range sends {
		fmt.Fprintf(os.Stderr, "Receiving funds from %s... ", send)
		err = receiveSend(send, privateKey)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "done")
	}
	return nil
}

func getPendingSends(address string) (sends []string, err error) {
	requestBody := fmt.Sprintf(`{"action": "pending", "account": "%s"}`, address)
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

func receiveSend(blockId string, privateKey *big.Int) error {
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
